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

package referenceserver

import (
	"context"
	"errors"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"connectrpc.com/conformance/internal"
	"connectrpc.com/conformance/internal/compression"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	grpcContentType                = "application/grpc"
	grpcContentTypePrefix          = grpcContentType + "+"
	grpcWebContentType             = "application/grpc-web"
	grpcWebContentTypePrefix       = grpcWebContentType + "+"
	connectUnaryContentTypePrefix  = "application/"
	connectStreamContentTypePrefix = "application/connect+"
	connectContentTypePrefix       = connectUnaryContentTypePrefix

	connectTimeoutHeader = "Connect-Timeout-Ms"
	grpcTimeoutHeader    = "Grpc-Timeout"

	codecProto = "proto"
	codecJSON  = "json"
)

type int32Enum interface {
	~int32
	protoreflect.Enum
}

type feedbackPrinter struct {
	p            internal.Printer
	testCaseName string
}

func (p *feedbackPrinter) Printf(format string, args ...any) {
	p.p.PrefixPrintf(p.testCaseName, format, args...)
}

type timeoutContextKey struct{}

// TODO - We should add a check for the Connect version header and/or query param to the reference server checks
// to verify that conformant client implementations always include it (to maximize inter-op, just in case a server is
// configured to require it).
func referenceServerChecks(handler http.Handler, errPrinter internal.Printer) http.HandlerFunc {
	var callsMu sync.Mutex
	calls := map[string]int{}
	return func(respWriter http.ResponseWriter, req *http.Request) {
		testCaseName, ok := getTestCaseName(respWriter, req)
		if !ok {
			// This is the only hard failure. Without the test case name, we cannot provide feedback.
			// All other checks below write to stderr to provide feedback and require the test case name.
			return
		}
		feedback := &feedbackPrinter{p: errPrinter, testCaseName: testCaseName}

		callsMu.Lock()
		count := calls[testCaseName]
		calls[testCaseName] = count + 1
		callsMu.Unlock()
		if count > 0 {
			feedback.Printf("client sent another request (#%d) for the same test case", count+1)
		}

		if httpVersion, ok := enumValue("X-Expect-Http-Version", req.Header, conformancev1.HTTPVersion(0), feedback); ok {
			checkHTTPVersion(httpVersion, req, feedback)
		}
		if protocol, ok := enumValue("X-Expect-Protocol", req.Header, conformancev1.Protocol(0), feedback); ok {
			checkProtocol(protocol, req, feedback)
			if timeout, ok := extractTimeout(req.Header, protocol, feedback); ok {
				// In reference mode, we *remove* the timeout in this middleware so that the server
				// will NOT enforce it. That way, we can test that the client is actually enforcing it.
				// We record the timeout in a context value, so that we can correctly include it in the
				// RPC response's request info.
				req = req.WithContext(contextWithTimeout(req.Context(), timeout))
			}
		}
		if codec, ok := enumValue("X-Expect-Codec", req.Header, conformancev1.Codec(0), feedback); ok {
			checkCodec(codec, req, feedback)
		}
		if compress, ok := enumValue("X-Expect-Compression", req.Header, conformancev1.Compression(0), feedback); ok {
			checkCompression(compress, req, feedback)
		}

		checkTLS(req, feedback)

		if expectedMethod, _ := getHeader(req.Header, "X-Expect-Http-Method", feedback); req.Method != expectedMethod {
			feedback.Printf("expected HTTP method %q, got %q", expectedMethod, req.Method)
		}

		handler.ServeHTTP(respWriter, req)

		// Make sure request body is drained so we can look for any trailers.
		// This is just best effort since the operation could have already been canceled.
		_, _ = io.Copy(io.Discard, req.Body)
		if len(req.Trailer) > 0 {
			feedback.Printf("request should NOT include any HTTP trailers (%d trailer keys found)", len(req.Trailer))
		}
	}
}

