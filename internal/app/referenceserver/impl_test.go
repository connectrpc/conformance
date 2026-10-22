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
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
	"time"

	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1/conformancev1connect"
	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestServerNameHandlerInterceptor_GRPCWebTrailersOnly(t *testing.T) {
	t.Parallel()

	svr := newTestServer(t)
	failedPrecondition := &conformancev1.Error{
		Code:    conformancev1.Code_CODE_FAILED_PRECONDITION,
		Message: new("failed"),
	}
	customHeaders := []*conformancev1.Header{{Name: "x-custom", Value: []string{"foo"}}}
	unaryError := &conformancev1.UnaryResponseDefinition{
		Response: &conformancev1.UnaryResponseDefinition_Error{Error: failedPrecondition},
	}

	testCases := []struct {
		name      string
		procedure string
		requests  []proto.Message
		// A trailers-only response is only allowed for an immediate error.
		wantTrailersOnly bool
		wantStatus       string
	}{
		{
			name:             "unary/error",
			procedure:        conformancev1connect.ConformanceServiceUnaryProcedure,
			requests:         []proto.Message{&conformancev1.UnaryRequest{ResponseDefinition: unaryError}},
			wantTrailersOnly: true,
			wantStatus:       "9",
		},
		{
			name:      "unary/error-with-trailers",
			procedure: conformancev1connect.ConformanceServiceUnaryProcedure,
			requests: []proto.Message{&conformancev1.UnaryRequest{
				ResponseDefinition: &conformancev1.UnaryResponseDefinition{
					ResponseTrailers: customHeaders,
					Response:         &conformancev1.UnaryResponseDefinition_Error{Error: failedPrecondition},
				},
			}},
			wantTrailersOnly: true,
			wantStatus:       "9",
		},
		{
			name:      "unary/success",
			procedure: conformancev1connect.ConformanceServiceUnaryProcedure,
			requests: []proto.Message{&conformancev1.UnaryRequest{
				ResponseDefinition: &conformancev1.UnaryResponseDefinition{
					Response: &conformancev1.UnaryResponseDefinition_ResponseData{ResponseData: []byte("data")},
				},
			}},
			wantStatus: "0",
		},
		{
			name:      "client-stream/error",
			procedure: conformancev1connect.ConformanceServiceClientStreamProcedure,
			requests: []proto.Message{
				&conformancev1.ClientStreamRequest{ResponseDefinition: unaryError},
				&conformancev1.ClientStreamRequest{},
			},
			wantTrailersOnly: true,
			wantStatus:       "9",
		},
		{
			name:      "server-stream/error-with-no-responses",
			procedure: conformancev1connect.ConformanceServiceServerStreamProcedure,
			requests: []proto.Message{&conformancev1.ServerStreamRequest{
				ResponseDefinition: &conformancev1.StreamResponseDefinition{Error: failedPrecondition},
			}},
			wantTrailersOnly: true,
			wantStatus:       "9",
		},
		{
			name:      "server-stream/error-with-responses",
			procedure: conformancev1connect.ConformanceServiceServerStreamProcedure,
			requests: []proto.Message{&conformancev1.ServerStreamRequest{
				ResponseDefinition: &conformancev1.StreamResponseDefinition{
					ResponseData: [][]byte{[]byte("data")},
					Error:        failedPrecondition,
				},
			}},
			wantStatus: "9",
		},
		{
			name:      "server-stream/error-with-headers",
			procedure: conformancev1connect.ConformanceServiceServerStreamProcedure,
			requests: []proto.Message{&conformancev1.ServerStreamRequest{
				ResponseDefinition: &conformancev1.StreamResponseDefinition{
					ResponseHeaders: customHeaders,
					Error:           failedPrecondition,
				},
			}},
			wantStatus: "9",
		},
		{
			name:       "server-stream/success-with-no-responses",
			procedure:  conformancev1connect.ConformanceServiceServerStreamProcedure,
			requests:   []proto.Message{&conformancev1.ServerStreamRequest{}},
			wantStatus: "0",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var reqBody bytes.Buffer
			for _, msg := range testCase.requests {
				data, err := proto.Marshal(msg)
				require.NoError(t, err)
				writeGRPCWebFrame(&reqBody, 0, data)
			}
			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, svr.URL+testCase.procedure, &reqBody)
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/grpc-web+proto")
			resp, err := svr.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			if testCase.wantTrailersOnly {
				assert.Empty(t, respBody)
				assert.Equal(t, testCase.wantStatus, resp.Header.Get("Grpc-Status"))
				assert.Contains(t, resp.Header.Get("Server"), serverName)
				return
			}
			assert.Empty(t, resp.Header.Get("Grpc-Status"))
			trailers := readGRPCWebTrailers(t, respBody)
			assert.Equal(t, testCase.wantStatus, trailers.Get("Grpc-Status"))
			assert.Contains(t, resp.Header.Get("Server"), serverName)
		})
	}
}

