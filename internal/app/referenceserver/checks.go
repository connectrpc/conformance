// Copyright 2023 The Connect Authors
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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"connectrpc.com/conformance/internal/compression"
	v2 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v2"
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

	codecProto = "proto"
	codecJSON  = "json"
	codecText  = "text"
)

func referenceServerChecks(handler http.Handler, errWriter io.Writer) http.HandlerFunc {
	return func(respWriter http.ResponseWriter, req *http.Request) {
		testCaseName := req.Header.Get("x-test-case-name")
		if testCaseName == "" {
			// This is the only hard failure. Without it, we cannot provide feedback.
			// All other checks below write to stderr to provide feedback.
			http.Error(respWriter, "missing x-test-case-name header", http.StatusBadRequest)
			return
		}

		feedback := &feedbackWriter{w: errWriter, testCaseName: testCaseName}

		if httpVersion, ok := enumValue("x-expect-http-version", req.Header, v2.HTTPVersion(0), feedback); ok {
			checkHTTPVersion(httpVersion, req, feedback)
		}
		if protocol, ok := enumValue("x-expect-protocol", req.Header, v2.Protocol(0), feedback); ok {
			checkProtocol(protocol, req, feedback)
		}
		if codec, ok := enumValue("x-expect-codec", req.Header, v2.Codec(0), feedback); ok {
			checkCodec(codec, req, feedback)
		}
		if compress, ok := enumValue("x-expect-compression", req.Header, v2.Compression(0), feedback); ok {
			checkCompression(compress, req, feedback)
		}

		checkTLS(req, feedback)

		if expectedMethod, _ := getHeader(req.Header, "x-expect-http-method", feedback); req.Method != expectedMethod {
			feedback.Printf("expected HTTP method %q, got %q", expectedMethod, req.Method)
		}

		handler.ServeHTTP(respWriter, req)
	}
}

type int32Enum interface {
	~int32
	protoreflect.Enum
}

func enumValue[E int32Enum](headerName string, headers http.Header, zero E, feedback *feedbackWriter) (E, bool) {
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

func checkHTTPVersion(expected v2.HTTPVersion, req *http.Request, feedback *feedbackWriter) {
	var expectVersion int
	switch expected {
	case v2.HTTPVersion_HTTP_VERSION_1:
		expectVersion = 1
	case v2.HTTPVersion_HTTP_VERSION_2:
		expectVersion = 2
	case v2.HTTPVersion_HTTP_VERSION_3:
		expectVersion = 3
	default:
		feedback.Printf("invalid expected HTTP version %d", expected)
		return
	}
	if req.ProtoMajor != expectVersion {
		feedback.Printf("expected HTTP version %d; instead got %d", expectVersion, req.ProtoMajor)
	}
}

func checkProtocol(expected v2.Protocol, req *http.Request, feedback *feedbackWriter) {
	var actual v2.Protocol
	contentType := req.Header.Get("content-type")
	switch {
	case contentType == grpcContentType || strings.HasPrefix(contentType, grpcContentTypePrefix):
		actual = v2.Protocol_PROTOCOL_GRPC
	case contentType == grpcWebContentType || strings.HasPrefix(contentType, grpcWebContentTypePrefix):
		actual = v2.Protocol_PROTOCOL_GRPC_WEB
	case strings.HasPrefix(contentType, connectContentTypePrefix) || req.Method == http.MethodGet:
		actual = v2.Protocol_PROTOCOL_CONNECT
	default:
		feedback.Printf("could not determine protocol from content-type %q", contentType)
		return
	}
	if expected != actual {
		feedback.Printf("expected protocol %v; instead got %v", expected, actual)
	}
}

func checkCodec(expected v2.Codec, req *http.Request, feedback *feedbackWriter) {
	var expect string
	switch expected {
	case v2.Codec_CODEC_PROTO:
		expect = codecProto
	case v2.Codec_CODEC_JSON:
		expect = codecJSON
	case v2.Codec_CODEC_TEXT:
		expect = codecText
	default:
		feedback.Printf("invalid expected codec %d", expected)
		return
	}
	contentType, hasContentType := getHeader(req.Header, "content-type", feedback)
	var actual string
	switch {
	case req.Method == http.MethodGet:
		if hasContentType {
			feedback.Printf("content-type header should not appear with method GET")
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

func checkCompression(expected v2.Compression, req *http.Request, feedback *feedbackWriter) {
	var expect string
	switch expected {
	case v2.Compression_COMPRESSION_IDENTITY:
		expect = compression.Identity
	case v2.Compression_COMPRESSION_GZIP:
		expect = compression.Gzip
	case v2.Compression_COMPRESSION_BR:
		expect = compression.Brotli
	case v2.Compression_COMPRESSION_ZSTD:
		expect = compression.Zstd
	case v2.Compression_COMPRESSION_DEFLATE:
		expect = compression.Deflate
	case v2.Compression_COMPRESSION_SNAPPY:
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
		contentType := req.Header.Get("content-type")
		var encodingHeader string
		switch {
		case contentType == grpcContentType || contentType == grpcWebContentType ||
			strings.HasPrefix(contentType, grpcContentTypePrefix) ||
			strings.HasPrefix(contentType, grpcWebContentTypePrefix):
			encodingHeader = "grpc-encoding"
		case strings.HasPrefix(contentType, connectStreamContentTypePrefix):
			encodingHeader = "connect-content-encoding"
		case strings.HasPrefix(contentType, connectUnaryContentTypePrefix):
			encodingHeader = "content-encoding"
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

func checkTLS(req *http.Request, feedback *feedbackWriter) {
	tlsHeaderVal, _ := getHeader(req.Header, "x-expect-tls", feedback)
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
	expectedClientCert, _ := getHeader(req.Header, "x-expect-client-cert", feedback)
	var actualClientCert string
	if len(req.TLS.PeerCertificates) > 0 {
		actualClientCert = req.TLS.PeerCertificates[0].Subject.CommonName
	}
	if expectedClientCert != actualClientCert {
		feedback.Printf("expecting client cert %q, instead was %q", expectedClientCert, actualClientCert)
	}
}

func getHeader(headers http.Header, headerName string, feedback *feedbackWriter) (string, bool) {
	headerVals := headers.Values(headerName)
	if len(headerVals) > 1 {
		feedback.Printf("%s header appears %d times; should appear just once", headerName, len(headerVals))
	}
	return headers.Get(headerName), len(headerVals) > 0
}

func getQueryParam(values url.Values, paramName string, feedback *feedbackWriter) (string, bool) {
	paramVals := values[paramName]
	if len(paramVals) > 1 {
		feedback.Printf("%s query string param appears %d times; should appear just once", paramName, len(paramVals))
	}
	return values.Get(paramName), len(paramVals) > 0
}

type feedbackWriter struct {
	w            io.Writer
	testCaseName string
}

func (w *feedbackWriter) Printf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	_, _ = fmt.Fprintf(w.w, "%s: %s", w.testCaseName, msg)
}
