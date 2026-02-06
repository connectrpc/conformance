// Copyright 2023-2024 The Connect Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package referenceclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/conformance/internal"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/grpcutil"
	"connectrpc.com/conformance/internal/tracer"
	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/exp/constraints"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/anypb"
)

type wireCaptureTransport struct {
	Transport http.RoundTripper
}

// newWireCaptureTransport returns a new round-tripper that delegates to the
// given one. The returned transport instruments the given one for tracing and
// the ability to capture wire-level details from that trace. Also see
// withWireCapture and examineWireDetails.
//
// The contextcheck lint issue is a false positive here. And ignoring it with
// nolint then causes a false positive for the nolintlint :(
//
//nolint:contextcheck,nolintlint
func newWireCaptureTransport(transport http.RoundTripper, trace *tracer.Tracer) http.RoundTripper {
	return &wireCaptureTransport{
		Transport: tracer.TracingRoundTripper(transport, &wireTracer{
			tracer: trace,
		}),
	}
}

// RoundTrip replaces the response body with a wireReader which captures bytes
// as they are read.
func (w *wireCaptureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := w.Transport.RoundTrip(req)
	wrapper, ok := req.Context().Value(wireCtxKey{}).(*wireWrapper)
	if err != nil || !ok {
		return resp, err
	}
	// If this is a unary error with JSON body, replace the body with a reader
	// that will save off the body bytes as they are read so that we can access
	// the body contents in the tracer
	if isUnaryJSONError(resp.Header.Get("Content-Type"), resp.StatusCode) {
		resp.Body = &wireReader{body: resp.Body, wrapper: wrapper}
	}
	return resp, nil
}

type wireReader struct {
	body    io.ReadCloser
	wrapper *wireWrapper
}

func (w *wireReader) Read(p []byte) (int, error) {
	n, err := w.body.Read(p)

	// Capture bytes as they are read
	w.wrapper.buf.Write(p[:n])

	return n, err
}

func (w *wireReader) Close() error {
	return w.body.Close()
}

type wireWrapper struct {
	traceAvailable chan struct{}
	trace          tracer.Trace
	// buf represents the read response body
	buf *bytes.Buffer
}

type connectError struct {
	Code    *string           `json:"code"`
	Message *string           `json:"message"`
	Details []json.RawMessage `json:"details"`
}

type connectErrorDetail struct {
	Type  *string         `json:"type"`
	Value *string         `json:"value"`
	Debug json.RawMessage `json:"debug"`
}

type connectEndStream struct {
	Error    json.RawMessage     `json:"error"`
	Metadata map[string][]string `json:"metadata"`
}

// The key associated with the wire details information stored in context.
type wireCtxKey struct{}

// withWireCapture returns a new context which will contain wire details during
// a roundtrip. Use examineWireDetails to extract the details after the operation
// completes.
func withWireCapture(ctx context.Context) context.Context {
	return context.WithValue(ctx, wireCtxKey{}, &wireWrapper{
		traceAvailable: make(chan struct{}),
		buf:            &bytes.Buffer{},
	})
}

// setWireTrace sets the given trace in the given context. Should never be called
// more than once for the same context.
func setWireTrace(ctx context.Context, trace tracer.Trace) {
	wrapper, ok := ctx.Value(wireCtxKey{}).(*wireWrapper)
	if !ok {
		return
	}
	wrapper.trace = trace
	close(wrapper.traceAvailable)
}

