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
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1/conformancev1connect"
	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	ignoreTimeoutHeader = "X-Ignore-Timeout"
	// Wire capture only sees traces for requests with this header, which the
	// test runner always sets.
	testCaseNameHeader = "X-Test-Case-Name"
)

type timeoutTestService struct {
	conformancev1connect.UnimplementedConformanceServiceHandler

	release        chan struct{}
	timeoutHeaders sync.Map // map[string]string, keyed by test case name
}

func (s *timeoutTestService) Unary(
	ctx context.Context,
	req *connect.Request[conformancev1.UnaryRequest],
) (*connect.Response[conformancev1.UnaryResponse], error) {
	s.timeoutHeaders.Store(req.Header().Get(testCaseNameHeader),
		req.Header().Get(connectTimeoutHeader)+req.Header().Get(grpcTimeoutHeader))
	if req.Header().Get(ignoreTimeoutHeader) != "" {
		<-s.release
		return connect.NewResponse(&conformancev1.UnaryResponse{}), nil
	}
	<-ctx.Done()
	return nil, connect.NewError(connect.CodeDeadlineExceeded, ctx.Err())
}

func TestInvoke_ServerTimeout(t *testing.T) {
	t.Parallel()

	service := &timeoutTestService{release: make(chan struct{})}
	mux := http.NewServeMux()
	mux.Handle(conformancev1connect.NewConformanceServiceHandler(service))
	svr := httptest.NewUnstartedServer(mux)
	svr.Config.Protocols = &http.Protocols{}
	svr.Config.Protocols.SetHTTP1(true)
	svr.Config.Protocols.SetUnencryptedHTTP2(true)
	svr.Start()
	t.Cleanup(svr.Close)
	// Runs before svr.Close, which waits for handlers to return.
	t.Cleanup(func() { close(service.release) })

	host, portStr, err := net.SplitHostPort(svr.Listener.Addr().String())
	require.NoError(t, err)
	port, err := strconv.ParseUint(portStr, 10, 32)
	require.NoError(t, err)
	msg, err := anypb.New(&conformancev1.UnaryRequest{})
	require.NoError(t, err)

	timeoutMs := uint32(50)
	testCases := []struct {
		name        string
		httpVersion conformancev1.HTTPVersion
		protocol    conformancev1.Protocol
		wantHeader  string
	}{
		{
			name:        "connect",
			httpVersion: conformancev1.HTTPVersion_HTTP_VERSION_1,
			protocol:    conformancev1.Protocol_PROTOCOL_CONNECT,
			wantHeader:  "50",
		},
		{
			name:        "grpc",
			httpVersion: conformancev1.HTTPVersion_HTTP_VERSION_2,
			protocol:    conformancev1.Protocol_PROTOCOL_GRPC,
			wantHeader:  "50m",
		},
		{
			name:        "grpc-web",
			httpVersion: conformancev1.HTTPVersion_HTTP_VERSION_1,
			protocol:    conformancev1.Protocol_PROTOCOL_GRPC_WEB,
			wantHeader:  "50m",
		},
	}
	for _, testCase := range testCases {
		newRequest := func(t *testing.T, ignoreTimeout bool) *conformancev1.ClientCompatRequest {
			t.Helper()
			req := &conformancev1.ClientCompatRequest{
				Host:            host,
				Port:            uint32(port),
				HttpVersion:     testCase.httpVersion,
				Protocol:        testCase.protocol,
				Codec:           conformancev1.Codec_CODEC_PROTO,
				Compression:     conformancev1.Compression_COMPRESSION_IDENTITY,
				Service:         new(conformancev1connect.ConformanceServiceName),
				Method:          new("Unary"),
				StreamType:      conformancev1.StreamType_STREAM_TYPE_UNARY,
				TimeoutMs:       &timeoutMs,
				RequestMessages: []*anypb.Any{msg},
				RequestHeaders:  []*conformancev1.Header{{Name: testCaseNameHeader, Value: []string{t.Name()}}},
			}
			if ignoreTimeout {
				req.RequestHeaders = append(req.RequestHeaders,
					&conformancev1.Header{Name: ignoreTimeoutHeader, Value: []string{"1"}})
			}
			return req
		}
		assertTimeoutHeader := func(t *testing.T, want string) {
			t.Helper()
			got, ok := service.timeoutHeaders.Load(t.Name())
			require.True(t, ok, "server did not receive request")
			assert.Equal(t, want, got)
		}
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			t.Run("enforced", func(t *testing.T) {
				t.Parallel()
				var transports transports
				result, err := invoke(t.Context(), &transports, newRequest(t, false), true, nil)
				require.NoError(t, err)
				assert.Equal(t, conformancev1.Code_CODE_DEADLINE_EXCEEDED, result.GetError().GetCode())
				assertTimeoutHeader(t, testCase.wantHeader)
			})
			t.Run("ignored", func(t *testing.T) {
				t.Parallel()
				var transports transports
				_, err := invoke(t.Context(), &transports, newRequest(t, true), true, nil)
				require.ErrorIs(t, err, errServerIgnoredTimeout)
				assertTimeoutHeader(t, testCase.wantHeader)
			})
			t.Run("ignored/not-reference-mode", func(t *testing.T) {
				t.Parallel()
				// Outside of reference mode, the client enforces the timeout itself.
				var transports transports
				result, err := invoke(t.Context(), &transports, newRequest(t, true), false, nil)
				require.NoError(t, err)
				assert.Equal(t, conformancev1.Code_CODE_DEADLINE_EXCEEDED, result.GetError().GetCode())
			})
		})
	}
}
