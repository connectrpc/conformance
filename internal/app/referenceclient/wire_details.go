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
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync/atomic"

	"connectrpc.com/conformance/internal"
	"connectrpc.com/conformance/internal/tracer"
)

// The key associated with the wire details information stored in context.
type wireCtxKey struct{}

type wireInterceptor struct {
	Transport http.RoundTripper
}

// newWireInterceptor creates a new wireInterceptor which wraps the given transport
// in a TracingRoundTripper.
func newWireInterceptor(transport http.RoundTripper, trace *tracer.Tracer) http.RoundTripper {
	return &wireInterceptor{
		Transport: tracer.TracingRoundTripper(transport, &wireTracer{
			tracer: trace,
		}),
	}
}

// RoundTrip replaces the response body with a wireReader which captures bytes
// as they are read.
func (w *wireInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := w.Transport.RoundTrip(req)
	wrapper, ok := req.Context().Value(wireCtxKey{}).(*wireWrapper)
	if err != nil || !ok {
		return resp, err
	}
	// If this is a unary error with JSON body, replace the body with a reader
	// that will save off the body bytes as they are read so that we can access
	// the body contents in the tracer
	if isUnaryJSONError(resp.Header.Get("content-type"), resp.StatusCode) {
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
	val atomic.Pointer[tracer.Trace]
	// buf represents the read response body
	buf *bytes.Buffer
}

// withWireCapture returns a new context which will contain wire details during
// a roundtrip.
func withWireCapture(ctx context.Context) context.Context {
	return context.WithValue(ctx, wireCtxKey{}, &wireWrapper{
		buf: &bytes.Buffer{},
	})
}

// setWireTrace sets the given trace in the given context.
func setWireTrace(ctx context.Context, trace *tracer.Trace) {
	wrapper, ok := ctx.Value(wireCtxKey{}).(*wireWrapper)
	if !ok {
		return
	}
	wrapper.val.Store(trace)
}

// examineWireDetails examines certain wire details of the call and returns the
// HTTP status code (or if there is not one). Feedback about the wire details will
// be printed to the given printer.
func examineWireDetails(ctx context.Context, printer internal.Printer) (statusCode int, ok bool) {
	wrapper, ok := ctx.Value(wireCtxKey{}).(*wireWrapper)
	if !ok {
		printer.Printf("unable to examine wire details: call context not configured (no wire wrapper found).")
		return 0, false
	}
	trace := wrapper.val.Load()
	if trace == nil {
		printer.Printf("unable to examine wire details: completed trace not found in call context.")
		return 0, false
	}
	if trace.Response == nil {
		// A nil response in the trace is valid if the HTTP round trip failed.
		return 0, false
	}

	// Check end-stream and/or error JSON data in the response.
	contentType := trace.Response.Header.Get("content-type")
	switch {
	case isUnaryJSONError(contentType, statusCode):
		// If this is a unary request that returned an error, then use the entire
		// response body as the wire error details.
		examineConnectError(wrapper.buf.Bytes(), printer)
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
			examineGRPCEndStream(endStreamContent, printer)
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

type wireTracer struct {
	tracer *tracer.Tracer
}

// Complete intercepts the Complete call for a tracer, extracting wire details
// from the passed trace. The wire details will be stored in the context acquired by
// withWireCapture and can be retrieved via examineWireDetails.
func (t *wireTracer) Complete(trace tracer.Trace) {
	setWireTrace(trace.Request.Context(), &trace)

	if t.tracer != nil {
		t.tracer.Complete(trace)
	}
}

// isUnaryJSONError returns whether the given content type and HTTP status code
// represents a unary JSON error.
func isUnaryJSONError(contentType string, statusCode int) bool {
	return contentType == "application/json" && statusCode != 200
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

func examineConnectError(_ json.RawMessage, _ internal.Printer) {
	// TODO
}

func examineConnectEndStream(_ json.RawMessage, _ internal.Printer) {
	// TODO
}

func examineGRPCEndStream(endStream string, printer internal.Printer) {
	if !strings.HasSuffix(endStream, "\r\n") {
		printer.Printf("grpc-web trailers should end with CRLF but does not")
	}
	// trim trailing CRLF so we don't have to deal with empty final line after we Split below
	endStream = strings.TrimSuffix(endStream, "\r\n")

	endStreamLines := strings.Split(endStream, "\r\n")
	for _, trailerLine := range endStreamLines {
		parts := strings.SplitN(trailerLine, ":", 2)
		if len(parts) != 2 {
			printer.Printf("grpc-web trailers include invalid field (missing colon): %q", trailerLine)
			continue
		}
		key := parts[0]
		// grpc-web protocol explicitly requires lower-case keys in end-stream message
		if key != strings.ToLower(key) {
			printer.Printf("grpc-web trailers include non-lower-case field key: %q", key)
		}
		// Leading and trailing whitespace is allowed, but only space and htab:
		// https://datatracker.ietf.org/doc/html/rfc7230#section-3.2
		val := strings.Trim(parts[1], " \t")
		if !isValidHTTPFieldValue(val) {
			printer.Printf("grpc-web trailers include invalid field (contains invalid non-visible characters): %q", trailerLine)
		}
	}
}

// isValidHTTPFieldValue returns true the given string is a valid
// HTTP header or trailer value.
// https://datatracker.ietf.org/doc/html/rfc7230#section-3.2
func isValidHTTPFieldValue(s string) bool {
	for _, r := range s {
		// Visible range is 32 (SPACE ' ') and up, excluding DEL (127).
		// Horizontal tab (9, '\t') is allowed but outside the visible range.
		if r != '\t' && (r < 32 || r == 127) {
			// not valid
			return false
		}
	}
	return true
}
