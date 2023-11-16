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

package connectconformance

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"connectrpc.com/conformance/internal"
	conformancev2 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v2"
	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestPopulateExpectedResponse(t *testing.T) {
	t.Parallel()

	requestHeaders := []*conformancev2.Header{
		{
			Name:  "reqHeader",
			Value: []string{"reqHeaderVal"},
		},
	}

	responseHeaders := []*conformancev2.Header{
		{
			Name:  "fooHeader",
			Value: []string{"fooHeaderVal"},
		},
		{
			Name:  "barHeader",
			Value: []string{"barHeaderVal1", "barHeaderVal2"},
		},
	}

	responseTrailers := []*conformancev2.Header{
		{
			Name:  "fooTrailer",
			Value: []string{"fooTrailerVal"},
		},
		{
			Name:  "barTrailer",
			Value: []string{"barTrailerVal1", "barTrailerVal2"},
		},
	}
	data1 := []byte("data1")
	data2 := []byte("data2")

	errorDef := &conformancev2.Error{
		Code:    int32(connect.CodeResourceExhausted),
		Message: "all resources exhausted",
	}

	// Unary Response Definitions
	unaryErrorResp := &conformancev2.UnaryResponseDefinition_Error{
		Error: errorDef,
	}
	unarySuccessDef := &conformancev2.UnaryResponseDefinition{
		ResponseHeaders: responseHeaders,
		Response: &conformancev2.UnaryResponseDefinition_ResponseData{
			ResponseData: data1,
		},
		ResponseTrailers: responseTrailers,
	}
	unaryErrorDef := &conformancev2.UnaryResponseDefinition{
		ResponseHeaders:  responseHeaders,
		Response:         unaryErrorResp,
		ResponseTrailers: responseTrailers,
	}
	unaryNoResponseDef := &conformancev2.UnaryResponseDefinition{
		ResponseHeaders:  responseHeaders,
		ResponseTrailers: responseTrailers,
	}
	// Stream Response Definitions
	streamSuccessDef := &conformancev2.StreamResponseDefinition{
		ResponseHeaders:  responseHeaders,
		ResponseData:     [][]byte{data1, data2},
		ResponseDelayMs:  1000,
		ResponseTrailers: responseTrailers,
	}
	streamErrorDef := &conformancev2.StreamResponseDefinition{
		ResponseHeaders:  responseHeaders,
		ResponseData:     [][]byte{data1, data2},
		ResponseDelayMs:  1000,
		Error:            errorDef,
		ResponseTrailers: responseTrailers,
	}
	streamNoResponseDef := &conformancev2.StreamResponseDefinition{
		ResponseHeaders:  responseHeaders,
		ResponseDelayMs:  1000,
		ResponseTrailers: responseTrailers,
	}

	// Requests

	// Unary Requests
	unarySuccessReq, err := anypb.New(&conformancev2.UnaryRequest{
		ResponseDefinition: unarySuccessDef,
	})
	require.NoError(t, err)

	unaryErrorReq, err := anypb.New(&conformancev2.UnaryRequest{
		ResponseDefinition: unaryErrorDef,
	})
	require.NoError(t, err)

	unaryNoResponseReq, err := anypb.New(&conformancev2.UnaryRequest{
		ResponseDefinition: unaryNoResponseDef,
	})
	require.NoError(t, err)

	unaryNoDefReq, err := anypb.New(&conformancev2.UnaryRequest{})
	require.NoError(t, err)

	// Client Stream Requests
	clientStreamSuccessReq, err := anypb.New(&conformancev2.ClientStreamRequest{
		ResponseDefinition: unarySuccessDef,
		RequestData:        data1,
	})
	require.NoError(t, err)

	clientStreamErrorReq, err := anypb.New(&conformancev2.ClientStreamRequest{
		ResponseDefinition: unaryErrorDef,
		RequestData:        data1,
	})
	require.NoError(t, err)

	clientStreamNoResponseReq, err := anypb.New(&conformancev2.ClientStreamRequest{
		ResponseDefinition: unaryNoResponseDef,
		RequestData:        data1,
	})
	require.NoError(t, err)

	clientStreamNoDefReq, err := anypb.New(&conformancev2.ClientStreamRequest{
		RequestData: data1,
	})
	require.NoError(t, err)

	clientStreamReq2, err := anypb.New(&conformancev2.ClientStreamRequest{
		RequestData: data1,
	})
	require.NoError(t, err)

	// Server Stream Requests
	serverStreamSuccessReq, err := anypb.New(&conformancev2.ServerStreamRequest{
		ResponseDefinition: streamSuccessDef,
	})
	require.NoError(t, err)

	serverStreamErrorReq, err := anypb.New(&conformancev2.ServerStreamRequest{
		ResponseDefinition: streamErrorDef,
	})
	require.NoError(t, err)

	serverStreamNoResponseReq, err := anypb.New(&conformancev2.ServerStreamRequest{
		ResponseDefinition: streamNoResponseDef,
	})
	require.NoError(t, err)

	serverStreamNoDefReq, err := anypb.New(&conformancev2.ServerStreamRequest{})
	require.NoError(t, err)

	// Bidi Stream Requests
	bidiStreamHalfDuplexSuccessReq, err := anypb.New(&conformancev2.BidiStreamRequest{
		ResponseDefinition: streamSuccessDef,
		RequestData:        data1,
	})
	require.NoError(t, err)

	bidiStreamReq2, err := anypb.New(&conformancev2.BidiStreamRequest{
		RequestData: data2,
	})
	require.NoError(t, err)

	bidiStreamHalfDuplexErrorReq, err := anypb.New(&conformancev2.BidiStreamRequest{
		ResponseDefinition: streamErrorDef,
		RequestData:        data1,
	})
	require.NoError(t, err)

	bidiStreamHalfDuplexNoResponseReq, err := anypb.New(&conformancev2.BidiStreamRequest{
		ResponseDefinition: streamNoResponseDef,
		RequestData:        data1,
	})
	require.NoError(t, err)

	bidiStreamHalfDuplexNoDefReq, err := anypb.New(&conformancev2.BidiStreamRequest{
		RequestData: data1,
	})
	require.NoError(t, err)

	bidiStreamFullDuplexSuccessReq, err := anypb.New(&conformancev2.BidiStreamRequest{
		ResponseDefinition: streamSuccessDef,
		RequestData:        data1,
		FullDuplex:         true,
	})
	require.NoError(t, err)

	bidiStreamFullDuplexErrorReq, err := anypb.New(&conformancev2.BidiStreamRequest{
		ResponseDefinition: streamErrorDef,
		RequestData:        data1,
		FullDuplex:         true,
	})
	require.NoError(t, err)

	bidiStreamFullDuplexNoResponseReq, err := anypb.New(&conformancev2.BidiStreamRequest{
		ResponseDefinition: streamNoResponseDef,
		RequestData:        data1,
		FullDuplex:         true,
	})
	require.NoError(t, err)

	bidiStreamFullDuplexNoDefReq, err := anypb.New(&conformancev2.BidiStreamRequest{
		RequestData: data1,
		FullDuplex:  true,
	})
	require.NoError(t, err)

	testCases := []struct {
		testName   string
		request    *conformancev2.ClientCompatRequest
		expected   *conformancev2.ClientResponseResult
		requireErr bool
	}{
		{
			testName: "unary success",
			request: &conformancev2.ClientCompatRequest{
				StreamType:      conformancev2.StreamType_STREAM_TYPE_UNARY,
				RequestMessages: []*anypb.Any{unarySuccessReq},
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev2.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Payloads: []*conformancev2.ConformancePayload{
					{
						Data: data1,
						RequestInfo: &conformancev2.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests:       []*anypb.Any{unarySuccessReq},
						},
					},
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "unary error",
			request: &conformancev2.ClientCompatRequest{
				StreamType:      conformancev2.StreamType_STREAM_TYPE_UNARY,
				RequestMessages: []*anypb.Any{unaryErrorReq},
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev2.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				// TODO - Payloads will be in error detail for unary response errors
				// Payloads: []*conformancev2.ConformancePayload{{}}
				Error:            unaryErrorResp.Error,
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "unary no response set",
			request: &conformancev2.ClientCompatRequest{
				StreamType:      conformancev2.StreamType_STREAM_TYPE_UNARY,
				RequestMessages: []*anypb.Any{unaryNoResponseReq},
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev2.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Payloads: []*conformancev2.ConformancePayload{
					{
						RequestInfo: &conformancev2.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests:       []*anypb.Any{unaryNoResponseReq},
						},
					},
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "unary no definition set",
			request: &conformancev2.ClientCompatRequest{
				StreamType:      conformancev2.StreamType_STREAM_TYPE_UNARY,
				RequestMessages: []*anypb.Any{unaryNoDefReq},
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev2.ClientResponseResult{
				Payloads: []*conformancev2.ConformancePayload{
					{
						RequestInfo: &conformancev2.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests:       []*anypb.Any{unaryNoDefReq},
						},
					},
				},
			},
		},
		{
			testName: "client stream success",
			request: &conformancev2.ClientCompatRequest{
				StreamType:      conformancev2.StreamType_STREAM_TYPE_CLIENT_STREAM,
				RequestMessages: []*anypb.Any{clientStreamSuccessReq, clientStreamReq2},
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev2.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Payloads: []*conformancev2.ConformancePayload{
					{
						Data: data1,
						RequestInfo: &conformancev2.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests:       []*anypb.Any{clientStreamSuccessReq, clientStreamReq2},
						},
					},
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "client stream error",
			request: &conformancev2.ClientCompatRequest{
				StreamType:      conformancev2.StreamType_STREAM_TYPE_CLIENT_STREAM,
				RequestMessages: []*anypb.Any{clientStreamErrorReq, clientStreamReq2},
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev2.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				// TODO - Payloads will be in error detail for unary response errors
				// Payloads: []*conformancev2.ConformancePayload{{}}
				Error:            unaryErrorResp.Error,
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "client stream no response set",
			request: &conformancev2.ClientCompatRequest{
				StreamType:      conformancev2.StreamType_STREAM_TYPE_CLIENT_STREAM,
				RequestMessages: []*anypb.Any{clientStreamNoResponseReq, clientStreamReq2},
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev2.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Payloads: []*conformancev2.ConformancePayload{
					{
						RequestInfo: &conformancev2.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests:       []*anypb.Any{clientStreamNoResponseReq, clientStreamReq2},
						},
					},
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "client stream no definition set",
			request: &conformancev2.ClientCompatRequest{
				StreamType:      conformancev2.StreamType_STREAM_TYPE_CLIENT_STREAM,
				RequestMessages: []*anypb.Any{clientStreamNoDefReq, clientStreamReq2},
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev2.ClientResponseResult{
				Payloads: []*conformancev2.ConformancePayload{
					{
						RequestInfo: &conformancev2.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests:       []*anypb.Any{clientStreamNoDefReq, clientStreamReq2},
						},
					},
				},
			},
		},
		{
			testName: "server stream success",
			request: &conformancev2.ClientCompatRequest{
				StreamType:      conformancev2.StreamType_STREAM_TYPE_SERVER_STREAM,
				RequestMessages: []*anypb.Any{serverStreamSuccessReq},
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev2.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Payloads: []*conformancev2.ConformancePayload{
					{
						Data: data1,
						RequestInfo: &conformancev2.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests:       []*anypb.Any{serverStreamSuccessReq},
						},
					},
					{
						Data: data2,
					},
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "server stream error",
			request: &conformancev2.ClientCompatRequest{
				StreamType:      conformancev2.StreamType_STREAM_TYPE_SERVER_STREAM,
				RequestMessages: []*anypb.Any{serverStreamErrorReq},
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev2.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Payloads: []*conformancev2.ConformancePayload{
					{
						Data: data1,
						RequestInfo: &conformancev2.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests:       []*anypb.Any{serverStreamErrorReq},
						},
					},
					{
						Data: data2,
					},
				},
				Error:            errorDef,
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "server stream no response set",
			request: &conformancev2.ClientCompatRequest{
				StreamType:      conformancev2.StreamType_STREAM_TYPE_SERVER_STREAM,
				RequestMessages: []*anypb.Any{serverStreamNoResponseReq},
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev2.ClientResponseResult{
				ResponseHeaders:  responseHeaders,
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "server stream no definition set",
			request: &conformancev2.ClientCompatRequest{
				StreamType:      conformancev2.StreamType_STREAM_TYPE_SERVER_STREAM,
				RequestMessages: []*anypb.Any{serverStreamNoDefReq},
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev2.ClientResponseResult{},
		},
		{
			testName: "half duplex bidi stream success",
			request: &conformancev2.ClientCompatRequest{
				StreamType:      conformancev2.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM,
				RequestMessages: []*anypb.Any{bidiStreamHalfDuplexSuccessReq, bidiStreamReq2},
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev2.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Payloads: []*conformancev2.ConformancePayload{
					{
						Data: data1,
						RequestInfo: &conformancev2.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests:       []*anypb.Any{bidiStreamHalfDuplexSuccessReq, bidiStreamReq2},
						},
					},
					{
						Data: data2,
					},
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "half duplex bidi stream error",
			request: &conformancev2.ClientCompatRequest{
				StreamType:      conformancev2.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM,
				RequestMessages: []*anypb.Any{bidiStreamHalfDuplexErrorReq, bidiStreamReq2},
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev2.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Payloads: []*conformancev2.ConformancePayload{
					{
						Data: data1,
						RequestInfo: &conformancev2.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests:       []*anypb.Any{bidiStreamHalfDuplexErrorReq, bidiStreamReq2},
						},
					},
					{
						Data: data2,
					},
				},
				Error:            errorDef,
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "half duplex bidi stream no response set",
			request: &conformancev2.ClientCompatRequest{
				StreamType:      conformancev2.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM,
				RequestMessages: []*anypb.Any{bidiStreamHalfDuplexNoResponseReq},
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev2.ClientResponseResult{
				ResponseHeaders:  responseHeaders,
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "half duplex bidi stream no definition set",
			request: &conformancev2.ClientCompatRequest{
				StreamType:      conformancev2.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM,
				RequestMessages: []*anypb.Any{bidiStreamHalfDuplexNoDefReq},
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev2.ClientResponseResult{},
		},
		{
			testName: "full duplex bidi stream success",
			request: &conformancev2.ClientCompatRequest{
				StreamType:      conformancev2.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
				RequestMessages: []*anypb.Any{bidiStreamFullDuplexSuccessReq, bidiStreamReq2},
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev2.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Payloads: []*conformancev2.ConformancePayload{
					{
						Data: data1,
						RequestInfo: &conformancev2.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests:       []*anypb.Any{bidiStreamFullDuplexSuccessReq},
						},
					},
					{
						Data: data2,
						RequestInfo: &conformancev2.ConformancePayload_RequestInfo{
							Requests: []*anypb.Any{bidiStreamReq2},
						},
					},
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "full duplex bidi stream error",
			request: &conformancev2.ClientCompatRequest{
				StreamType:      conformancev2.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
				RequestMessages: []*anypb.Any{bidiStreamFullDuplexErrorReq, bidiStreamReq2},
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev2.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Payloads: []*conformancev2.ConformancePayload{
					{
						Data: data1,
						RequestInfo: &conformancev2.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests:       []*anypb.Any{bidiStreamFullDuplexErrorReq},
						},
					},
					{
						Data: data2,
						RequestInfo: &conformancev2.ConformancePayload_RequestInfo{
							Requests: []*anypb.Any{bidiStreamReq2},
						},
					},
				},
				Error:            errorDef,
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "full duplex bidi stream no response set",
			request: &conformancev2.ClientCompatRequest{
				StreamType:      conformancev2.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
				RequestMessages: []*anypb.Any{bidiStreamFullDuplexNoResponseReq},
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev2.ClientResponseResult{
				ResponseHeaders:  responseHeaders,
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "full duplex bidi stream no definition set",
			request: &conformancev2.ClientCompatRequest{
				StreamType:      conformancev2.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
				RequestMessages: []*anypb.Any{bidiStreamFullDuplexNoDefReq},
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev2.ClientResponseResult{},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			tc := &conformancev2.TestCase{ //nolint:varnamelen
				Request: testCase.request,
			}
			err := populateExpectedResponse(tc)
			require.NoError(t, err)
			if testCase.requireErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				assert.Empty(t, cmp.Diff(testCase.expected, tc.ExpectedResponse, protocmp.Transform()))
			}
		})
	}
}

func TestRunTestCasesForServer(t *testing.T) {
	t.Parallel()

	var svrResponseBuf bytes.Buffer
	svrResponse := &conformancev2.ServerCompatResponse{
		Host: "127.0.0.1",
		Port: 12345,
	}
	err := internal.WriteDelimitedMessage(&svrResponseBuf, svrResponse)
	require.NoError(t, err)
	svrResponseData := svrResponseBuf.Bytes()

	svrInstance := serverInstance{
		protocol:    conformancev2.Protocol_PROTOCOL_GRPC_WEB,
		httpVersion: conformancev2.HTTPVersion_HTTP_VERSION_1,
		useTLS:      false,
	}
	var expectedSvrReqBuf bytes.Buffer
	err = internal.WriteDelimitedMessage(&expectedSvrReqBuf, &conformancev2.ServerCompatRequest{
		Protocol:            conformancev2.Protocol_PROTOCOL_GRPC_WEB,
		HttpVersion:         conformancev2.HTTPVersion_HTTP_VERSION_1,
		UseTls:              false,
		ClientTlsCert:       nil,
		MessageReceiveLimit: 200 * 1024,
	})
	require.NoError(t, err)
	expectedSvrReqData := expectedSvrReqBuf.Bytes()

	testCaseData := []*conformancev2.TestCase{
		{
			Request: &conformancev2.ClientCompatRequest{
				TestName: "TestSuite1/testcase1",
			},
			ExpectedResponse: &conformancev2.ClientResponseResult{
				Payloads: []*conformancev2.ConformancePayload{{Data: []byte("data")}},
			},
		},
		{
			Request: &conformancev2.ClientCompatRequest{
				TestName: "TestSuite1/testcase2",
			},
			ExpectedResponse: &conformancev2.ClientResponseResult{
				Payloads: []*conformancev2.ConformancePayload{{Data: []byte("data")}},
			},
		},
		{
			Request: &conformancev2.ClientCompatRequest{
				TestName: "TestSuite2/testcase1",
			},
			ExpectedResponse: &conformancev2.ClientResponseResult{
				Error: &conformancev2.Error{Code: int32(connect.CodeAborted), Message: "ruh roh"},
			},
		},
		{
			Request: &conformancev2.ClientCompatRequest{
				TestName: "TestSuite2/testcase2",
			},
			ExpectedResponse: &conformancev2.ClientResponseResult{
				Payloads: []*conformancev2.ConformancePayload{{Data: []byte("data")}},
			},
		},
	}

	requests := make([]*conformancev2.ClientCompatRequest, len(testCaseData))
	responses := make([]*conformancev2.ClientCompatResponse, len(testCaseData))
	for i, testCase := range testCaseData {
		requests[i] = proto.Clone(testCase.Request).(*conformancev2.ClientCompatRequest) //nolint:errcheck,forcetypeassert
		requests[i].Host = svrResponse.Host
		requests[i].Port = svrResponse.Port
		requests[i].ServerTlsCert = svrResponse.PemCert

		if i == 2 {
			responses[i] = &conformancev2.ClientCompatResponse{
				TestName: testCase.Request.TestName,
				Result: &conformancev2.ClientCompatResponse_Error{
					Error: &conformancev2.ClientErrorResult{
						Message: "whoopsy daisy",
					},
				},
			}
		} else {
			responses[i] = &conformancev2.ClientCompatResponse{
				TestName: testCase.Request.TestName,
				Result: &conformancev2.ClientCompatResponse_Response{
					Response: testCase.ExpectedResponse,
				},
			}
		}
	}

	testCases := []struct {
		name              string
		isReferenceServer bool
		svrFailsToStart   bool
		svrErrorReader    io.Reader
		clientCloseAfter  int // close client after num responses read
		svrKillAfter      int // kill server process after num requests sent to client

		expectResults map[string]bool
	}{
		{
			name: "normal",
			expectResults: map[string]bool{
				"TestSuite1/testcase1": true,
				"TestSuite1/testcase2": true,
				"TestSuite2/testcase1": false,
				"TestSuite2/testcase2": true,
			},
		},
		{
			name:              "server sends sideband info",
			isReferenceServer: true,
			svrErrorReader: strings.NewReader(strings.Join([]string{
				"TestSuite1/testcase1: server didn't like this request",
				"This line is ignored because it doesn't look right",
				"Blah:Blah/blah: ignored because this isn't a valid test case name",
			}, "\n")),
			expectResults: map[string]bool{
				"TestSuite1/testcase1": false, // error due to sideband info
				"TestSuite1/testcase2": true,
				"TestSuite2/testcase1": false,
				"TestSuite2/testcase2": true,
			},
		},
		{
			name:            "server fails to start",
			svrFailsToStart: true,
			expectResults: map[string]bool{
				"TestSuite1/testcase1": false,
				"TestSuite1/testcase2": false,
				"TestSuite2/testcase1": false,
				"TestSuite2/testcase2": false,
			},
		},
		{
			name:         "server crashes",
			svrKillAfter: 1,
			expectResults: map[string]bool{
				"TestSuite1/testcase1": true,
				"TestSuite1/testcase2": false, // rest fail due to server crash
				"TestSuite2/testcase1": false,
				"TestSuite2/testcase2": false,
			},
		},
		{
			name:             "client crashes",
			clientCloseAfter: 2,
			expectResults: map[string]bool{
				"TestSuite1/testcase1": true,
				"TestSuite1/testcase2": true,
				"TestSuite2/testcase1": false, // rest fail due to client crash
				"TestSuite2/testcase2": false,
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			results := newResults(&knownFailingTrie{})

			var procAddr atomic.Pointer[process] // populated when server process created
			var actualSvrRequest bytes.Buffer
			var svrProcess processStarter
			if testCase.svrFailsToStart {
				svrProcess = newStillbornProcess(&actualSvrRequest, strings.NewReader("oops"), strings.NewReader("oops"))
			} else {
				svrProcess = newFakeProcess(&actualSvrRequest, bytes.NewReader(svrResponseData), testCase.svrErrorReader)
			}
			hookedProcess := func(ctx context.Context, pipeStderr bool) (*process, error) {
				proc, err := svrProcess(ctx, pipeStderr)
				// capture the process when it is created, so we have a way to kill it
				// after a certain amount of request messages are written.
				if err == nil && !procAddr.CompareAndSwap(nil, proc) {
					return nil, errors.New("process already created!?")
				}
				return proc, err
			}

			var client fakeClient
			expectedRequests := requests
			if testCase.svrKillAfter > 0 {
				client.requestHook = func() {
					procAddr.Load().abort()
				}
				client.requestHookCount = testCase.svrKillAfter
				expectedRequests = requests[:testCase.svrKillAfter]
			}
			if testCase.isReferenceServer {
				copyOfRequests := make([]*conformancev2.ClientCompatRequest, len(expectedRequests))
				// runner adds headers for the reference server
				for i, req := range expectedRequests {
					req = proto.Clone(req).(*conformancev2.ClientCompatRequest) //nolint:errcheck,forcetypeassert
					req.RequestHeaders = append(req.RequestHeaders,
						&conformancev2.Header{Name: "x-test-case-name", Value: []string{req.TestName}},
						// we didn't set this above, so they're all zero/unspecified
						&conformancev2.Header{Name: "x-expect-http-version", Value: []string{"0"}},
						&conformancev2.Header{Name: "x-expect-http-method", Value: []string{"POST"}},
						&conformancev2.Header{Name: "x-expect-protocol", Value: []string{"0"}},
						&conformancev2.Header{Name: "x-expect-codec", Value: []string{"0"}},
						&conformancev2.Header{Name: "x-expect-compression", Value: []string{"0"}},
						&conformancev2.Header{Name: "x-expect-tls", Value: []string{"false"}},
					)
					copyOfRequests[i] = req
				}
				expectedRequests = copyOfRequests
			}

			responsesToSend := responses
			if testCase.clientCloseAfter > 0 {
				if len(expectedRequests) > testCase.clientCloseAfter {
					expectedRequests = expectedRequests[:testCase.clientCloseAfter]
				}
				responsesToSend = responses[:testCase.clientCloseAfter]
			}
			client.responses = make(map[string]*conformancev2.ClientCompatResponse, len(responsesToSend))
			for _, resp := range responsesToSend {
				client.responses[resp.TestName] = resp
			}

			runTestCasesForServer(
				context.Background(),
				!testCase.isReferenceServer,
				testCase.isReferenceServer,
				svrInstance,
				testCaseData,
				nil, // TODO: client cert
				hookedProcess,
				results,
				&client,
			)

			if testCase.svrFailsToStart {
				assert.Empty(t, client.actualRequests)
			} else {
				assert.Empty(t, cmp.Diff(expectedRequests, client.actualRequests, protocmp.Transform()))
			}

			assert.Empty(t, cmp.Diff(expectedSvrReqData, actualSvrRequest.Bytes()))

			actualResults := func() map[string]bool {
				res := map[string]bool{}
				results.mu.Lock()
				defer results.mu.Unlock()
				results.processSidebandInfoLocked()
				for name, outcome := range results.outcomes {
					res[name] = outcome.actualFailure == nil
				}
				return res
			}()
			assert.Empty(t, cmp.Diff(testCase.expectResults, actualResults))
		})
	}
}

// fakeProcess is a process starter that represents a fictitious process
// that is runs until the stop method is called.
type fakeProcess struct {
	mu           sync.Mutex
	done         bool
	err          error
	atEndActions []func(error)
}

func newFakeProcess(stdin io.Writer, stdout, stderr io.Reader) processStarter {
	return func(ctx context.Context, pipeStderr bool) (*process, error) {
		proc := &fakeProcess{}
		return &process{
			processController: proc,
			stdin:             &procWriter{w: stdin, proc: proc},
			stdout:            &procReader{r: stdout, proc: proc},
			// Allow stderr to be fully consumed, just in case we try to kill
			// the server process before we've read all sideband info.
			stderr: stderr,
		}, nil
	}
}

func newStillbornProcess(stdin io.Writer, stdout, stderr io.Reader) processStarter {
	return func(ctx context.Context, pipeStderr bool) (*process, error) {
		proc := &fakeProcess{}
		stdout = &hookReader{
			r:    stdout,
			hook: proc.abort,
		}
		return &process{
			processController: proc,
			stdin:             &procWriter{w: stdin, proc: proc},
			stdout:            &procReader{r: stdout, proc: proc},
			stderr:            &procReader{r: stderr, proc: proc},
		}, nil
	}
}

func (f *fakeProcess) stop(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.done {
		return
	}
	f.done = true
	f.err = err
	for _, fn := range f.atEndActions {
		fn(err)
	}
	f.atEndActions = nil
}

func (f *fakeProcess) tryResult() (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.done, f.err
}

func (f *fakeProcess) result() error {
	ch := make(chan struct{})
	f.whenDone(func(_ error) { close(ch) })
	<-ch
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.err
}

func (f *fakeProcess) abort() {
	f.stop(errors.New("process killed by call to abort"))
}

func (f *fakeProcess) whenDone(action func(error)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.done {
		action(f.err)
	}
	f.atEndActions = append(f.atEndActions, action)
}

// procWriter delegates to the given writer but will instead
// immediately return an error if the given process has stopped.
type procWriter struct {
	w      io.Writer
	proc   *fakeProcess
	closed atomic.Bool
}

func (p *procWriter) Write(data []byte) (n int, err error) {
	if p.closed.Load() {
		return 0, errors.New("closed")
	}
	if done, _ := p.proc.tryResult(); done {
		return 0, errors.New("closed")
	}
	return p.w.Write(data)
}

func (p *procWriter) Close() error {
	p.closed.Store(true)
	return nil
}

// procReader delegates to the given reader but will instead
// immediately return an error if the given process has stopped.
type procReader struct {
	r      io.Reader
	proc   *fakeProcess
	closed atomic.Bool
}

func (p *procReader) Read(data []byte) (n int, err error) {
	if p.closed.Load() {
		return 0, io.EOF
	}
	if done, _ := p.proc.tryResult(); done {
		return 0, io.EOF
	}
	return p.r.Read(data)
}

func (p *procReader) Close() error {
	p.closed.Store(true)
	return nil
}

// hookReader calls the given hook function upon reaching EOF.
type hookReader struct {
	r    io.Reader
	hook func()
}

func (h *hookReader) Read(data []byte) (n int, err error) {
	n, err = h.r.Read(data)
	if err != nil && h.hook != nil {
		h.hook()
		// don't run hook 2x
		h.hook = nil
	}
	return n, err
}

// fakeClient immediately calls whenDone when responses that have been prepared
// in the responses field. If requestHookCount is greater than zero and requestHook
// is non-nil, requestHook will be invoked after sendRequest is called requestHookCount
// times.
type fakeClient struct {
	actualRequests   []*conformancev2.ClientCompatRequest
	responses        map[string]*conformancev2.ClientCompatResponse
	requestHookCount int
	requestHook      func()
}

func (f *fakeClient) sendRequest(req *conformancev2.ClientCompatRequest, whenDone func(string, *conformancev2.ClientCompatResponse, error)) error {
	if len(f.responses) == 0 {
		return errors.New("no more")
	}

	f.actualRequests = append(f.actualRequests, req)
	resp := f.responses[req.TestName]
	delete(f.responses, req.TestName)
	if resp != nil {
		whenDone(req.TestName, resp, nil)
	} else {
		whenDone(req.TestName, nil, errors.New("no configured response"))
	}

	f.requestHookCount--
	if f.requestHookCount == 0 && f.requestHook != nil {
		f.requestHook()
	}

	return nil
}

func (f *fakeClient) closeSend() {
	f.responses = nil
}

func (f *fakeClient) waitForResponses() error {
	return nil
}

func (f *fakeClient) isRunning() bool {
	return false
}

func (f *fakeClient) stop() {
}
