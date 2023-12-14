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
	"encoding/base64"
	"sort"
	"testing"

	"connectrpc.com/conformance/internal"
	"connectrpc.com/conformance/internal/app/connectconformance/testsuites"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestNewTestCaseLibrary(t *testing.T) {
	t.Parallel()

	testData := map[string]string{
		"basic.yaml": `
                    name: Basic
                    testCases:
                      - request:
                            testName: basic-unary
                            streamType: STREAM_TYPE_UNARY
                      - request:
                            testName: basic-client-stream
                            streamType: STREAM_TYPE_CLIENT_STREAM
                      - request:
                            testName: basic-server-stream
                            streamType: STREAM_TYPE_SERVER_STREAM
                      - request:
                            testName: basic-bidi-stream
                            streamType: STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM`,
		"tls.yaml": `
                    name: TLS
                    reliesOnTls: true
                    testCases:
                      - request:
                            testName: tls-unary
                            streamType: STREAM_TYPE_UNARY
                      - request:
                            testName: tls-client-stream
                            streamType: STREAM_TYPE_CLIENT_STREAM
                      - request:
                            testName: tls-server-stream
                            streamType: STREAM_TYPE_SERVER_STREAM
                      - request:
                            testName: tls-bidi-stream
                            streamType: STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM`,
		"tls-client-certs.yaml": `
                    name: TLS Client Certs
                    reliesOnTls: true
                    reliesOnTlsClientCerts: true
                    testCases:
                      - request:
                            testName: tls-client-cert-unary
                            streamType: STREAM_TYPE_UNARY`,
		"connect-get.yaml": `
                    name: Connect GET
                    relevantProtocols: [PROTOCOL_CONNECT]
                    reliesOnConnectGet: true
                    testCases:
                      - request:
                            testName: connect-get-unary
                            streamType: STREAM_TYPE_UNARY`,
		"connect-version-client-required.yaml": `
                    name: Connect Version Required (client)
                    mode: TEST_MODE_CLIENT
                    relevantProtocols: [PROTOCOL_CONNECT]
                    connectVersionMode: CONNECT_VERSION_MODE_REQUIRE
                    testCases:
                      - request:
                            testName: unary-without-connect-version-header
                            streamType: STREAM_TYPE_UNARY`,
		"connect-version-server-required.yaml": `
                    name: Connect Version Required (server)
                    mode: TEST_MODE_SERVER
                    relevantProtocols: [PROTOCOL_CONNECT]
                    connectVersionMode: CONNECT_VERSION_MODE_REQUIRE
                    testCases:
                      - request:
                            testName: unary-without-connect-version-header
                            streamType: STREAM_TYPE_UNARY`,
		"connect-version-client-not-required.yaml": `
                    name: Connect Version Optional (client)
                    mode: TEST_MODE_CLIENT
                    relevantProtocols: [PROTOCOL_CONNECT]
                    connectVersionMode: CONNECT_VERSION_MODE_IGNORE
                    testCases:
                      - request:
                            testName: unary-without-connect-version-header
                            streamType: STREAM_TYPE_UNARY`,
		"connect-version-server-not-required.yaml": `
                    name: Connect Version Optional (server)
                    mode: TEST_MODE_SERVER
                    relevantProtocols: [PROTOCOL_CONNECT]
                    connectVersionMode: CONNECT_VERSION_MODE_IGNORE
                    testCases:
                      - request:
                            testName: unary-without-connect-version-header
                            streamType: STREAM_TYPE_UNARY`,
		"max-receive-limit": `
                    name: Max Receive Size (server)
                    mode: TEST_MODE_SERVER
                    reliesOnMessageReceiveLimit: true
                    testCases:
                      - request:
                            testName: unary-exceeds-limit
                            streamType: STREAM_TYPE_UNARY`,
	}
	testSuiteData := make(map[string][]byte, len(testData))
	for k, v := range testData {
		testSuiteData[k] = []byte(v)
	}
	testSuites, err := parseTestSuites(testSuiteData)
	require.NoError(t, err)

	// there is some repetition, but we want them to be able to
	// vary and evolve independently, so we won't consolidate
	//nolint:dupl
	testCases := []struct {
		name   string
		config []configCase
		mode   conformancev1.TestSuite_TestMode
		cases  map[serverInstance][]string
	}{
		{
			name: "client mode",
			config: []configCase{
				{
					Version:     conformancev1.HTTPVersion_HTTP_VERSION_1,
					Protocol:    conformancev1.Protocol_PROTOCOL_CONNECT,
					Codec:       conformancev1.Codec_CODEC_PROTO,
					Compression: conformancev1.Compression_COMPRESSION_IDENTITY,
					StreamType:  conformancev1.StreamType_STREAM_TYPE_UNARY,
				},
				{
					Version:     conformancev1.HTTPVersion_HTTP_VERSION_1,
					Protocol:    conformancev1.Protocol_PROTOCOL_CONNECT,
					Codec:       conformancev1.Codec_CODEC_PROTO,
					Compression: conformancev1.Compression_COMPRESSION_IDENTITY,
					StreamType:  conformancev1.StreamType_STREAM_TYPE_UNARY,
					UseTLS:      true,
				},
				{
					Version:           conformancev1.HTTPVersion_HTTP_VERSION_1,
					Protocol:          conformancev1.Protocol_PROTOCOL_CONNECT,
					Codec:             conformancev1.Codec_CODEC_PROTO,
					Compression:       conformancev1.Compression_COMPRESSION_IDENTITY,
					StreamType:        conformancev1.StreamType_STREAM_TYPE_UNARY,
					UseTLS:            true,
					UseTLSClientCerts: true,
				},
				{
					Version:       conformancev1.HTTPVersion_HTTP_VERSION_1,
					Protocol:      conformancev1.Protocol_PROTOCOL_CONNECT,
					Codec:         conformancev1.Codec_CODEC_PROTO,
					Compression:   conformancev1.Compression_COMPRESSION_IDENTITY,
					StreamType:    conformancev1.StreamType_STREAM_TYPE_UNARY,
					UseConnectGET: true,
				},
				{
					Version:       conformancev1.HTTPVersion_HTTP_VERSION_1,
					Protocol:      conformancev1.Protocol_PROTOCOL_CONNECT,
					Codec:         conformancev1.Codec_CODEC_PROTO,
					Compression:   conformancev1.Compression_COMPRESSION_IDENTITY,
					StreamType:    conformancev1.StreamType_STREAM_TYPE_UNARY,
					UseConnectGET: true,
					UseTLS:        true,
				},
				{
					Version:            conformancev1.HTTPVersion_HTTP_VERSION_1,
					Protocol:           conformancev1.Protocol_PROTOCOL_CONNECT,
					Codec:              conformancev1.Codec_CODEC_PROTO,
					Compression:        conformancev1.Compression_COMPRESSION_IDENTITY,
					StreamType:         conformancev1.StreamType_STREAM_TYPE_UNARY,
					ConnectVersionMode: conformancev1.TestSuite_CONNECT_VERSION_MODE_REQUIRE,
				},
				{
					Version:                conformancev1.HTTPVersion_HTTP_VERSION_1,
					Protocol:               conformancev1.Protocol_PROTOCOL_CONNECT,
					Codec:                  conformancev1.Codec_CODEC_PROTO,
					Compression:            conformancev1.Compression_COMPRESSION_IDENTITY,
					StreamType:             conformancev1.StreamType_STREAM_TYPE_UNARY,
					UseMessageReceiveLimit: true,
				},
				{
					Version:     conformancev1.HTTPVersion_HTTP_VERSION_2,
					Protocol:    conformancev1.Protocol_PROTOCOL_GRPC,
					Codec:       conformancev1.Codec_CODEC_PROTO,
					Compression: conformancev1.Compression_COMPRESSION_IDENTITY,
					StreamType:  conformancev1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
				},
			},
			mode: conformancev1.TestSuite_TEST_MODE_CLIENT,
			cases: map[serverInstance][]string{
				{
					protocol:    conformancev1.Protocol_PROTOCOL_CONNECT,
					httpVersion: conformancev1.HTTPVersion_HTTP_VERSION_1,
					useTLS:      false,
				}: {
					"Basic/HTTPVersion:1/Protocol:PROTOCOL_CONNECT/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/TLS:false/basic-unary",
					"Connect GET/HTTPVersion:1/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/TLS:false/connect-get-unary",
					"Connect Version Required (client)/HTTPVersion:1/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/TLS:false/unary-without-connect-version-header",
				},
				{
					protocol:    conformancev1.Protocol_PROTOCOL_CONNECT,
					httpVersion: conformancev1.HTTPVersion_HTTP_VERSION_1,
					useTLS:      true,
				}: {
					"Basic/HTTPVersion:1/Protocol:PROTOCOL_CONNECT/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/TLS:true/basic-unary",
					"Connect GET/HTTPVersion:1/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/TLS:true/connect-get-unary",
					"TLS/HTTPVersion:1/Protocol:PROTOCOL_CONNECT/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/tls-unary",
				},
				{
					protocol:          conformancev1.Protocol_PROTOCOL_CONNECT,
					httpVersion:       conformancev1.HTTPVersion_HTTP_VERSION_1,
					useTLS:            true,
					useTLSClientCerts: true,
				}: {
					"TLS Client Certs/HTTPVersion:1/Protocol:PROTOCOL_CONNECT/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/tls-client-cert-unary",
				},
				{
					protocol:    conformancev1.Protocol_PROTOCOL_GRPC,
					httpVersion: conformancev1.HTTPVersion_HTTP_VERSION_2,
					useTLS:      false,
				}: {
					"Basic/HTTPVersion:2/Protocol:PROTOCOL_GRPC/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/TLS:false/basic-bidi-stream",
				},
			},
		},

		{
			name: "server mode",
			config: []configCase{
				{
					Version:     conformancev1.HTTPVersion_HTTP_VERSION_1,
					Protocol:    conformancev1.Protocol_PROTOCOL_CONNECT,
					Codec:       conformancev1.Codec_CODEC_PROTO,
					Compression: conformancev1.Compression_COMPRESSION_IDENTITY,
					StreamType:  conformancev1.StreamType_STREAM_TYPE_UNARY,
				},
				{
					Version:     conformancev1.HTTPVersion_HTTP_VERSION_1,
					Protocol:    conformancev1.Protocol_PROTOCOL_CONNECT,
					Codec:       conformancev1.Codec_CODEC_PROTO,
					Compression: conformancev1.Compression_COMPRESSION_IDENTITY,
					StreamType:  conformancev1.StreamType_STREAM_TYPE_UNARY,
					UseTLS:      true,
				},
				{
					Version:           conformancev1.HTTPVersion_HTTP_VERSION_1,
					Protocol:          conformancev1.Protocol_PROTOCOL_CONNECT,
					Codec:             conformancev1.Codec_CODEC_PROTO,
					Compression:       conformancev1.Compression_COMPRESSION_IDENTITY,
					StreamType:        conformancev1.StreamType_STREAM_TYPE_UNARY,
					UseTLS:            true,
					UseTLSClientCerts: true,
				},
				{
					Version:       conformancev1.HTTPVersion_HTTP_VERSION_1,
					Protocol:      conformancev1.Protocol_PROTOCOL_CONNECT,
					Codec:         conformancev1.Codec_CODEC_PROTO,
					Compression:   conformancev1.Compression_COMPRESSION_IDENTITY,
					StreamType:    conformancev1.StreamType_STREAM_TYPE_UNARY,
					UseConnectGET: true,
				},
				{
					Version:       conformancev1.HTTPVersion_HTTP_VERSION_1,
					Protocol:      conformancev1.Protocol_PROTOCOL_CONNECT,
					Codec:         conformancev1.Codec_CODEC_PROTO,
					Compression:   conformancev1.Compression_COMPRESSION_IDENTITY,
					StreamType:    conformancev1.StreamType_STREAM_TYPE_UNARY,
					UseConnectGET: true,
					UseTLS:        true,
				},
				{
					Version:            conformancev1.HTTPVersion_HTTP_VERSION_1,
					Protocol:           conformancev1.Protocol_PROTOCOL_CONNECT,
					Codec:              conformancev1.Codec_CODEC_PROTO,
					Compression:        conformancev1.Compression_COMPRESSION_IDENTITY,
					StreamType:         conformancev1.StreamType_STREAM_TYPE_UNARY,
					ConnectVersionMode: conformancev1.TestSuite_CONNECT_VERSION_MODE_IGNORE,
				},
				{
					Version:                conformancev1.HTTPVersion_HTTP_VERSION_1,
					Protocol:               conformancev1.Protocol_PROTOCOL_CONNECT,
					Codec:                  conformancev1.Codec_CODEC_PROTO,
					Compression:            conformancev1.Compression_COMPRESSION_IDENTITY,
					StreamType:             conformancev1.StreamType_STREAM_TYPE_UNARY,
					UseMessageReceiveLimit: true,
				},
				{
					Version:     conformancev1.HTTPVersion_HTTP_VERSION_2,
					Protocol:    conformancev1.Protocol_PROTOCOL_GRPC,
					Codec:       conformancev1.Codec_CODEC_PROTO,
					Compression: conformancev1.Compression_COMPRESSION_IDENTITY,
					StreamType:  conformancev1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
				},
			},
			mode: conformancev1.TestSuite_TEST_MODE_SERVER,
			cases: map[serverInstance][]string{
				{
					protocol:    conformancev1.Protocol_PROTOCOL_CONNECT,
					httpVersion: conformancev1.HTTPVersion_HTTP_VERSION_1,
					useTLS:      false,
				}: {
					"Basic/HTTPVersion:1/Protocol:PROTOCOL_CONNECT/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/TLS:false/basic-unary",
					"Connect GET/HTTPVersion:1/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/TLS:false/connect-get-unary",
					"Connect Version Optional (server)/HTTPVersion:1/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/TLS:false/unary-without-connect-version-header",
					"Max Receive Size (server)/HTTPVersion:1/Protocol:PROTOCOL_CONNECT/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/TLS:false/unary-exceeds-limit",
				},
				{
					protocol:    conformancev1.Protocol_PROTOCOL_CONNECT,
					httpVersion: conformancev1.HTTPVersion_HTTP_VERSION_1,
					useTLS:      true,
				}: {
					"Basic/HTTPVersion:1/Protocol:PROTOCOL_CONNECT/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/TLS:true/basic-unary",
					"Connect GET/HTTPVersion:1/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/TLS:true/connect-get-unary",
					"TLS/HTTPVersion:1/Protocol:PROTOCOL_CONNECT/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/tls-unary",
				},
				{
					protocol:          conformancev1.Protocol_PROTOCOL_CONNECT,
					httpVersion:       conformancev1.HTTPVersion_HTTP_VERSION_1,
					useTLS:            true,
					useTLSClientCerts: true,
				}: {
					"TLS Client Certs/HTTPVersion:1/Protocol:PROTOCOL_CONNECT/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/tls-client-cert-unary",
				},
				{
					protocol:    conformancev1.Protocol_PROTOCOL_GRPC,
					httpVersion: conformancev1.HTTPVersion_HTTP_VERSION_2,
					useTLS:      false,
				}: {
					"Basic/HTTPVersion:2/Protocol:PROTOCOL_GRPC/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/TLS:false/basic-bidi-stream",
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testCaseLib, err := newTestCaseLibrary(testSuites, testCase.config, testCase.mode)
			require.NoError(t, err)
			results := make(map[serverInstance][]string, len(testCaseLib.casesByServer))
			for svrKey, testCaseProtos := range testCaseLib.casesByServer {
				names := make([]string, len(testCaseProtos))
				for i, testCaseProto := range testCaseProtos {
					names[i] = testCaseProto.Request.TestName
				}
				sort.Strings(names)
				results[svrKey] = names
			}
			require.Empty(t, cmp.Diff(testCase.cases, results), "- wanted; + got")
		})
	}
}

func TestParseTestSuites_EmbeddedTestSuites(t *testing.T) {
	t.Parallel()
	testSuiteData, err := testsuites.LoadTestSuites()
	require.NoError(t, err)
	allSuites, err := parseTestSuites(testSuiteData)
	require.NoError(t, err)
	_ = allSuites
	// TODO: basic assertions about the embedded test suites?
}

func TestExpandRequestData(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name         string
		testCaseJSON string
		expectErr    string
		expectSizes  []int
	}{
		{
			name: "unary-no-expand",
			testCaseJSON: `{
				"request": {
					"requestMessages":[
						{
							"@type": "type.googleapis.com/connectrpc.conformance.v1.UnaryRequest",
							"request_data": "abcdefgh"
						}
					]
				},
				"expandRequests":[]
			}`,
			/* negative means size is unchanged */
			expectSizes: []int{-1},
		},
		{
			name: "unary-expand",
			testCaseJSON: `{
				"request": {
					"requestMessages":[
						{
							"@type": "type.googleapis.com/connectrpc.conformance.v1.UnaryRequest",
							"request_data": "abcdefgh"
						}
					]
				},
				"expandRequests":[{"size_relative_to_limit":123}]
			}`,
			expectSizes: []int{200*1024 + 123},
		},
		{
			name: "server-stream-no-expand",
			testCaseJSON: `{
				"request": {
					"requestMessages":[
						{
							"@type": "type.googleapis.com/connectrpc.conformance.v1.ServerStreamRequest",
							"request_data": "abcdefgh"
						}
					]
				},
				"expandRequests":[]
			}`,
			expectSizes: []int{-1},
		},
		{
			name: "server-stream-expand",
			testCaseJSON: `{
				"request": {
					"requestMessages":[
						{
							"@type": "type.googleapis.com/connectrpc.conformance.v1.ServerStreamRequest",
							"request_data": "abcdefgh"
						}
					]
				},
				"expandRequests":[{"size_relative_to_limit":123}]
			}`,
			expectSizes: []int{200*1024 + 123},
		},
		{
			name: "client-stream-no-expand",
			testCaseJSON: `{
				"request": {
					"requestMessages":[
						{
							"@type": "type.googleapis.com/connectrpc.conformance.v1.ClientStreamRequest",
							"request_data": "abcdefgh"
						},
						{
							"@type": "type.googleapis.com/connectrpc.conformance.v1.ClientStreamRequest",
							"request_data": "abcdefgh"
						},
						{
							"@type": "type.googleapis.com/connectrpc.conformance.v1.ClientStreamRequest",
							"request_data": "abcdefgh"
						}
					]
				},
				"expandRequests":[]
			}`,
			expectSizes: []int{-1, -1, -1},
		},
		{
			name: "client-stream-expand",
			testCaseJSON: `{
				"request": {
					"requestMessages":[
						{
							"@type": "type.googleapis.com/connectrpc.conformance.v1.ClientStreamRequest",
							"request_data": "abcdefgh"
						},
						{
							"@type": "type.googleapis.com/connectrpc.conformance.v1.ClientStreamRequest",
							"request_data": "abcdefgh"
						},
						{
							"@type": "type.googleapis.com/connectrpc.conformance.v1.ClientStreamRequest",
							"request_data": "abcdefgh"
						},
						{
							"@type": "type.googleapis.com/connectrpc.conformance.v1.ClientStreamRequest",
							"request_data": "abcdefgh"
						}
					]
				},
				"expandRequests":[
					{"size_relative_to_limit":123},
					{"size_relative_to_limit":null},
					{"size_relative_to_limit":-123}
				]
			}`,
			expectSizes: []int{200*1024 + 123, -1, 200*1024 - 123, -1},
		},
		{
			name: "bidi-stream-no-expand",
			testCaseJSON: `{
				"request": {
					"requestMessages":[
						{
							"@type": "type.googleapis.com/connectrpc.conformance.v1.BidiStreamRequest",
							"request_data": "abcdefgh"
						},
						{
							"@type": "type.googleapis.com/connectrpc.conformance.v1.BidiStreamRequest",
							"request_data": "abcdefgh"
						}
					]
				},
				"expandRequests":[]
			}`,
			expectSizes: []int{-1, -1},
		},
		{
			name: "bidi-stream-expand",
			testCaseJSON: `{
				"request": {
					"requestMessages":[
						{
							"@type": "type.googleapis.com/connectrpc.conformance.v1.BidiStreamRequest",
							"request_data": "abcdefgh"
						},
						{
							"@type": "type.googleapis.com/connectrpc.conformance.v1.BidiStreamRequest",
							"request_data": "abcdefgh"
						},
						{
							"@type": "type.googleapis.com/connectrpc.conformance.v1.BidiStreamRequest",
							"request_data": "abcdefgh"
						}
					]
				},
				"expandRequests":[
					{"size_relative_to_limit":null},
					{"size_relative_to_limit":null},
					{"size_relative_to_limit":0}
				]
			}`,
			expectSizes: []int{-1, -1, 200 * 1024},
		},
		{
			name: "too-many-expand-directives",
			testCaseJSON: `{
				"request": {
					"requestMessages":[]
				},
				"expandRequests":[
					{"size_relative_to_limit":null},
					{"size_relative_to_limit":123}
				]
			}`,
			expectErr: "expand directives indicate 2 messages, but there are only 0 requests",
		},
		{
			name: "invalid-adjustment",
			testCaseJSON: `{
				"request": {
					"requestMessages":[
						{
							"@type": "type.googleapis.com/connectrpc.conformance.v1.BidiStreamRequest",
							"request_data": "abcdefgh"
						}
					]
				},
				"expandRequests":[
					{"size_relative_to_limit":-300000}
				]
			}`,
			expectErr: "expand directive #1 (-300000) results in an invalid request size: -95200",
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var testCaseProto conformancev1.TestCase
			err := protojson.Unmarshal([]byte(testCase.testCaseJSON), &testCaseProto)
			require.NoError(t, err)
			reqs := testCaseProto.Request.RequestMessages
			initialSizes := make([]int, len(reqs))
			for i, req := range reqs {
				initialSizes[i] = len(req.Value)
			}
			err = expandRequestData(&testCaseProto)
			if testCase.expectErr != "" {
				require.ErrorContains(t, err, testCase.expectErr)
				return
			}
			require.NoError(t, err)
			for i, req := range reqs {
				expectedSize := testCase.expectSizes[i]
				if expectedSize < 0 {
					expectedSize = initialSizes[i]
				}
				require.Len(t, req.Value, expectedSize)
			}
		})
	}
}