// examineWireDetails examines certain wire details of the call and returns the
// HTTP status code (or if there is not one). Feedback about the wire details will
// be printed to the given printer. This also records errors to the given printer
// if the given context was never configured using withWireCapture or if the wire
// details are not available within 1 second of the call.
func examineWireDetails(ctx context.Context, printer internal.Printer) (statusCode int, ok bool) {
	wrapper, ok := ctx.Value(wireCtxKey{}).(*wireWrapper)
	if !ok {
		printer.Printf("unable to examine wire details: call context not configured (no wire wrapper found).")
		return 0, false
	}
	// Usually, the trace should be already available because examineWireDetails should not be
	// called until the underlying HTTP operation is complete. However, if the operation times
	// out (via context deadline exceeded), it is possible for the calling code to think the
	// operation is complete and call this function, concurrently with the HTTP round-tripper
	// completing the call and storing the trace. So to avoid a race and avoid flaky instances
	// where the trace isn't available yet, we allow a one-second grace period.
	select {
	case <-wrapper.traceAvailable:
	case <-time.After(time.Second):
		printer.Printf("unable to examine wire details: completed trace not found in call context.")
		return 0, false
	}
	trace := &wrapper.trace
	if trace.Response == nil {
		// A nil response in the trace is valid if the HTTP round trip failed.
		return 0, false
	}

	// Check end-stream and/or error JSON data in the response.
	contentType := trace.Response.Header.Get("Content-Type")
	switch {
	case isUnaryJSONError(contentType, trace.Response.StatusCode):
		// If this is a unary request that returned an error, then use the entire
		// response body as the wire error details.
		decomp := tracer.GetDecompressor(trace.Response.Header.Get("Content-Encoding"))
		if err := decomp.Reset(wrapper.buf); err == nil {
			if body, err := io.ReadAll(decomp); err == nil {
				examineConnectError(body, printer)
			}
		}
	case strings.HasPrefix(contentType, "application/connect+"):
		// If this is a streaming Connect request, then look through the trace events
		// for the ResponseBodyEndStream event and parse its content into an
		// endStreamError to see if there are any error details.
		endStreamContent, ok := getBodyEndStream(trace)
		if ok {
			examineConnectEndStream([]byte(endStreamContent), printer)
		}
	case strings.HasPrefix(contentType, "application/grpc-web"):
		// For gRPC-Web, capture the trailers in the body. We don't do any case normalization
		// or trimming of excess whitespace so that the full values are available to check.
		endStreamContent, ok := getBodyEndStream(trace)
		if ok {
			headers := examineGRPCEndStream(endStreamContent, printer)
			checkGRPCStatus(headers, printer)
		} else if isTrailersOnlyResponse(trace) {
			checkGRPCStatus(trace.Response.Header, printer)
		}
	case strings.HasPrefix(contentType, "application/grpc"):
		if isTrailersOnlyResponse(trace) {
			checkGRPCStatus(trace.Response.Header, printer)
		} else if len(trace.Response.Trailer) > 0 {
			checkGRPCStatus(trace.Response.Trailer, printer)
		}
	}

	if contentType != "application/grpc" && !strings.HasPrefix(contentType, "application/grpc+") {
		// It's not gRPC protocol, so there should be no HTTP trailers.
		if len(trace.Response.Trailer) > 0 {
			printer.Printf("response included %d HTTP trailers but should not have any", len(trace.Response.Trailer))
		}
	}

	return trace.Response.StatusCode, true
}

// examineStreamDelivery checks whether streaming response messages were
// delivered incrementally rather than buffered into a single batch. It
// examines the timing of ResponseBodyData events in the wire trace and
// reports feedback if all messages arrived too close together relative to
// the configured response delay.
func examineStreamDelivery(ctx context.Context, numExpectedMessages int, delayMs uint32, printer internal.Printer) {
	if delayMs == 0 || numExpectedMessages < 2 {
		return
	}
	wrapper, ok := ctx.Value(wireCtxKey{}).(*wireWrapper)
	if !ok {
		return
	}
	select {
	case <-wrapper.traceAvailable:
	case <-time.After(time.Second):
		return
	}
	trace := &wrapper.trace

	// Collect the offset of each response body data event.
	var offsets []time.Duration
	for _, event := range trace.Events {
		if rbd, ok := event.(*tracer.ResponseBodyData); ok {
			offsets = append(offsets, rbd.Offset)
		}
	}
	if len(offsets) < 2 {
		return
	}

	actualSpan := offsets[len(offsets)-1] - offsets[0]
	// Use 25% of the expected span as a generous minimum to avoid CI flakiness.
	expectedSpan := time.Duration(len(offsets)-1) * time.Duration(delayMs) * time.Millisecond
	minSpan := expectedSpan / 4

	if actualSpan < minSpan {
		printer.Printf(
			"response messages were not delivered incrementally: "+
				"%d messages arrived within %v of each other, "+
				"but with a %dms response delay between %d messages expected a span of at least %v",
			len(offsets), actualSpan.Round(time.Millisecond),
			delayMs, len(offsets), minSpan.Round(time.Millisecond),
		)
	}
}