func enumValue[E int32Enum](headerName string, headers http.Header, zero E, feedback *feedbackPrinter) (E, bool) {
	val, _ := getHeader(headers, headerName, feedback)
	intVal, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		feedback.Printf("invalid value for %q header: %q: %v", headerName, val, err)
		return 0, false
	}
	if zero.Descriptor().Values().ByNumber(protoreflect.EnumNumber(intVal)) == nil {
		feedback.Printf("invalid value for %q header: %d is not in range", headerName, intVal)
		return 0, false
	}
	return E(int32(intVal)), true
}

func checkHTTPVersion(expected conformancev1.HTTPVersion, req *http.Request, feedback *feedbackPrinter) {
	var expectVersion int
	switch expected {
	case conformancev1.HTTPVersion_HTTP_VERSION_1:
		expectVersion = 1
	case conformancev1.HTTPVersion_HTTP_VERSION_2:
		expectVersion = 2
	case conformancev1.HTTPVersion_HTTP_VERSION_3:
		expectVersion = 3
	default:
		feedback.Printf("invalid expected HTTP version %d", expected)
		return
	}
	if req.ProtoMajor != expectVersion {
		feedback.Printf("expected HTTP version %d; instead got %d", expectVersion, req.ProtoMajor)
	}
}

func checkProtocol(expected conformancev1.Protocol, req *http.Request, feedback *feedbackPrinter) {
	var actual conformancev1.Protocol
	contentType := req.Header.Get("Content-Type")
	switch {
	case contentType == grpcContentType || strings.HasPrefix(contentType, grpcContentTypePrefix):
		actual = conformancev1.Protocol_PROTOCOL_GRPC
	case contentType == grpcWebContentType || strings.HasPrefix(contentType, grpcWebContentTypePrefix):
		actual = conformancev1.Protocol_PROTOCOL_GRPC_WEB
	case strings.HasPrefix(contentType, connectContentTypePrefix) || req.Method == http.MethodGet:
		actual = conformancev1.Protocol_PROTOCOL_CONNECT
	default:
		feedback.Printf("could not determine protocol from content-type %q", contentType)
		return
	}
	if expected != actual {
		feedback.Printf("expected protocol %v; instead got %v", expected, actual)
	} else if expected == conformancev1.Protocol_PROTOCOL_GRPC && req.Header.Get("Te") != "trailers" {
		feedback.Printf("gRPC protocol client should use 'te: trailers' header to indicate to proxies that it expects trailer")
	}
}

func checkCodec(expected conformancev1.Codec, req *http.Request, feedback *feedbackPrinter) {
	var expect string
	switch expected {
	case conformancev1.Codec_CODEC_PROTO:
		expect = codecProto
	case conformancev1.Codec_CODEC_JSON:
		expect = codecJSON
	default:
		feedback.Printf("invalid expected codec %d", expected)
		return
	}
	contentType, hasContentType := getHeader(req.Header, "Content-Type", feedback)
	var actual string
	switch {
	case req.Method == http.MethodGet:
		// GET requests should not have a Content-Type header
		if hasContentType {
			feedback.Printf("content-type header should not appear with method GET")
		}
		// Servers should test for an empty request body by attempting a read.
		// If no body is present, it should return an immediate EOF.
		_, err := req.Body.Read([]byte{})
		if !errors.Is(err, io.EOF) {
			feedback.Printf("GET methods should not have a request body")
		}
		var hasActual bool
		actual, hasActual = getQueryParam(req.URL.Query(), "encoding", feedback)
		if !hasActual {
			feedback.Printf("encoding query parameter is missing")
			return
		}
	case contentType == "application/grpc" || contentType == "application/grpc-web":
		actual = codecProto // these protocols default to proto if they have no "+codec" suffix
	case strings.HasPrefix(contentType, "application/grpc+"):
		actual = strings.TrimPrefix(contentType, "application/grpc+")
	case strings.HasPrefix(contentType, "application/grpc-web+"):
		actual = strings.TrimPrefix(contentType, "application/grpc-web+")
	case strings.HasPrefix(contentType, "application/connect+"):
		actual = strings.TrimPrefix(contentType, "application/connect+")
	case strings.HasPrefix(contentType, "application/"):
		actual = strings.TrimPrefix(contentType, "application/")
	default:
		// We already complained about bad content-type when checking protocol.
		return
	}
	if expect != actual {
		feedback.Printf("expected codec %v; instead got %v", expect, actual)
	}
}

