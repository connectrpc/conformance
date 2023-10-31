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
	"sort"
	"testing"

	"connectrpc.com/conformance/internal/app/connectconformance/testsuites"
	conformancev1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestLoadTestCases(t *testing.T) {
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
		mode   conformancev1alpha1.TestSuite_TestMode
		cases  map[serverInstance][]string
	}{
		{
			name: "client mode",
			config: []configCase{
				{
					Version:     conformancev1alpha1.HTTPVersion_HTTP_VERSION_1,
					Protocol:    conformancev1alpha1.Protocol_PROTOCOL_CONNECT,
					Codec:       conformancev1alpha1.Codec_CODEC_PROTO,
					Compression: conformancev1alpha1.Compression_COMPRESSION_IDENTITY,
					StreamType:  conformancev1alpha1.StreamType_STREAM_TYPE_UNARY,
				},
				{
					Version:     conformancev1alpha1.HTTPVersion_HTTP_VERSION_1,
					Protocol:    conformancev1alpha1.Protocol_PROTOCOL_CONNECT,
					Codec:       conformancev1alpha1.Codec_CODEC_PROTO,
					Compression: conformancev1alpha1.Compression_COMPRESSION_IDENTITY,
					StreamType:  conformancev1alpha1.StreamType_STREAM_TYPE_UNARY,
					UseTLS:      true,
				},
				{
					Version:       conformancev1alpha1.HTTPVersion_HTTP_VERSION_1,
					Protocol:      conformancev1alpha1.Protocol_PROTOCOL_CONNECT,
					Codec:         conformancev1alpha1.Codec_CODEC_PROTO,
					Compression:   conformancev1alpha1.Compression_COMPRESSION_IDENTITY,
					StreamType:    conformancev1alpha1.StreamType_STREAM_TYPE_UNARY,
					UseConnectGET: true,
				},
				{
					Version:            conformancev1alpha1.HTTPVersion_HTTP_VERSION_1,
					Protocol:           conformancev1alpha1.Protocol_PROTOCOL_CONNECT,
					Codec:              conformancev1alpha1.Codec_CODEC_PROTO,
					Compression:        conformancev1alpha1.Compression_COMPRESSION_IDENTITY,
					StreamType:         conformancev1alpha1.StreamType_STREAM_TYPE_UNARY,
					ConnectVersionMode: conformancev1alpha1.TestSuite_CONNECT_VERSION_MODE_REQUIRE,
				},
				{
					Version:     conformancev1alpha1.HTTPVersion_HTTP_VERSION_2,
					Protocol:    conformancev1alpha1.Protocol_PROTOCOL_GRPC,
					Codec:       conformancev1alpha1.Codec_CODEC_PROTO,
					Compression: conformancev1alpha1.Compression_COMPRESSION_IDENTITY,
					StreamType:  conformancev1alpha1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
				},
			},
			mode: conformancev1alpha1.TestSuite_TEST_MODE_CLIENT,
			cases: map[serverInstance][]string{
				{
					protocol:    conformancev1alpha1.Protocol_PROTOCOL_CONNECT,
					httpVersion: conformancev1alpha1.HTTPVersion_HTTP_VERSION_1,
					useTLS:      false,
				}: {
					"Basic/HTTPVersion:1/Protocol:PROTOCOL_CONNECT/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/basic-unary",
					"Connect GET/HTTPVersion:1/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/connect-get-unary",
					"Connect Version Required (client)/HTTPVersion:1/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/unary-without-connect-version-header",
				},
				{
					protocol:    conformancev1alpha1.Protocol_PROTOCOL_CONNECT,
					httpVersion: conformancev1alpha1.HTTPVersion_HTTP_VERSION_1,
					useTLS:      true,
				}: {
					"TLS/HTTPVersion:1/Protocol:PROTOCOL_CONNECT/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/tls-unary",
				},
				{
					protocol:    conformancev1alpha1.Protocol_PROTOCOL_GRPC,
					httpVersion: conformancev1alpha1.HTTPVersion_HTTP_VERSION_2,
					useTLS:      false,
				}: {
					"Basic/HTTPVersion:2/Protocol:PROTOCOL_GRPC/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/basic-bidi-stream",
				},
			},
		},

		{
			name: "server mode",
			config: []configCase{
				{
					Version:     conformancev1alpha1.HTTPVersion_HTTP_VERSION_1,
					Protocol:    conformancev1alpha1.Protocol_PROTOCOL_CONNECT,
					Codec:       conformancev1alpha1.Codec_CODEC_PROTO,
					Compression: conformancev1alpha1.Compression_COMPRESSION_IDENTITY,
					StreamType:  conformancev1alpha1.StreamType_STREAM_TYPE_UNARY,
				},
				{
					Version:     conformancev1alpha1.HTTPVersion_HTTP_VERSION_1,
					Protocol:    conformancev1alpha1.Protocol_PROTOCOL_CONNECT,
					Codec:       conformancev1alpha1.Codec_CODEC_PROTO,
					Compression: conformancev1alpha1.Compression_COMPRESSION_IDENTITY,
					StreamType:  conformancev1alpha1.StreamType_STREAM_TYPE_UNARY,
					UseTLS:      true,
				},
				{
					Version:       conformancev1alpha1.HTTPVersion_HTTP_VERSION_1,
					Protocol:      conformancev1alpha1.Protocol_PROTOCOL_CONNECT,
					Codec:         conformancev1alpha1.Codec_CODEC_PROTO,
					Compression:   conformancev1alpha1.Compression_COMPRESSION_IDENTITY,
					StreamType:    conformancev1alpha1.StreamType_STREAM_TYPE_UNARY,
					UseConnectGET: true,
				},
				{
					Version:            conformancev1alpha1.HTTPVersion_HTTP_VERSION_1,
					Protocol:           conformancev1alpha1.Protocol_PROTOCOL_CONNECT,
					Codec:              conformancev1alpha1.Codec_CODEC_PROTO,
					Compression:        conformancev1alpha1.Compression_COMPRESSION_IDENTITY,
					StreamType:         conformancev1alpha1.StreamType_STREAM_TYPE_UNARY,
					ConnectVersionMode: conformancev1alpha1.TestSuite_CONNECT_VERSION_MODE_IGNORE,
				},
				{
					Version:     conformancev1alpha1.HTTPVersion_HTTP_VERSION_2,
					Protocol:    conformancev1alpha1.Protocol_PROTOCOL_GRPC,
					Codec:       conformancev1alpha1.Codec_CODEC_PROTO,
					Compression: conformancev1alpha1.Compression_COMPRESSION_IDENTITY,
					StreamType:  conformancev1alpha1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
				},
			},
			mode: conformancev1alpha1.TestSuite_TEST_MODE_SERVER,
			cases: map[serverInstance][]string{
				{
					protocol:    conformancev1alpha1.Protocol_PROTOCOL_CONNECT,
					httpVersion: conformancev1alpha1.HTTPVersion_HTTP_VERSION_1,
					useTLS:      false,
				}: {
					"Basic/HTTPVersion:1/Protocol:PROTOCOL_CONNECT/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/basic-unary",
					"Connect GET/HTTPVersion:1/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/connect-get-unary",
					"Connect Version Optional (server)/HTTPVersion:1/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/unary-without-connect-version-header",
				},
				{
					protocol:    conformancev1alpha1.Protocol_PROTOCOL_CONNECT,
					httpVersion: conformancev1alpha1.HTTPVersion_HTTP_VERSION_1,
					useTLS:      true,
				}: {
					"TLS/HTTPVersion:1/Protocol:PROTOCOL_CONNECT/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/tls-unary",
				},
				{
					protocol:    conformancev1alpha1.Protocol_PROTOCOL_GRPC,
					httpVersion: conformancev1alpha1.HTTPVersion_HTTP_VERSION_2,
					useTLS:      false,
				}: {
					"Basic/HTTPVersion:2/Protocol:PROTOCOL_GRPC/Codec:CODEC_PROTO/Compression:COMPRESSION_IDENTITY/basic-bidi-stream",
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testCaseLib, err := loadTestCases(testSuites, testCase.config, testCase.mode)
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
	// TODO: basic assertions about the embedded test suites
}