type wireTracer struct {
	tracer *tracer.Tracer
}

// Complete intercepts the Complete call for a tracer, extracting wire details
// from the passed trace. The wire details will be stored in the context acquired by
// withWireCapture and can be retrieved via examineWireDetails.
func (t *wireTracer) Complete(trace tracer.Trace) {
	setWireTrace(trace.Request.Context(), trace)
	if t.tracer != nil {
		t.tracer.Complete(trace)
	}
}

// isUnaryJSONError returns whether the given content type and HTTP status code
// represents a unary JSON error.
func isUnaryJSONError(contentType string, statusCode int) bool {
	return contentType == "application/json" && statusCode != http.StatusOK
}

// getBodyEndStream returns the contents of any end-stream message in the trace.
// The bool value will be true if an end-stream message is found and otherwise
// it will be false.
func getBodyEndStream(trace *tracer.Trace) (string, bool) {
	for _, event := range trace.Events {
		endStream, ok := event.(*tracer.ResponseBodyEndStream)
		if !ok {
			continue
		}
		return endStream.Content, true
	}
	return "", false
}

func examineConnectError(errJSON json.RawMessage, printer internal.Printer) {
	var connErr *connectError
	var hasCode, hasDetails bool
	okay := examineJSON(errJSON, &connErr, "connect error JSON", printer, func(key string, val any) {
		switch key {
		case "code":
			hasCode = true
			strVal, ok := val.(string)
			if !ok {
				printer.Printf(`connect error JSON: value for key "code" is a %T instead of a string`, val)
				break
			}
			var found bool
			// Valid RPC codes range from 1 to 16. (Zero means no an error; negative and >16 are illegal values.)
			for code := connect.Code(1); code <= 16; code++ {
				if strVal == code.String() {
					found = true
					break
				}
			}
			if !found {
				printer.Printf(`connect error JSON: value for key "code" is not a recognized error code name: %q`, val)
			}
		case "message":
			if _, isStr := val.(string); !isStr {
				printer.Printf(`connect error JSON: value for key "message" is a %T instead of a string`, val)
			}
		case "details":
			hasDetails = true
			if _, isSlice := val.([]any); !isSlice {
				printer.Printf(`connect error JSON: value for key "details" is a %T instead of a slice`, val)
			}
		default:
			printer.Printf("connect error JSON: invalid key %q", key)
		}
	})
	if !okay {
		return
	}
	// There is one required field.
	if !hasCode {
		printer.Printf(`connect error JSON: missing required key "code"`)
	}
	// Also check enclosed details.
	if hasDetails {
		for i, detail := range connErr.Details {
			examineConnectErrorDetail(i, detail, printer)
		}
	}
}