func TestPopulateExpectedResponse(t *testing.T) {
	t.Parallel()

	requestHeaders := []*conformancev1.Header{
		{
			Name:  "reqHeader",
			Value: []string{"reqHeaderVal"},
		},
	}
	responseHeaders := []*conformancev1.Header{
		{
			Name:  "fooHeader",
			Value: []string{"fooHeaderVal"},
		},
		{
			Name:  "barHeader",
			Value: []string{"barHeaderVal1", "barHeaderVal2"},
		},
	}
	responseTrailers := []*conformancev1.Header{
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

	header := &conformancev1.Header{
		Name:  "error detail test",
		Value: []string{"val1", "val2"},
	}

	unarySuccessReq := &conformancev1.UnaryRequest{
		ResponseDefinition: &conformancev1.UnaryResponseDefinition{
			ResponseHeaders: responseHeaders,
			Response: &conformancev1.UnaryResponseDefinition_ResponseData{
				ResponseData: []byte("data1"),
			},
			ResponseTrailers: responseTrailers,
		},
	}

	unaryErrorReq := &conformancev1.UnaryRequest{
		ResponseDefinition: &conformancev1.UnaryResponseDefinition{
			ResponseHeaders: responseHeaders,
			Response: &conformancev1.UnaryResponseDefinition_Error{
				Error: &conformancev1.Error{
					Code:    int32(connect.CodeResourceExhausted),
					Message: proto.String("message"),
				},
			},
			ResponseTrailers: responseTrailers,
		},
	}

	testCases := []struct {
		testName   string
		request    *conformancev1.ClientCompatRequest
		expected   *conformancev1.ClientResponseResult
		requireErr bool
	}{
		{
			testName: "unary success",
			request: &conformancev1.ClientCompatRequest{
				StreamType:      conformancev1.StreamType_STREAM_TYPE_UNARY,
				RequestMessages: asAnySlice(t, unarySuccessReq),
				RequestHeaders:  requestHeaders,
				TimeoutMs:       proto.Uint32(42),
			},
			expected: &conformancev1.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Payloads: []*conformancev1.ConformancePayload{
					{
						Data: data1,
						RequestInfo: &conformancev1.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests:       asAnySlice(t, unarySuccessReq),
							TimeoutMs:      proto.Int64(42),
						},
					},
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "unary error",
			request: &conformancev1.ClientCompatRequest{
				StreamType:      conformancev1.StreamType_STREAM_TYPE_UNARY,
				RequestMessages: asAnySlice(t, unaryErrorReq),
				RequestHeaders:  requestHeaders,
				TimeoutMs:       proto.Uint32(42),
			},
			expected: &conformancev1.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Error: &conformancev1.Error{
					Code:    int32(connect.CodeResourceExhausted),
					Message: proto.String("message"),
					Details: asAnySlice(t, &conformancev1.ConformancePayload_RequestInfo{
						RequestHeaders: requestHeaders,
						Requests:       asAnySlice(t, unaryErrorReq),
						TimeoutMs:      proto.Int64(42),
					}),
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "unary error specifying details appends req info to details",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_UNARY,
				RequestMessages: asAnySlice(t, &conformancev1.UnaryRequest{
					ResponseDefinition: &conformancev1.UnaryResponseDefinition{
						Response: &conformancev1.UnaryResponseDefinition_Error{
							Error: &conformancev1.Error{
								Code:    int32(connect.CodeResourceExhausted),
								Message: proto.String("message"),
								Details: asAnySlice(t, header),
							},
						},
					},
				}),
			},
			expected: &conformancev1.ClientResponseResult{
				Error: &conformancev1.Error{
					Code:    int32(connect.CodeResourceExhausted),
					Message: proto.String("message"),
					Details: asAnySlice(t, header, &conformancev1.ConformancePayload_RequestInfo{
						Requests: asAnySlice(t, &conformancev1.UnaryRequest{
							ResponseDefinition: &conformancev1.UnaryResponseDefinition{
								Response: &conformancev1.UnaryResponseDefinition_Error{
									Error: &conformancev1.Error{
										Code:    int32(connect.CodeResourceExhausted),
										Message: proto.String("message"),
										Details: asAnySlice(t, header),
									},
								},
							},
						}),
					}),
				},
			},
		},
		{
			testName: "unary no response data specified",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_UNARY,
				RequestMessages: asAnySlice(t, &conformancev1.UnaryRequest{
					ResponseDefinition: &conformancev1.UnaryResponseDefinition{},
				}),
			},
			expected: &conformancev1.ClientResponseResult{
				Payloads: []*conformancev1.ConformancePayload{
					{
						RequestInfo: &conformancev1.ConformancePayload_RequestInfo{
							Requests: asAnySlice(t, &conformancev1.UnaryRequest{
								ResponseDefinition: &conformancev1.UnaryResponseDefinition{},
							}),
						},
					},
				},
			},
		},
		{
			testName: "unary no response definition specified",
			request: &conformancev1.ClientCompatRequest{
				StreamType:      conformancev1.StreamType_STREAM_TYPE_UNARY,
				RequestMessages: asAnySlice(t, &conformancev1.UnaryRequest{}),
			},
			expected: &conformancev1.ClientResponseResult{
				Payloads: []*conformancev1.ConformancePayload{
					{
						RequestInfo: &conformancev1.ConformancePayload_RequestInfo{
							Requests: asAnySlice(t, &conformancev1.UnaryRequest{}),
						},
					},
				},
			},
		},
		{
			testName: "idempotent unary with json success",
			request: &conformancev1.ClientCompatRequest{
				StreamType:       conformancev1.StreamType_STREAM_TYPE_UNARY,
				RequestMessages:  asAnySlice(t, unarySuccessReq),
				RequestHeaders:   requestHeaders,
				UseGetHttpMethod: true,
				Codec:            conformancev1.Codec_CODEC_JSON,
			},
			expected: &conformancev1.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Payloads: []*conformancev1.ConformancePayload{
					{
						Data: data1,
						RequestInfo: &conformancev1.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests:       asAnySlice(t, unarySuccessReq),
							ConnectGetInfo: &conformancev1.ConformancePayload_ConnectGetInfo{
								QueryParams: []*conformancev1.Header{
									{
										Name:  "message",
										Value: []string{marshalToString(t, true, unarySuccessReq)},
									},
									{
										Name:  "encoding",
										Value: []string{"json"},
									},
									{
										Name:  "connect",
										Value: []string{"v1"},
									},
								},
							},
						},
					},
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "idempotent unary with proto success",
			request: &conformancev1.ClientCompatRequest{
				StreamType:       conformancev1.StreamType_STREAM_TYPE_UNARY,
				RequestMessages:  asAnySlice(t, unarySuccessReq),
				RequestHeaders:   requestHeaders,
				UseGetHttpMethod: true,
				Codec:            conformancev1.Codec_CODEC_PROTO,
			},
			expected: &conformancev1.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Payloads: []*conformancev1.ConformancePayload{
					{
						Data: data1,
						RequestInfo: &conformancev1.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests:       asAnySlice(t, unarySuccessReq),
							ConnectGetInfo: &conformancev1.ConformancePayload_ConnectGetInfo{
								QueryParams: []*conformancev1.Header{
									{
										Name:  "message",
										Value: []string{marshalToString(t, false, unarySuccessReq)},
									},
									{
										Name:  "encoding",
										Value: []string{"proto"},
									},
									{
										Name:  "connect",
										Value: []string{"v1"},
									},
								},
							},
						},
					},
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "idempotent unary with json error",
			request: &conformancev1.ClientCompatRequest{
				StreamType:       conformancev1.StreamType_STREAM_TYPE_UNARY,
				RequestMessages:  asAnySlice(t, unaryErrorReq),
				RequestHeaders:   requestHeaders,
				UseGetHttpMethod: true,
				Codec:            conformancev1.Codec_CODEC_JSON,
			},
			expected: &conformancev1.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Error: &conformancev1.Error{
					Code:    int32(connect.CodeResourceExhausted),
					Message: proto.String("message"),
					Details: asAnySlice(t, &conformancev1.ConformancePayload_RequestInfo{
						RequestHeaders: requestHeaders,
						Requests:       asAnySlice(t, unaryErrorReq),
						ConnectGetInfo: &conformancev1.ConformancePayload_ConnectGetInfo{
							QueryParams: []*conformancev1.Header{
								{
									Name:  "message",
									Value: []string{marshalToString(t, true, unaryErrorReq)},
								},
								{
									Name:  "encoding",
									Value: []string{"json"},
								},
								{
									Name:  "connect",
									Value: []string{"v1"},
								},
							},
						},
					}),
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "idempotent unary with proto error",
			request: &conformancev1.ClientCompatRequest{
				StreamType:       conformancev1.StreamType_STREAM_TYPE_UNARY,
				RequestMessages:  asAnySlice(t, unaryErrorReq),
				RequestHeaders:   requestHeaders,
				UseGetHttpMethod: true,
				Codec:            conformancev1.Codec_CODEC_PROTO,
			},
			expected: &conformancev1.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Error: &conformancev1.Error{
					Code:    int32(connect.CodeResourceExhausted),
					Message: proto.String("message"),
					Details: asAnySlice(t, &conformancev1.ConformancePayload_RequestInfo{
						RequestHeaders: requestHeaders,
						Requests:       asAnySlice(t, unaryErrorReq),
						ConnectGetInfo: &conformancev1.ConformancePayload_ConnectGetInfo{
							QueryParams: []*conformancev1.Header{
								{
									Name:  "message",
									Value: []string{marshalToString(t, false, unaryErrorReq)},
								},
								{
									Name:  "encoding",
									Value: []string{"proto"},
								},
								{
									Name:  "connect",
									Value: []string{"v1"},
								},
							},
						},
					}),
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "client stream success",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_CLIENT_STREAM,
				RequestMessages: asAnySlice(t, &conformancev1.ClientStreamRequest{
					ResponseDefinition: &conformancev1.UnaryResponseDefinition{
						ResponseHeaders: responseHeaders,
						Response: &conformancev1.UnaryResponseDefinition_ResponseData{
							ResponseData: data1,
						},
						ResponseTrailers: responseTrailers,
					},
					RequestData: data1,
				}, &conformancev1.ClientStreamRequest{
					RequestData: data1,
				}),
				RequestHeaders: requestHeaders,
			},
			expected: &conformancev1.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Payloads: []*conformancev1.ConformancePayload{
					{
						Data: data1,
						RequestInfo: &conformancev1.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests: asAnySlice(t, &conformancev1.ClientStreamRequest{
								ResponseDefinition: &conformancev1.UnaryResponseDefinition{
									ResponseHeaders: responseHeaders,
									Response: &conformancev1.UnaryResponseDefinition_ResponseData{
										ResponseData: data1,
									},
									ResponseTrailers: responseTrailers,
								},
								RequestData: data1,
							}, &conformancev1.ClientStreamRequest{
								RequestData: data1,
							}),
						},
					},
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "client stream error",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_CLIENT_STREAM,
				RequestMessages: asAnySlice(t, &conformancev1.ClientStreamRequest{
					ResponseDefinition: &conformancev1.UnaryResponseDefinition{
						ResponseHeaders: responseHeaders,
						Response: &conformancev1.UnaryResponseDefinition_Error{
							Error: &conformancev1.Error{
								Code:    int32(connect.CodeResourceExhausted),
								Message: proto.String("message"),
							},
						},
						ResponseTrailers: responseTrailers,
					},
				}, &conformancev1.ClientStreamRequest{
					RequestData: data1,
				}),
				RequestHeaders: requestHeaders,
			},
			expected: &conformancev1.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Error: &conformancev1.Error{
					Code:    int32(connect.CodeResourceExhausted),
					Message: proto.String("message"),
					Details: asAnySlice(t, &conformancev1.ConformancePayload_RequestInfo{
						RequestHeaders: requestHeaders,
						Requests: asAnySlice(t, &conformancev1.ClientStreamRequest{
							ResponseDefinition: &conformancev1.UnaryResponseDefinition{
								ResponseHeaders: responseHeaders,
								Response: &conformancev1.UnaryResponseDefinition_Error{
									Error: &conformancev1.Error{
										Code:    int32(connect.CodeResourceExhausted),
										Message: proto.String("message"),
									},
								},
								ResponseTrailers: responseTrailers,
							},
						}, &conformancev1.ClientStreamRequest{
							RequestData: data1,
						}),
					}),
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "client stream error specifying details appends req info to details",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_CLIENT_STREAM,
				RequestMessages: asAnySlice(t, &conformancev1.ClientStreamRequest{
					ResponseDefinition: &conformancev1.UnaryResponseDefinition{
						Response: &conformancev1.UnaryResponseDefinition_Error{
							Error: &conformancev1.Error{
								Code:    int32(connect.CodeResourceExhausted),
								Message: proto.String("message"),
								Details: asAnySlice(t, header),
							},
						},
					},
				}, &conformancev1.ClientStreamRequest{
					RequestData: data1,
				}),
				TimeoutMs: proto.Uint32(42),
			},
			expected: &conformancev1.ClientResponseResult{
				Error: &conformancev1.Error{
					Code:    int32(connect.CodeResourceExhausted),
					Message: proto.String("message"),
					Details: asAnySlice(t, header, &conformancev1.ConformancePayload_RequestInfo{
						Requests: asAnySlice(t, &conformancev1.ClientStreamRequest{
							ResponseDefinition: &conformancev1.UnaryResponseDefinition{
								Response: &conformancev1.UnaryResponseDefinition_Error{
									Error: &conformancev1.Error{
										Code:    int32(connect.CodeResourceExhausted),
										Message: proto.String("message"),
										Details: asAnySlice(t, header),
									},
								},
							},
						}, &conformancev1.ClientStreamRequest{
							RequestData: data1,
						}),
						TimeoutMs: proto.Int64(42),
					}),
				},
			},
		},
		{
			testName: "client stream no response data specified",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_CLIENT_STREAM,
				RequestMessages: asAnySlice(t, &conformancev1.ClientStreamRequest{
					ResponseDefinition: &conformancev1.UnaryResponseDefinition{},
					RequestData:        data1,
				}, &conformancev1.ClientStreamRequest{
					RequestData: data1,
				}),
			},
			expected: &conformancev1.ClientResponseResult{
				Payloads: []*conformancev1.ConformancePayload{
					{
						RequestInfo: &conformancev1.ConformancePayload_RequestInfo{
							Requests: asAnySlice(t, &conformancev1.ClientStreamRequest{
								ResponseDefinition: &conformancev1.UnaryResponseDefinition{},
								RequestData:        data1,
							}, &conformancev1.ClientStreamRequest{
								RequestData: data1,
							}),
						},
					},
				},
			},
		},
		{
			testName: "client stream no response definition specified",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_CLIENT_STREAM,
				RequestMessages: asAnySlice(t, &conformancev1.ClientStreamRequest{
					RequestData: data1,
				}, &conformancev1.ClientStreamRequest{
					RequestData: data1,
				}),
			},
			expected: &conformancev1.ClientResponseResult{
				Payloads: []*conformancev1.ConformancePayload{
					{
						RequestInfo: &conformancev1.ConformancePayload_RequestInfo{
							Requests: asAnySlice(t, &conformancev1.ClientStreamRequest{
								RequestData: data1,
							}, &conformancev1.ClientStreamRequest{
								RequestData: data1,
							}),
						},
					},
				},
			},
		},
		{
			testName: "server stream success",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_SERVER_STREAM,
				RequestMessages: asAnySlice(t, &conformancev1.ServerStreamRequest{
					ResponseDefinition: &conformancev1.StreamResponseDefinition{
						ResponseHeaders:  responseHeaders,
						ResponseData:     [][]byte{data1, data2},
						ResponseDelayMs:  1000,
						ResponseTrailers: responseTrailers,
					},
				}),
				RequestHeaders: requestHeaders,
			},
			expected: &conformancev1.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Payloads: []*conformancev1.ConformancePayload{
					{
						Data: data1,
						RequestInfo: &conformancev1.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests: asAnySlice(t, &conformancev1.ServerStreamRequest{
								ResponseDefinition: &conformancev1.StreamResponseDefinition{
									ResponseHeaders:  responseHeaders,
									ResponseData:     [][]byte{data1, data2},
									ResponseDelayMs:  1000,
									ResponseTrailers: responseTrailers,
								},
							}),
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
			testName: "server stream error with responses",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_SERVER_STREAM,
				RequestMessages: asAnySlice(t, &conformancev1.ServerStreamRequest{
					ResponseDefinition: &conformancev1.StreamResponseDefinition{
						ResponseHeaders: responseHeaders,
						ResponseData:    [][]byte{data1, data2},
						ResponseDelayMs: 1000,
						Error: &conformancev1.Error{
							Code:    int32(connect.CodeResourceExhausted),
							Message: proto.String("message"),
						},
						ResponseTrailers: responseTrailers,
					},
				}),
				RequestHeaders: requestHeaders,
			},
			expected: &conformancev1.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Payloads: []*conformancev1.ConformancePayload{
					{
						Data: data1,
						RequestInfo: &conformancev1.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests: asAnySlice(t, &conformancev1.ServerStreamRequest{
								ResponseDefinition: &conformancev1.StreamResponseDefinition{
									ResponseHeaders: responseHeaders,
									ResponseData:    [][]byte{data1, data2},
									ResponseDelayMs: 1000,
									Error: &conformancev1.Error{
										Code:    int32(connect.CodeResourceExhausted),
										Message: proto.String("message"),
									},
									ResponseTrailers: responseTrailers,
								},
							}),
						},
					},
					{
						Data: data2,
					},
				},
				Error: &conformancev1.Error{
					Code:    int32(connect.CodeResourceExhausted),
					Message: proto.String("message"),
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "server stream error with no responses",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_SERVER_STREAM,
				RequestMessages: asAnySlice(t, &conformancev1.ServerStreamRequest{
					ResponseDefinition: &conformancev1.StreamResponseDefinition{
						ResponseHeaders: responseHeaders,
						ResponseDelayMs: 1000,
						Error: &conformancev1.Error{
							Code:    int32(connect.CodeResourceExhausted),
							Message: proto.String("message"),
						},
						ResponseTrailers: responseTrailers,
					},
				}),
				RequestHeaders: requestHeaders,
				TimeoutMs:      proto.Uint32(42),
			},
			expected: &conformancev1.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Error: &conformancev1.Error{
					Code:    int32(connect.CodeResourceExhausted),
					Message: proto.String("message"),
					Details: asAnySlice(t, &conformancev1.ConformancePayload_RequestInfo{
						RequestHeaders: requestHeaders,
						Requests: asAnySlice(t, &conformancev1.ServerStreamRequest{
							ResponseDefinition: &conformancev1.StreamResponseDefinition{
								ResponseHeaders: responseHeaders,
								ResponseDelayMs: 1000,
								Error: &conformancev1.Error{
									Code:    int32(connect.CodeResourceExhausted),
									Message: proto.String("message"),
								},
								ResponseTrailers: responseTrailers,
							},
						}),
						TimeoutMs: proto.Int64(42),
					}),
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "server stream no response data specified",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_SERVER_STREAM,
				RequestMessages: asAnySlice(t, &conformancev1.ServerStreamRequest{
					ResponseDefinition: &conformancev1.StreamResponseDefinition{
						ResponseHeaders:  responseHeaders,
						ResponseDelayMs:  1000,
						ResponseTrailers: responseTrailers,
					},
				}),
			},
			expected: &conformancev1.ClientResponseResult{
				ResponseHeaders:  responseHeaders,
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "server stream no response definition specified",
			request: &conformancev1.ClientCompatRequest{
				StreamType:      conformancev1.StreamType_STREAM_TYPE_SERVER_STREAM,
				RequestMessages: asAnySlice(t, &conformancev1.ServerStreamRequest{}),
				RequestHeaders:  requestHeaders,
			},
			expected: &conformancev1.ClientResponseResult{},
		},
		{
			testName: "half duplex bidi stream success",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM,
				RequestMessages: asAnySlice(t, &conformancev1.BidiStreamRequest{
					ResponseDefinition: &conformancev1.StreamResponseDefinition{
						ResponseHeaders:  responseHeaders,
						ResponseData:     [][]byte{data1, data2},
						ResponseDelayMs:  1000,
						ResponseTrailers: responseTrailers,
					},
					RequestData: data1,
				}, &conformancev1.BidiStreamRequest{
					RequestData: data2,
				}),
				RequestHeaders: requestHeaders,
			},
			expected: &conformancev1.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Payloads: []*conformancev1.ConformancePayload{
					{
						Data: data1,
						RequestInfo: &conformancev1.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests: asAnySlice(t, &conformancev1.BidiStreamRequest{
								ResponseDefinition: &conformancev1.StreamResponseDefinition{
									ResponseHeaders:  responseHeaders,
									ResponseData:     [][]byte{data1, data2},
									ResponseDelayMs:  1000,
									ResponseTrailers: responseTrailers,
								},
								RequestData: data1,
							}, &conformancev1.BidiStreamRequest{
								RequestData: data2,
							}),
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
			testName: "half duplex bidi stream error with responses",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM,
				RequestMessages: asAnySlice(t, &conformancev1.BidiStreamRequest{
					ResponseDefinition: &conformancev1.StreamResponseDefinition{
						ResponseHeaders: responseHeaders,
						ResponseData:    [][]byte{data1, data2},
						ResponseDelayMs: 1000,
						Error: &conformancev1.Error{
							Code:    int32(connect.CodeResourceExhausted),
							Message: proto.String("message"),
						},
						ResponseTrailers: responseTrailers,
					},
					RequestData: data1,
				}, &conformancev1.BidiStreamRequest{
					RequestData: data2,
				}),
				RequestHeaders: requestHeaders,
			},
			expected: &conformancev1.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Payloads: []*conformancev1.ConformancePayload{
					{
						Data: data1,
						RequestInfo: &conformancev1.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests: asAnySlice(t, &conformancev1.BidiStreamRequest{
								ResponseDefinition: &conformancev1.StreamResponseDefinition{
									ResponseHeaders: responseHeaders,
									ResponseData:    [][]byte{data1, data2},
									ResponseDelayMs: 1000,
									Error: &conformancev1.Error{
										Code:    int32(connect.CodeResourceExhausted),
										Message: proto.String("message"),
									},
									ResponseTrailers: responseTrailers,
								},
								RequestData: data1,
							}, &conformancev1.BidiStreamRequest{
								RequestData: data2,
							}),
						},
					},
					{
						Data: data2,
					},
				},
				Error: &conformancev1.Error{
					Code:    int32(connect.CodeResourceExhausted),
					Message: proto.String("message"),
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "half duplex bidi stream error with no responses",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM,
				RequestMessages: asAnySlice(t, &conformancev1.BidiStreamRequest{
					ResponseDefinition: &conformancev1.StreamResponseDefinition{
						ResponseHeaders: responseHeaders,
						ResponseDelayMs: 1000,
						Error: &conformancev1.Error{
							Code:    int32(connect.CodeResourceExhausted),
							Message: proto.String("message"),
						},
						ResponseTrailers: responseTrailers,
					},
					RequestData: data1,
				}, &conformancev1.BidiStreamRequest{
					RequestData: data2,
				}),
				RequestHeaders: requestHeaders,
				TimeoutMs:      proto.Uint32(42),
			},
			expected: &conformancev1.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Error: &conformancev1.Error{
					Code:    int32(connect.CodeResourceExhausted),
					Message: proto.String("message"),
					Details: asAnySlice(t, &conformancev1.ConformancePayload_RequestInfo{
						RequestHeaders: requestHeaders,
						Requests: asAnySlice(t, &conformancev1.BidiStreamRequest{
							ResponseDefinition: &conformancev1.StreamResponseDefinition{
								ResponseHeaders: responseHeaders,
								ResponseDelayMs: 1000,
								Error: &conformancev1.Error{
									Code:    int32(connect.CodeResourceExhausted),
									Message: proto.String("message"),
								},
								ResponseTrailers: responseTrailers,
							},
							RequestData: data1,
						}, &conformancev1.BidiStreamRequest{
							RequestData: data2,
						}),
						TimeoutMs: proto.Int64(42),
					}),
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "half duplex bidi stream no response data specified",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM,
				RequestMessages: asAnySlice(t, &conformancev1.BidiStreamRequest{
					ResponseDefinition: &conformancev1.StreamResponseDefinition{
						ResponseHeaders:  responseHeaders,
						ResponseDelayMs:  1000,
						ResponseTrailers: responseTrailers,
					},
					RequestData: data1,
				}),
			},
			expected: &conformancev1.ClientResponseResult{
				ResponseHeaders:  responseHeaders,
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "half duplex bidi stream no response definition specified",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM,
				RequestMessages: asAnySlice(t, &conformancev1.BidiStreamRequest{
					RequestData: data1,
				}),
			},
			expected: &conformancev1.ClientResponseResult{},
		},
		{
			testName: "full duplex bidi stream success",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
				RequestMessages: asAnySlice(t, &conformancev1.BidiStreamRequest{
					ResponseDefinition: &conformancev1.StreamResponseDefinition{
						ResponseHeaders:  responseHeaders,
						ResponseData:     [][]byte{data1, data2},
						ResponseDelayMs:  1000,
						ResponseTrailers: responseTrailers,
					},
					RequestData: data1,
					FullDuplex:  true,
				}, &conformancev1.BidiStreamRequest{
					RequestData: data2,
				}),
				RequestHeaders: requestHeaders,
			},
			expected: &conformancev1.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Payloads: []*conformancev1.ConformancePayload{
					{
						Data: data1,
						RequestInfo: &conformancev1.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests: asAnySlice(t, &conformancev1.BidiStreamRequest{
								ResponseDefinition: &conformancev1.StreamResponseDefinition{
									ResponseHeaders:  responseHeaders,
									ResponseData:     [][]byte{data1, data2},
									ResponseDelayMs:  1000,
									ResponseTrailers: responseTrailers,
								},
								RequestData: data1,
								FullDuplex:  true,
							}),
						},
					},
					{
						Data: data2,
						RequestInfo: &conformancev1.ConformancePayload_RequestInfo{
							Requests: asAnySlice(t, &conformancev1.BidiStreamRequest{
								RequestData: data2,
							}),
						},
					},
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "full duplex bidi stream error with responses",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
				RequestMessages: asAnySlice(t, &conformancev1.BidiStreamRequest{
					ResponseDefinition: &conformancev1.StreamResponseDefinition{
						ResponseHeaders: responseHeaders,
						ResponseData:    [][]byte{data1, data2},
						ResponseDelayMs: 1000,
						Error: &conformancev1.Error{
							Code:    int32(connect.CodeResourceExhausted),
							Message: proto.String("message"),
						},
						ResponseTrailers: responseTrailers,
					},
					RequestData: data1,
					FullDuplex:  true,
				}, &conformancev1.BidiStreamRequest{
					RequestData: data2,
				}),
				RequestHeaders: requestHeaders,
			},
			expected: &conformancev1.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Payloads: []*conformancev1.ConformancePayload{
					{
						Data: data1,
						RequestInfo: &conformancev1.ConformancePayload_RequestInfo{
							RequestHeaders: requestHeaders,
							Requests: asAnySlice(t, &conformancev1.BidiStreamRequest{
								ResponseDefinition: &conformancev1.StreamResponseDefinition{
									ResponseHeaders: responseHeaders,
									ResponseData:    [][]byte{data1, data2},
									ResponseDelayMs: 1000,
									Error: &conformancev1.Error{
										Code:    int32(connect.CodeResourceExhausted),
										Message: proto.String("message"),
									},
									ResponseTrailers: responseTrailers,
								},
								RequestData: data1,
								FullDuplex:  true,
							}),
						},
					},
					{
						Data: data2,
						RequestInfo: &conformancev1.ConformancePayload_RequestInfo{
							Requests: asAnySlice(t, &conformancev1.BidiStreamRequest{
								RequestData: data2,
							}),
						},
					},
				},
				Error: &conformancev1.Error{
					Code:    int32(connect.CodeResourceExhausted),
					Message: proto.String("message"),
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "full duplex bidi stream error with no responses",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
				RequestMessages: asAnySlice(t, &conformancev1.BidiStreamRequest{
					ResponseDefinition: &conformancev1.StreamResponseDefinition{
						ResponseHeaders: responseHeaders,
						ResponseDelayMs: 1000,
						Error: &conformancev1.Error{
							Code:    int32(connect.CodeResourceExhausted),
							Message: proto.String("message"),
						},
						ResponseTrailers: responseTrailers,
					},
					RequestData: data1,
					FullDuplex:  true,
				}, &conformancev1.BidiStreamRequest{
					RequestData: data2,
				}),
			},
			expected: &conformancev1.ClientResponseResult{
				ResponseHeaders: responseHeaders,
				Error: &conformancev1.Error{
					Code:    int32(connect.CodeResourceExhausted),
					Message: proto.String("message"),
					Details: asAnySlice(t, &conformancev1.ConformancePayload_RequestInfo{
						Requests: asAnySlice(t, &conformancev1.BidiStreamRequest{
							ResponseDefinition: &conformancev1.StreamResponseDefinition{
								ResponseHeaders: responseHeaders,
								ResponseDelayMs: 1000,
								Error: &conformancev1.Error{
									Code:    int32(connect.CodeResourceExhausted),
									Message: proto.String("message"),
								},
								ResponseTrailers: responseTrailers,
							},
							RequestData: data1,
							FullDuplex:  true,
						}, &conformancev1.BidiStreamRequest{
							RequestData: data2,
						}),
					}),
				},
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "full duplex bidi stream no response data specified",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
				RequestMessages: asAnySlice(t, &conformancev1.BidiStreamRequest{
					ResponseDefinition: &conformancev1.StreamResponseDefinition{
						ResponseHeaders:  responseHeaders,
						ResponseDelayMs:  1000,
						ResponseTrailers: responseTrailers,
					},
					RequestData: data1,
					FullDuplex:  true,
				}),
			},
			expected: &conformancev1.ClientResponseResult{
				ResponseHeaders:  responseHeaders,
				ResponseTrailers: responseTrailers,
			},
		},
		{
			testName: "full duplex bidi stream no response definition specified",
			request: &conformancev1.ClientCompatRequest{
				StreamType: conformancev1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
				RequestMessages: asAnySlice(t, &conformancev1.BidiStreamRequest{
					RequestData: data1,
					FullDuplex:  true,
				}),
			},
			expected: &conformancev1.ClientResponseResult{},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			tc := &conformancev1.TestCase{ //nolint:varnamelen
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

// asAnySlice converts the given variadic arg of proto messages to a slice of Any protos
// and verifies there are no errors during the conversion
func asAnySlice(t *testing.T, msgs ...proto.Message) []*anypb.Any {
	arr := make([]*anypb.Any, 0, len(msgs))
	for _, msg := range msgs {
		asAny, err := anypb.New(msg)
		require.NoError(t, err)
		arr = append(arr, asAny)
	}
	return arr
}

// marshalToString marshals the given proto message to a string mirroring the
// logic that Connect specifies for GET requests.
// If asJSON is true, the message is first marshalled to JSON and the bytes are
// then converted to a string.
// If asJSON is false, the message is marshalled to binary and the bytes are then
// base64-encoded as a string.
func marshalToString(t *testing.T, asJSON bool, msg proto.Message) string {
	codec := internal.NewCodec(asJSON)

	bytes, err := codec.MarshalStable(msg)
	require.NoError(t, err)

	if asJSON {
		return string(bytes)
	}

	return base64.RawURLEncoding.EncodeToString(bytes)
}