func checkCompression(expected conformancev1.Compression, req *http.Request, feedback *feedbackPrinter) {
	var expect string
	switch expected {
	case conformancev1.Compression_COMPRESSION_IDENTITY:
		expect = compression.Identity
	case conformancev1.Compression_COMPRESSION_GZIP:
		expect = compression.Gzip
	case conformancev1.Compression_COMPRESSION_BR:
		expect = compression.Brotli
	case conformancev1.Compression_COMPRESSION_ZSTD:
		expect = compression.Zstd
	case conformancev1.Compression_COMPRESSION_DEFLATE:
		expect = compression.Deflate
	case conformancev1.Compression_COMPRESSION_SNAPPY:
		expect = compression.Snappy
	default:
		feedback.Printf("invalid expected compression %d", expected)
		return
	}
	var actual string
	var hasActual bool
	if req.Method == http.MethodGet {
		actual, hasActual = getQueryParam(req.URL.Query(), "compression", feedback)
	} else {
		contentType := req.Header.Get("Content-Type")
		var encodingHeader string
		switch {
		case contentType == grpcContentType || contentType == grpcWebContentType ||
			strings.HasPrefix(contentType, grpcContentTypePrefix) ||
			strings.HasPrefix(contentType, grpcWebContentTypePrefix):
			encodingHeader = "Grpc-Encoding"
		case strings.HasPrefix(contentType, connectStreamContentTypePrefix):
			encodingHeader = "Connect-Content-Encoding"
		case strings.HasPrefix(contentType, connectUnaryContentTypePrefix):
			encodingHeader = "Content-Encoding"
		default:
			// We already complained about bad content-type when checking protocol.
			return
		}
		actual, hasActual = getHeader(req.Header, encodingHeader, feedback)
	}

	if !hasActual {
		actual = compression.Identity
	}

	if expect != actual {
		feedback.Printf("expected compression %v; instead got %v", expect, actual)
	}
}

func checkTLS(req *http.Request, feedback *feedbackPrinter) {
	tlsHeaderVal, _ := getHeader(req.Header, "X-Expect-Tls", feedback)
	expectTLS, err := strconv.ParseBool(tlsHeaderVal)
	if err != nil {
		feedback.Printf("invalid value for %q header: %q: %v", "x-expect-tls", tlsHeaderVal, err)
		return
	}
	if expectTLS && req.TLS == nil {
		feedback.Printf("expecting TLS request but instead was plain-text")
		return
	} else if !expectTLS && req.TLS != nil {
		feedback.Printf("expecting plain-text request but instead was TLS")
		return
	}
	if req.TLS == nil {
		return
	}
	expectedClientCert, _ := getHeader(req.Header, "X-Expect-Client-Cert", feedback)
	var actualClientCert string
	if len(req.TLS.PeerCertificates) > 0 {
		actualClientCert = req.TLS.PeerCertificates[0].Subject.CommonName
	}
	if expectedClientCert != actualClientCert {
		feedback.Printf("expecting client cert %q, instead was %q", expectedClientCert, actualClientCert)
	}
}

func getHeader(headers http.Header, headerName string, feedback *feedbackPrinter) (string, bool) {
	headerVals := headers.Values(headerName)
	if len(headerVals) > 1 {
		feedback.Printf("%s header appears %d times; should appear just once", headerName, len(headerVals))
	}
	return headers.Get(headerName), len(headerVals) > 0
}

func getQueryParam(values url.Values, paramName string, feedback *feedbackPrinter) (string, bool) {
	paramVals := values[paramName]
	if len(paramVals) > 1 {
		feedback.Printf("%s query string param appears %d times; should appear just once", paramName, len(paramVals))
	}
	return values.Get(paramName), len(paramVals) > 0
}

func getTestCaseName(respWriter http.ResponseWriter, req *http.Request) (string, bool) {
	testCaseName := req.Header.Get("X-Test-Case-Name")
	if testCaseName == "" {
		_ = connect.NewErrorWriter().Write(
			respWriter,
			req,
			connect.NewError(connect.CodeInvalidArgument, errors.New("missing x-test-case-name header")),
		)
		return "", false
	}
	return testCaseName, true
}