func examineConnectErrorDetail(i int, detailJSON json.RawMessage, printer internal.Printer) {
	var detail *connectErrorDetail
	prefix := fmt.Sprintf("connect error JSON: details[%d]", i)
	var decodedVal []byte
	var hasType, hasValue, hasDebug bool
	okay := examineJSON(detailJSON, &detail, prefix, printer, func(key string, val any) {
		switch key {
		case "type":
			hasType = true
			str, ok := val.(string)
			if !ok {
				printer.Printf(`%s: value for key "type" is a %T instead of a string`, prefix, val)
				break
			}
			if !protoreflect.FullName(str).IsValid() {
				printer.Printf(`%s: value for key "type", %q, is not a valid type name`, prefix, val)
				break
			}
		case "value":
			hasValue = true
			str, ok := val.(string)
			if !ok {
				printer.Printf(`%s: value for key "value" is a %T instead of a string`, prefix, val)
				break
			}
			decoded, err := base64.RawStdEncoding.DecodeString(str)
			if err != nil {
				printer.Printf(`%s: value for key "value", %q, is not valid unpadded base64-encoding: %v`, prefix, val, err)
				detail.Value = nil // this will skip the comparison of "value" and "debug" info below
				break
			}
			decodedVal = decoded
		case "debug":
			hasDebug = len(detail.Debug) > 0
			// Since a detail message could be a google.protobuf.Value, it
			// could technically be *any* kind of JSON value here. So no
			// further checks here. We'll check below that the debug field
			// (if present) actually agrees with the value field.
		default:
			printer.Printf("%s: invalid key %q", prefix, key)
		}
	})
	if !okay {
		return
	}
	// Two required fields:
	if !hasType {
		printer.Printf(`connect error JSON: details[%d]: missing required key "type"`, i)
	}
	if !hasValue {
		printer.Printf(`connect error JSON: details[%d]: missing required key "value"`, i)
	}
	if detail != nil && detail.Type != nil && detail.Value != nil && hasDebug {
		// Let's check the debug data by using it as JSON to unmarshal a message, and
		// then also unmarshal the bytes in value. If they are not equal messages,
		// there is an issues with the debug JSON data.
		examineConnectErrorDetailDebugData(i, *detail.Type, decodedVal, detail.Debug, printer)
	}
}

func examineConnectErrorDetailDebugData(i int, msgName string, data []byte, debugJSON []byte, printer internal.Printer) {
	msgType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(msgName))
	if err != nil {
		printer.Printf("connect error JSON: details[%d]: could not check debug data because message type %q could not be resolved: %v", i, msgName, err)
		return
	}
	msgFromValue := msgType.New()
	if err := proto.Unmarshal(data, msgFromValue.Interface()); err != nil {
		printer.Printf("connect error JSON: details[%d]: could not unmarshal message %q from value: %v", i, msgName, err)
		return
	}
	msgFromDebug := msgType.New()
	if err := protojson.Unmarshal(debugJSON, msgFromDebug.Interface()); err != nil {
		// It is possible that the debug data is actually a google.protobuf.Any, and the "@type" property
		// is causing the above unmarshal step to fail. So we'll try unmarshaling from Any next to see
		// if that is successful.
		//
		// NOTE: this fallback logic doesn't really work if the actual message type is a
		// google.protobuf.Value or google.protobuf.Struct. These types allow arbitrary JSON and will
		// accept the data in the above step, even if it has the extra "@type" property. Luckily, that
		// is not an issue since none of the conformance test cases try to use these types as error
		// details.
		//
		// TODO: This fallback is mainly to accommodate the current connect-go implementation, which
		//       includes the "@type" attribute because it just marshals the Any error detail message
		//       for the debug value. But we'd prefer that attribute NOT be in the debug value, in which
		//       case we could remove this fallback from here and let these checks be a bit more strict.
		var anyMsg anypb.Any
		if anyErr := protojson.Unmarshal(debugJSON, &anyMsg); anyErr != nil {
			// It's not a valid Any either. So report the original error
			printer.Printf("connect error JSON: details[%d]: could not unmarshal message %q from debug JSON: %v", i, msgName, err)
			return
		}
		typeNameFromAny := anyMsg.TypeUrl[strings.LastIndexByte(anyMsg.TypeUrl, '/')+1:]
		if typeNameFromAny != msgName {
			printer.Printf("connect error JSON: details[%d]: debug data indicates type %q but should indicate type %q", i, typeNameFromAny, msgName)
			return
		}
		msgFromAny, err := anyMsg.UnmarshalNew()
		if err != nil {
			printer.Printf("connect error JSON: details[%d]: could not unmarshal message %q from debug JSON: %v", i, msgName, err)
			return
		}
		msgFromDebug = msgFromAny.ProtoReflect()
	}
	diff := cmp.Diff(msgFromValue.Interface(), msgFromDebug.Interface(), protocmp.Transform())
	if diff != "" {
		printer.Printf("connect error JSON: details[%d]: debug data does not match value: - value, + debug\n%s", i, diff)
	}
}

