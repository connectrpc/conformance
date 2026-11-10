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
	"net/http"
	"strconv"
	"sync"
	"time"

	"connectrpc.com/conformance/internal"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
)

const wsTimeoutParam = "connect-timeout-ms"

// wsHandshakeKey stores the *wsHandshake of a WebSocket RPC in the context.
type wsHandshakeKey struct{}

// wsHandshake is what the reference server observed in a client's WebSocket
// handshake. The checks that depend on it wait for the client's M message,
// which carries the test case name and expectations.
type wsHandshake struct {
	request        *http.Request
	invalidTimeout string
}

// withWebSocketHandshake records req, a WebSocket handshake, in its context.
// It removes the deadline from the handshake URI, so that the server does
// not enforce it, and records it so the RPC can still report it. That lets
// the test check that the client enforces its own deadline.
func withWebSocketHandshake(req *http.Request) *http.Request {
	handshake := &wsHandshake{request: req}
	ctx := context.WithValue(req.Context(), wsHandshakeKey{}, handshake)
	query := req.URL.Query()
	if values, ok := query[wsTimeoutParam]; ok {
		millis, err := strconv.ParseInt(values[0], 10, 64)
		if err != nil || millis < 0 || len(values[0]) > 10 || len(values) > 1 {
			// connect-go rejects it, and the RPC reports the error.
			handshake.invalidTimeout = req.URL.RawQuery
		} else {
			ctx = contextWithTimeout(ctx, time.Duration(millis)*time.Millisecond)
			query.Del(wsTimeoutParam)
			url := *req.URL
			url.RawQuery = query.Encode()
			req = req.Clone(ctx)
			req.URL = &url
			handshake.request = req
			return req
		}
	}
	return req.WithContext(ctx)
}

// webSocketChecks is the counterpart of referenceServerChecks for an RPC
// carried over WebSocket, whose headers arrive in its M message.
func webSocketChecks(errPrinter internal.Printer, calls *callCounter) connect.ServerInterceptor {
	return func(next connect.ServerFunc) connect.ServerFunc {
		return func(ctx context.Context, spec connect.Spec, stream connect.ServerStream) error {
			handshake, ok := ctx.Value(wsHandshakeKey{}).(*wsHandshake)
			info, hasInfo := connect.CallInfoForServerContext(ctx)
			if !ok || !hasInfo {
				return next(ctx, spec, stream)
			}
			header := make(http.Header)
			for key, values := range info.RequestHeader().All() {
				header[http.CanonicalHeaderKey(key)] = values
			}
			testCaseName := header.Get("X-Test-Case-Name")
			if testCaseName == "" {
				return connect.NewError(connect.CodeInvalidArgument, "missing x-test-case-name header")
			}
			feedback := &feedbackPrinter{p: errPrinter, testCaseName: testCaseName}
			if writer, ok := ctx.Value(wsRawResponseKey{}).(*wsRawWriter); ok {
				writer.setFeedback(feedback)
			}
			if count := calls.add(testCaseName); count > 1 {
				feedback.Printf("client sent another request (#%d) for the same test case", count)
			}
			handshake.check(ctx, info, header, feedback)
			return next(ctx, spec, stream)
		}
	}
}

func (h *wsHandshake) check(ctx context.Context, info *connect.CallInfo, header http.Header, feedback *feedbackPrinter) {
	if h.invalidTimeout != "" {
		feedback.Printf("invalid %s query parameter in handshake URI query %q", wsTimeoutParam, h.invalidTimeout)
	}
	if httpVersion, ok := enumValue("X-Expect-Http-Version", header, conformancev1.HTTPVersion(0), feedback); ok {
		checkHTTPVersion(httpVersion, h.request, feedback)
	}
	if protocol, ok := enumValue("X-Expect-Protocol", header, conformancev1.Protocol(0), feedback); ok &&
		protocol != conformancev1.Protocol_PROTOCOL_CONNECT_WEBSOCKET {
		feedback.Printf("expected protocol %v; instead got %v", protocol, conformancev1.Protocol_PROTOCOL_CONNECT_WEBSOCKET)
	}
	if codec, ok := enumValue("X-Expect-Codec", header, conformancev1.Codec(0), feedback); ok {
		expect := codecProto
		if codec == conformancev1.Codec_CODEC_JSON {
			expect = codecJSON
		}
		if info.Codec != expect {
			feedback.Printf("expected codec %q; instead got %q", expect, info.Codec)
		}
	}
	if compression, ok := enumValue("X-Expect-Compression", header, conformancev1.Compression(0), feedback); ok &&
		compression != conformancev1.Compression_COMPRESSION_IDENTITY {
		feedback.Printf("expected compression %v, which Connect over WebSocket does not have", compression)
	}
	// The expectations arrive in the M message rather than the handshake.
	withMetadata := h.request.Clone(ctx)
	withMetadata.Header = header
	checkTLS(withMetadata, feedback)
	if expectedMethod, _ := getHeader(header, "X-Expect-Http-Method", feedback); h.request.Method != expectedMethod {
		feedback.Printf("expected HTTP method %q, got %q", expectedMethod, h.request.Method)
	}
}

// isWebSocketUpgrade reports whether req is a WebSocket handshake.
func isWebSocketUpgrade(req *http.Request) bool {
	return connecthttp.IsWebSocketUpgrade(req)
}

// callCounter counts the requests a client sends for each test case.
type callCounter struct {
	mu    sync.Mutex
	calls map[string]int
}

func newCallCounter() *callCounter {
	return &callCounter{calls: map[string]int{}}
}

// add records a request for testCaseName and returns how many there have
// been, including this one.
func (c *callCounter) add(testCaseName string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls[testCaseName]++
	return c.calls[testCaseName]
}