// The trailer fallback is only for gRPC-Web, since the Connect protocol has no
// trailers-only responses.
func TestServerNameHandlerInterceptor_ConnectErrors(t *testing.T) {
	t.Parallel()

	svr := newTestServer(t)
	failedPrecondition := &conformancev1.Error{
		Code:    conformancev1.Code_CODE_FAILED_PRECONDITION,
		Message: new("failed"),
	}
	unaryError := &conformancev1.UnaryResponseDefinition{
		Response: &conformancev1.UnaryResponseDefinition_Error{Error: failedPrecondition},
	}

	testCases := []struct {
		name        string
		procedure   string
		contentType string
		request     proto.Message
		wantStatus  int
	}{
		{
			name:        "unary/error",
			procedure:   conformancev1connect.ConformanceServiceUnaryProcedure,
			contentType: "application/proto",
			request:     &conformancev1.UnaryRequest{ResponseDefinition: unaryError},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "client-stream/error",
			procedure:   conformancev1connect.ConformanceServiceClientStreamProcedure,
			contentType: "application/connect+proto",
			request:     &conformancev1.ClientStreamRequest{ResponseDefinition: unaryError},
			wantStatus:  http.StatusOK,
		},
		{
			name:        "server-stream/error-with-no-responses",
			procedure:   conformancev1connect.ConformanceServiceServerStreamProcedure,
			contentType: "application/connect+proto",
			request: &conformancev1.ServerStreamRequest{
				ResponseDefinition: &conformancev1.StreamResponseDefinition{Error: failedPrecondition},
			},
			wantStatus: http.StatusOK,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			data, err := proto.Marshal(testCase.request)
			require.NoError(t, err)
			var reqBody bytes.Buffer
			if testCase.contentType == "application/proto" {
				reqBody.Write(data)
			} else {
				writeGRPCWebFrame(&reqBody, 0, data)
			}
			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, svr.URL+testCase.procedure, &reqBody)
			require.NoError(t, err)
			req.Header.Set("Content-Type", testCase.contentType)
			resp, err := svr.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			_, err = io.Copy(io.Discard, resp.Body)
			require.NoError(t, err)
			require.Equal(t, testCase.wantStatus, resp.StatusCode)
			assert.Contains(t, resp.Header.Get("Server"), serverName)
			assert.Empty(t, resp.Header.Get("Trailer-Server"))
		})
	}
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := connect.NewServer(serverNameHandlerInterceptor)
	conformancev1connect.RegisterConformanceServiceHandler(server, &conformanceServer{})
	mux := http.NewServeMux()
	connecthttp.Mount(mux, server)
	svr := httptest.NewServer(mux)
	t.Cleanup(svr.Close)
	return svr
}

func writeGRPCWebFrame(buf *bytes.Buffer, flags byte, data []byte) {
	buf.WriteByte(flags)
	_ = binary.Write(buf, binary.BigEndian, uint32(len(data)))
	buf.Write(data)
}

// readGRPCWebTrailers returns the trailers from the final frame of a gRPC-Web
// response body.
func readGRPCWebTrailers(t *testing.T, body []byte) textproto.MIMEHeader {
	t.Helper()
	for len(body) > 0 {
		require.GreaterOrEqual(t, len(body), 5, "incomplete frame prefix")
		flags := body[0]
		size := int(binary.BigEndian.Uint32(body[1:5]))
		require.GreaterOrEqual(t, len(body)-5, size, "incomplete frame")
		data, rest := body[5:5+size], body[5+size:]
		if flags&0x80 != 0 {
			require.Empty(t, rest, "data after trailers frame")
			data = append(data, "\r\n"...)
			trailers, err := textproto.NewReader(bufio.NewReader(bytes.NewReader(data))).ReadMIMEHeader()
			require.NoError(t, err)
			return trailers
		}
		body = rest
	}
	require.Fail(t, "response body has no trailers frame")
	return nil
}

func TestHeadersReturnedImmediately(t *testing.T) {
	t.Parallel()
	svr := newTestServer(t)
	const responseDelay = 1 * time.Second
	testCases := []struct {
		name        string
		contentType string
	}{
		{
			name:        "connect_server_stream",
			contentType: "application/connect+proto",
		},
		{
			name:        "grpc_web_server_stream",
			contentType: "application/grpc-web+proto",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			data, err := proto.Marshal(&conformancev1.ServerStreamRequest{
				ResponseDefinition: &conformancev1.StreamResponseDefinition{
					ResponseHeaders: []*conformancev1.Header{
						{Name: "x-custom-header", Value: []string{"foo"}},
					},
					// SendHeaders is only called when there is response data.
					ResponseData:    [][]byte{{1, 2, 3}},
					ResponseDelayMs: uint32(responseDelay.Milliseconds()),
				},
			})
			require.NoError(t, err)
			var reqBody bytes.Buffer
			writeGRPCWebFrame(&reqBody, 0, data)
			ctx, cancel := context.WithTimeout(t.Context(), 2*responseDelay)
			t.Cleanup(cancel)
			req, err := http.NewRequestWithContext(
				ctx,
				http.MethodPost,
				svr.URL+conformancev1connect.ConformanceServiceServerStreamProcedure,
				&reqBody,
			)
			require.NoError(t, err)
			req.Header.Set("Content-Type", testCase.contentType)
			start := time.Now()
			// Do returns once the response headers arrive, before the body.
			resp, err := svr.Client().Do(req)
			elapsed := time.Since(start)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, "foo", resp.Header.Get("X-Custom-Header"))
			assert.Less(t, elapsed, responseDelay/2, "headers were not sent before the first response message")
			_, err = io.Copy(io.Discard, resp.Body)
			require.NoError(t, err)
		})
	}
}