func examineConnectEndStream(endStreamJSON json.RawMessage, printer internal.Printer) {
	var endStream *connectEndStream
	var hasError bool
	okay := examineJSON(endStreamJSON, &endStream, "connect end stream JSON", printer, func(key string, val any) {
		switch key {
		case "error":
			if _, isObj := val.(map[string]any); !isObj {
				printer.Printf(`connect end stream JSON: value for key "error" is a %T instead of a map/object`, val)
				break
			}
			hasError = true
		case "metadata":
			mapVal, ok := val.(map[string]any)
			if !ok {
				printer.Printf(`connect end stream JSON: value for key "metadata" is a %T instead of a map/object`, val)
			}
			for name, values := range mapVal {
				if !isValidHTTPFieldName(name) {
					printer.Printf(`connect end stream JSON: metadata[%q]: entry key is not a valid HTTP field name`, name)
				}
				valSlice, ok := values.([]any)
				if !ok {
					printer.Printf(`connect end stream JSON: metadata[%q]: value is a %T instead of an array of strings`, name, values)
					continue
				}
				for i, val := range valSlice {
					valStr, ok := val.(string)
					if !ok {
						printer.Printf(`connect end stream JSON: metadata[%q]: value #%d is a %T instead of a string`, name, i+1, val)
						continue
					}
					if !isValidHTTPFieldValue(valStr) {
						printer.Printf(`connect end stream JSON: metadata[%q]: value #%d is not a valid HTTP field value: %q`, name, i+1, val)
					}
				}
			}
		default:
			printer.Printf("connect end stream JSON: invalid key %q", key)
		}
	})
	if !okay {
		return
	}
	if hasError {
		examineConnectError(endStream.Error, printer)
	}
}