// extractTimeout gets the RPC timeout from the headers, removing the header and
// returning the value, if present.
func extractTimeout(headers http.Header, protocol conformancev1.Protocol, feedback *feedbackPrinter) (time.Duration, bool) {
	switch protocol {
	case conformancev1.Protocol_PROTOCOL_CONNECT:
		val, ok := getHeader(headers, connectTimeoutHeader, feedback)
		if !ok {
			break
		}
		headers.Del(connectTimeoutHeader)
		intVal, err := strconv.ParseInt(val, 10, 64)
		if err != nil || intVal < 0 {
			feedback.Printf("invalid numeric value for %q header: %q", connectTimeoutHeader, val)
			break
		}
		if intVal > 9999999999 { // 10 digit max
			feedback.Printf("invalid numeric value (>10 digits) in %q header: %q", connectTimeoutHeader, val)
			break
		}
		timeout := time.Duration(intVal) * time.Millisecond
		if timeout.Milliseconds() != intVal {
			// Overflow. Just use max possible value
			timeout = time.Duration(math.MaxInt64)
		}
		return timeout, true
	case conformancev1.Protocol_PROTOCOL_GRPC, conformancev1.Protocol_PROTOCOL_GRPC_WEB:
		val, ok := getHeader(headers, grpcTimeoutHeader, feedback)
		if !ok {
			break
		}
		headers.Del(grpcTimeoutHeader)
		if val == "" {
			feedback.Printf("invalid value for %q header: %q", grpcTimeoutHeader, val)
			break
		}
		timeoutStr, unit := val[:len(val)-1], rune(val[len(val)-1])
		if !strings.ContainsRune("HMSmun", unit) {
			feedback.Printf("invalid unit in %q header: %q", grpcTimeoutHeader, val)
			break
		}
		intVal, err := strconv.ParseInt(timeoutStr, 10, 64)
		if err != nil || intVal < 0 {
			feedback.Printf("invalid numeric value in %q header: %q", grpcTimeoutHeader, val)
			break
		}
		if intVal > 99999999 { // 8 digit max
			feedback.Printf("invalid numeric value (>8 digits) in %q header: %q", grpcTimeoutHeader, val)
			break
		}
		var timeout time.Duration
		var roundTripped int64
		switch unit {
		case 'H':
			timeout = time.Duration(intVal) * time.Hour
			roundTripped = int64(timeout.Hours())
		case 'M':
			timeout = time.Duration(intVal) * time.Minute
			roundTripped = int64(timeout.Minutes())
		case 'S':
			timeout = time.Duration(intVal) * time.Second
			roundTripped = int64(timeout.Seconds())
		case 'm':
			timeout = time.Duration(intVal) * time.Millisecond
			roundTripped = timeout.Milliseconds()
		case 'u':
			timeout = time.Duration(intVal) * time.Microsecond
			roundTripped = timeout.Microseconds()
		case 'n':
			timeout = time.Duration(intVal) * time.Nanosecond
			roundTripped = timeout.Nanoseconds()
		}
		if roundTripped != intVal {
			// Overflow. Just use max possible value
			timeout = time.Duration(math.MaxInt64)
		}
		return timeout, true
	}
	return 0, false
}

func contextWithTimeout(ctx context.Context, timeout time.Duration) context.Context {
	return context.WithValue(ctx, timeoutContextKey{}, timeout)
}

func timeoutFromContext(ctx context.Context) (time.Duration, bool) {
	timeout, ok := ctx.Value(timeoutContextKey{}).(time.Duration)
	if ok {
		return timeout, ok
	}
	// We use a special value in the context for the timeout in reference mode,
	// in order to test the client's enforcement of the timeout. (We don't want
	// the server enforcing the timeout and client tests passing even if the
	// client doesn't correctly implement timeouts.)
	//
	// But if the value is not there, we may be in non-reference mode, in which
	// case the Connect library uses a normal context deadline. So check that.
	deadline, ok := ctx.Deadline()
	if ok {
		return time.Until(deadline), ok
	}
	return 0, false
}