func examineGRPCEndStream(endStream string, printer internal.Printer) http.Header {
	// We break it using just LF, so we can handle alternate line ending. But we then complain
	// below if CRLF isn't used.
	endStreamLines := strings.Split(endStream, "\n")
	trailers := make(http.Header, len(endStreamLines))
	var linesWithoutCR int
	var blankLines int
	var endsInCRLF bool
	var blankLineAtEnd bool
	var obsLineFolds int
	var prevKey string
	for i, trailerLine := range endStreamLines {
		// Whole thing should end with CRLF, so last line should be blank
		switch {
		case i == len(endStreamLines)-1:
			if trailerLine == "" {
				endsInCRLF = true
				continue
			}
		case !strings.HasSuffix(trailerLine, "\r"):
			// Note: This is an "else" because we don't check the last line if
			// it's not blank since that means the whole block did not have a
			// terminating LF (which means the last line also doesn't need
			// trailing CR).
			linesWithoutCR++
		default:
			// Strip trailing CR.
			trailerLine = strings.TrimSuffix(trailerLine, "\r")
		}

		if trailerLine == "" {
			blankLines++
			if i == len(endStreamLines)-2 {
				blankLineAtEnd = true
			}
			continue
		}

		parts := strings.SplitN(trailerLine, ":", 2)
		key := parts[0]
		if i > blankLines && len(key) > 0 && (key[0] == ' ' || key[0] == '\t') {
			// Obsolete line-folding.
			// (See spec https://datatracker.ietf.org/doc/html/rfc7230#section-3.2.4)
			obsLineFolds++
			vals := trailers[prevKey]
			if len(vals) == 0 {
				// shouldn't be possible...
				trailers.Set(prevKey, strings.Trim(trailerLine, " \t"))
			} else {
				vals[len(vals)-1] += " " + strings.Trim(trailerLine, " \t")
			}
			continue
		}
		canonicalKey := textproto.CanonicalMIMEHeaderKey(key)
		if len(parts) != 2 {
			printer.Printf("grpc-web trailers include invalid field (missing colon): %q", trailerLine)
			trailers[canonicalKey] = append(trailers[canonicalKey], "")
			prevKey = canonicalKey
			continue
		}
		if !isValidHTTPFieldName(key) {
			printer.Printf("grpc-web trailers include invalid field; name contains invalid characters: %q", trailerLine)
		}
		// grpc-web protocol explicitly requires lower-case keys in end-stream message
		if key != strings.ToLower(key) {
			printer.Printf("grpc-web trailers include non-lower-case field key: %q", key)
		}
		// Leading and trailing whitespace is allowed, but only space and htab:
		// https://datatracker.ietf.org/doc/html/rfc7230#section-3.2
		val := strings.Trim(parts[1], " \t")
		if !isValidHTTPFieldValue(val) {
			printer.Printf("grpc-web trailers include invalid field; value contains invalid characters: %q", trailerLine)
		}
		trailers[canonicalKey] = append(trailers[canonicalKey], val)
		prevKey = canonicalKey
	}
	if obsLineFolds > 0 {
		printer.Printf("grpc-web trailers use obsolete line-folding")
	}
	if blankLines > 0 {
		if blankLines == 1 && blankLineAtEnd {
			printer.Printf("grpc-web trailers ends in extra blank line")
		} else {
			printer.Printf("grpc-web trailers include blank lines")
		}
	}
	if linesWithoutCR > 0 {
		printer.Printf("grpc-web trailers have lines with LF line ending instead of CRLF")
	}
	if !endsInCRLF {
		printer.Printf("grpc-web trailers should end with CRLF but does not")
	}
	return trailers
}

// isValidHTTPFieldValue returns true if the given string is a valid
// HTTP header or trailer value.
// https://datatracker.ietf.org/doc/html/rfc7230#section-3.2
func isValidHTTPFieldValue(s string) bool {
	// Not using "range s" because that uses UTF8 decoding to iterate
	// through runes. But spec is in terms of bytes.
	for i := range len(s) {
		char := s[i]
		// Visible range is 32 (SPACE ' ') and up, excluding DEL (127).
		// Horizontal tab (9, '\t') is allowed but outside the visible range.
		if char != '\t' && (char < 32 || char == 127) {
			// not valid
			return false
		}
	}
	return true
}

// isValidHTTPFieldName returns true if the given string is a valid
// HTTP header or trailer key.
// https://datatracker.ietf.org/doc/html/rfc7230#section-3.2.6 (see rule for token)
func isValidHTTPFieldName(s string) bool {
	// Not using "range s" because that uses UTF8 decoding to iterate
	// through runes. But spec is in terms of bytes.
	for i := range len(s) {
		char := s[i]
		switch char {
		case '!', '#', '$', '%', '&', '\'', '*', '+',
			'-', '.', '^', '_', '`', '|', '~': // allowed special chars
		default:
			switch {
			case char >= '0' && char <= '9': // digit
			case char >= 'a' && char <= 'z': // alpha
			case char >= 'A' && char <= 'Z':
			default:
				// Not one of the above? Disallowed.
				return false
			}
		}
	}
	return true
}

func examineJSON[T any](rawJSON []byte, dest **T, messagePrefix string, printer internal.Printer, forEachKey func(key string, val any)) bool {
	if err := json.Unmarshal(rawJSON, dest); err != nil {
		printer.Printf("%s: %v", messagePrefix, err)
		return false
	}
	// Extra checks, since encoding/json above is lenient.
	if *dest == nil {
		printer.Printf("%s: expecting an object but got <nil>", messagePrefix)
		return false
	}
	if _, err := checkNoDuplicateKeys("", json.NewDecoder(bytes.NewReader(rawJSON))); err != nil {
		printer.Printf("%s: %v", messagePrefix, err)
		return false
	}
	var asAny map[string]any
	if err := json.Unmarshal(rawJSON, &asAny); err != nil {
		// note: since above unmarshal step succeeded, this should never fail
		printer.Printf("%s: %v", messagePrefix, err)
		return false
	}
	for _, key := range sortedKeys(asAny) {
		val := asAny[key]
		forEachKey(key, val)
	}
	return true
}

func sortedKeys[K constraints.Ordered, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

// checkNoDuplicateKeys checks that the data read by the given decoder does
// not contain any JSON objects/maps that have duplicate keys. The "encoding/json"
// package does not check and will simply overwrite earlier values with later ones.
func checkNoDuplicateKeys(what string, dec *json.Decoder) (json.Token, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if tok != json.Delim('{') { //nolint:nestif
		if tok == json.Delim('[') {
			// consume entire array, checking each element
			var i int
			for {
				elemWhat := fmt.Sprintf("%s[%d]", what, i)
				tok, err := checkNoDuplicateKeys(elemWhat, dec)
				if err != nil {
					return nil, err
				}
				if tok == json.Delim(']') {
					break
				}
				i++
			}
		}
		// Not an object, so it's now fully consumed.
		return tok, nil
	}
	// The value must be an object. So look at all keys to make sure there are no duplicates.
	var prefix string
	if what != "" {
		prefix = what + ": "
	}
	keys := map[string]struct{}{}
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		if tok == json.Delim('}') {
			return tok, nil // done
		}
		key, ok := tok.(string)
		if !ok {
			// This should not be possible since the encoding/json package handles
			// validating this aspect of JSON structure.
			return nil, fmt.Errorf("%stoken for object key has type %T instead of string", prefix, tok)
		}
		if _, exists := keys[key]; exists {
			return nil, fmt.Errorf("%scontains duplicate key %q", prefix, key)
		}
		keys[key] = struct{}{}
		// recursively check the field value
		var elemWhat string
		if what == "" {
			elemWhat = key
		} else {
			elemWhat = fmt.Sprintf("%s.%s", what, key)
		}
		if _, err := checkNoDuplicateKeys(elemWhat, dec); err != nil {
			return nil, err
		}
	}
}

func isTrailersOnlyResponse(trace *tracer.Trace) bool {
	if trace.Response == nil {
		// no response received
		return false
	}
	if trace.Err != nil {
		// an error prevented us from seeing the end of the
		// response, so we can't tell if it's trailers-only or not
		return false
	}
	for _, trailer := range trace.Response.Trailer {
		if len(trailer) > 0 {
			// got actual trailers (in addition to headers)
			// so not trailers-only
			return false
		}
	}
	for _, event := range trace.Events {
		if _, ok := event.(*tracer.ResponseBodyData); ok {
			// got response body, so not trailers-only
			return false
		}
	}
	// We have a response that has no body and no trailers, so we
	// can indeed interpret it as a trailers-only response.
	return true
}

func checkGRPCStatus(headers http.Header, printer internal.Printer) { //nolint:gocyclo
	statusVals := headers.Values("Grpc-Status")
	var statusCode *int
	switch {
	case len(statusVals) > 1:
		printer.Printf("trailers include multiple 'grpc-status' keys (%d)", len(statusVals))
	case len(statusVals) == 0:
		printer.Printf("trailers did not include 'grpc-status' key")
	default:
		statusStr := statusVals[0]
		code, err := strconv.Atoi(statusStr)
		if err != nil {
			printer.Printf("trailers include invalid 'grpc-status' value %q: %v", statusStr, err)
		} else {
			statusCode = &code
			if code < 0 || code > 16 {
				printer.Printf("trailers include invalid 'grpc-status' value %d: should be >= 0 && <= 16", code)
			}
		}
	}

	msgVals := headers.Values("Grpc-Message")
	if len(msgVals) > 1 {
		printer.Printf("trailers include multiple 'grpc-message' keys (%d)", len(msgVals))
	}
	var msg *string
	if len(msgVals) > 0 { //nolint:nestif
		msgStr := msgVals[0]
		var expectHex int
		for i := range len(msgStr) {
			char := msgStr[i]
			if expectHex > 0 {
				if (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F') || (char >= '0' && char <= '9') {
					expectHex--
					continue
				}
				printer.Printf("trailers include incorrectly-encoded 'grpc-message' value %q: byte at position %d (0x%02x) should be hexadecimal digit", msgStr, i, char)
				expectHex = 0
				break
			}
			if char == '%' {
				expectHex = 2
				continue
			}
			if grpcutil.ShouldEscapeByteInMessage(char) {
				printer.Printf("trailers include incorrectly-encoded 'grpc-message' value %q: byte at position %d (0x%02x) should be percent-encoded", msgStr, i, char)
				break
			}
		}
		if expectHex > 0 {
			printer.Printf("trailers include incorrectly-encoded 'grpc-message' value %q: incomplete percent-encoded character at the end", msgStr)
		}
		if statusCode != nil && *statusCode == 0 && msgStr != "" {
			printer.Printf("trailers include a non-empty 'grpc-message' value with zero/okay 'grpc-status'")
		}
		if decoded, err := url.PathUnescape(msgStr); err == nil {
			msg = &decoded
		}
	}

	detailsBinVals := headers.Values("Grpc-Status-Details-Bin")
	if len(detailsBinVals) > 1 {
		printer.Printf("trailers include multiple 'grpc-status-details-bin' keys (%d)", len(detailsBinVals))
	}
	if len(detailsBinVals) == 0 {
		return // nothing else to check
	}
	detailsBin := detailsBinVals[0]
	data, err := base64.RawStdEncoding.DecodeString(detailsBin)
	if err != nil {
		data, err = base64.StdEncoding.DecodeString(detailsBin)
		if err != nil {
			printer.Printf("trailers include incorrectly-encoded 'grpc-status-details-bin' value: %v", err)
			return
		}
		printer.Printf("trailers include 'grpc-status-details-bin' value with padding but servers should emit unpadded: %s", detailsBin)
	}
	var statusProto status.Status
	if err := proto.Unmarshal(data, &statusProto); err != nil {
		printer.Printf("trailers include un-parseable 'grpc-status-details-bin' value: %v", err)
		return
	}
	if statusCode != nil && statusProto.Code != int32(*statusCode) {
		printer.Printf("trailers include 'grpc-status-details-bin' value that disagrees with 'grpc-status' value: %d != %d", statusProto.Code, *statusCode)
	}
	if statusProto.Code == 0 && len(statusProto.Details) > 0 {
		printer.Printf("trailers include 'grpc-status-details-bin' value with zero/okay 'grpc-status' and non-empty details")
	}
	if msg != nil && statusProto.Message != *msg {
		printer.Printf("trailers include 'grpc-status-details-bin' value that disagrees with 'grpc-message' value: %q != %q", statusProto.Message, *msg)
	}
}

func checkBinaryMetadata(name string, metadata []*conformancev1.Header, printer internal.Printer) {
	for _, entry := range metadata {
		lowerName := strings.ToLower(entry.Name)
		if !strings.HasSuffix(lowerName, "-bin") || lowerName == "grpc-status-details-bin" {
			// checked elsewhere when also verifying details match grpc-status and grpc-message trailers
			continue
		}
		for _, val := range entry.Value {
			_, err := base64.RawStdEncoding.DecodeString(val)
			if err != nil {
				_, err = base64.StdEncoding.DecodeString(val)
				if err != nil {
					printer.Printf("%s include incorrectly-encoded '%s' value: %v", name, entry.Name, err)
					return
				}
				printer.Printf("%s include '%s' value with padding but servers should emit unpadded: %s", name, entry.Name, val)
			}
		}
	}
}
