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

package connectconformance

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"

	"connectrpc.com/conformance/internal/app/connectconformance/testsuites"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
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
                            service: connectrpc.conformance.v1.ConformanceService
                            method: Unary
                            streamType: STREAM_TYPE_UNARY
                      - request:
                            testName: basic-client-stream
                            service: connectrpc.conformance.v1.ConformanceService
                            method: ClientStream
                            streamType: STREAM_TYPE_CLIENT_STREAM
                      - request:
                            testName: basic-server-stream
                            service: connectrpc.conformance.v1.ConformanceService
                            method: ServerStream
                            streamType: STREAM_TYPE_SERVER_STREAM
                      - request:
                            testName: basic-bidi-stream
                            service: connectrpc.conformance.v1.ConformanceService
                            method: BidStream
                            streamType: STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM`,
		"tls.yaml": `
                    name: TLS
                    reliesOnTls: true
                    testCases:
                      - request:
                            testName: tls-unary
                            service: connectrpc.conformance.v1.ConformanceService
                            method: Unary
                            streamType: STREAM_TYPE_UNARY
                      - request:
                            testName: tls-client-stream
                            service: connectrpc.conformance.v1.ConformanceService
                            method: ClientStream
                            streamType: STREAM_TYPE_CLIENT_STREAM
                      - request:
                            testName: tls-server-stream
                            service: connectrpc.conformance.v1.ConformanceService
                            method: ServerStream
                            streamType: STREAM_TYPE_SERVER_STREAM
                      - request:
                            testName: tls-bidi-stream
                            service: connectrpc.conformance.v1.ConformanceService
                            method: BidiStream
                            streamType: STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM`,
		"tls-client-certs.yaml": `
                    name: TLS Client Certs
                    reliesOnTls: true
                    reliesOnTlsClientCerts: true
                    testCases:
                      - request:
                            testName: tls-client-cert-unary
                            service: connectrpc.conformance.v1.ConformanceService
                            method: Unary
                            streamType: STREAM_TYPE_UNARY`,
		"connect-get.yaml": `
                    name: Connect GET
                    relevantProtocols: [PROTOCOL_CONNECT]
                    reliesOnConnectGet: true
                    testCases:
                      - request:
                            testName: connect-get-unary
                            service: connectrpc.conformance.v1.ConformanceService
                            method: Unary
                            streamType: STREAM_TYPE_UNARY`,
		"connect-version-client-required.yaml": `
                    name: Connect Version Required (client)
                    mode: TEST_MODE_CLIENT
                    relevantProtocols: [PROTOCOL_CONNECT]
                    connectVersionMode: CONNECT_VERSION_MODE_REQUIRE
                    testCases:
                      - request:
                            testName: unary-without-connect-version-header
                            service: connectrpc.conformance.v1.ConformanceService
                            method: Unary
                            streamType: STREAM_TYPE_UNARY`,
		"connect-version-server-required.yaml": `
                    name: Connect Version Required (server)
                    mode: TEST_MODE_SERVER
                    relevantProtocols: [PROTOCOL_CONNECT]
                    connectVersionMode: CONNECT_VERSION_MODE_REQUIRE
                    testCases:
                      - request:
                            testName: unary-without-connect-version-header
                            service: connectrpc.conformance.v1.ConformanceService
                            method: Unary
                            streamType: STREAM_TYPE_UNARY`,
		"connect-version-client-not-required.yaml": `
                    name: Connect Version Optional (client)
                    mode: TEST_MODE_CLIENT
                    relevantProtocols: [PROTOCOL_CONNECT]
                    connectVersionMode: CONNECT_VERSION_MODE_IGNORE
                    testCases:
                      - request:
                            testName: unary-without-connect-version-header
                            service: connectrpc.conformance.v1.ConformanceService
                            method: Unary
                            streamType: STREAM_TYPE_UNARY`,
		"connect-version-server-not-required.yaml": `
                    name: Connect Version Optional (server)
                    mode: TEST_MODE_SERVER
                    relevantProtocols: [PROTOCOL_CONNECT]
                    connectVersionMode: CONNECT_VERSION_MODE_IGNORE
                    testCases:
                      - request:
                            testName: unary-without-connect-version-header
                            service: connectrpc.conformance.v1.ConformanceService
                            method: Unary
                            streamType: STREAM_TYPE_UNARY`,
		"max-receive-limit": `
                    name: Max Receive Size (server)
                    mode: TEST_MODE_SERVER
                    reliesOnMessageReceiveLimit: true
                    testCases:
                      - request:
                            testName: unary-exceeds-limit
                            service: connectrpc.conformance.v1.ConformanceService
                            method: Unary
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

	// Validate naming conventions:
	//   File names are "lower_snake_case" with ".yaml" extension
	//   Suite names are "Title Case" case without punctuation
	//   Test case names are paths (with "/") in "kebab-case"
	for filename, testSuite := range allSuites {
		filename = filepath.Base(filename)
		ext := filepath.Ext(filename)
		assert.Equal(t, ".yaml", ext, "filename %q must have '.yaml' extension", filename)
		if ext != "" {
			filename = strings.TrimSuffix(filename, ext)
		}
		assert.False(t, strings.ContainsFunc(filename, func(r rune) bool {
			const allowed = "abcdefghijklmnopqrstuvwxyz0123456789_"
			return !strings.ContainsRune(allowed, r)
		}), "filename %q may only have lower-case letters, numbers, and underscores", filename)

		testSuiteName := strings.TrimSpace(testSuite.Name)
		if assert.NotEmpty(t, testSuiteName, "test suite name for file %q is blank", filename) {
			// Title case means starting with capital letter, but we allow
			// lower-case "g" if first word is "gRPC".
			assert.True(t, strings.HasPrefix(testSuiteName, "gRPC") || unicode.IsUpper(rune(testSuiteName[0])),
				"test suite name %q should start with capital letter", testSuiteName)
		}
		assert.Equal(t, testSuite.Name, testSuiteName,
			"test suite name %q should not have leading or trailing spaces", testSuiteName)
		assert.False(t, strings.ContainsFunc(testSuiteName, func(r rune) bool {
			const allowed = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789- "
			return !strings.ContainsRune(allowed, r)
		}), "test suite name %q may only have letters, numbers, hyphens, and spaces", testSuiteName)

		for _, testCase := range testSuite.TestCases {
			testCaseName := testCase.Request.TestName
			assert.False(t, strings.ContainsFunc(testCaseName, func(r rune) bool {
				const allowed = "abcdefghijklmnopqrstuvwxyz0123456789-/"
				return !strings.ContainsRune(allowed, r)
			}), "test case name %q (in test suite %q) may only have lower-case letters, numbers, hyphens, and slashes", testCaseName, testSuiteName)
		}
	}

	require.NoError(t, err)
	configCases, err := parseConfig("config.yaml", nil)
	require.NoError(t, err)
	lib, err := newTestCaseLibrary(allSuites, configCases, conformancev1.TestSuite_TEST_MODE_CLIENT)
	require.NoError(t, err)
	// Count assertions are sanity checks based on the number of test cases we have
	// (as of Mar 1st, 2024).
	require.GreaterOrEqual(t, len(lib.testCases), 4000)

	// Some basic validation of the internal structures.
	require.Len(t, lib.testCaseNames, len(lib.testCases))
	for fullName, testCase := range lib.testCases {
		// check some things in the name
		require.Equal(t, fullName, testCase.Request.TestName)
		baseName := lib.testCaseNames[fullName]
		require.NotEmpty(t, baseName)
		require.True(t, strings.HasSuffix(fullName, baseName))
		prefix := strings.TrimSuffix(fullName, baseName)
		require.True(t, strings.HasSuffix(prefix, "/"))
	}
	var totalCases int
	for _, testCases := range lib.casesByServer {
		totalCases += len(testCases)
	}
	require.Len(t, lib.testCases, totalCases)

	// Compute permutations and do some more basic checks.
	allTestCases := lib.allPermutations(false, true)
	require.GreaterOrEqual(t, len(allTestCases), 4500)
	var grpcMarkers int
	for _, testCase := range allTestCases {
		fullName := testCase.Request.TestName
		require.NotContains(t, fullName, grpcClientImplMarker)
		require.NotContains(t, fullName, grpcImplMarker)
		if strings.Contains(fullName, grpcServerImplMarker) {
			grpcMarkers++
			// more name checks
			origName := strings.Replace(strings.Replace(fullName, grpcServerImplMarker, "", 1), "//", "/", 1)
			baseName := lib.testCaseNames[origName]
			require.NotEmpty(t, baseName)
			require.True(t, strings.HasSuffix(fullName, baseName))
			prefix := strings.TrimSuffix(fullName, baseName)
			require.True(t, strings.HasSuffix(prefix, "/"))
		}
	}
	require.GreaterOrEqual(t, grpcMarkers, 500)
}

func TestAddGRPCMarkerToName(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		full, base                 string
		clientIsGRPC, serverIsGRPC bool
		expected                   string
	}{
		{
			full:         "success",
			base:         "success",
			clientIsGRPC: true,
			expected:     "(grpc client impl)/success",
		},
		{
			full:         "ABC/success",
			base:         "success",
			clientIsGRPC: true,
			expected:     "ABC/(grpc client impl)/success",
		},
		{
			full:         "ABC/X=123/B:C/Foo:bar/unary/success",
			base:         "unary/success",
			clientIsGRPC: true,
			expected:     "ABC/X=123/B:C/Foo:bar/(grpc client impl)/unary/success",
		},
	}
	for _, testCase := range testCases {
		markedName := addGRPCMarkerToName(testCase.full, testCase.base, testCase.clientIsGRPC, testCase.serverIsGRPC)
		assert.Equal(t, testCase.expected, markedName, "%q (client=%v, server=%v)", testCase.full, testCase.clientIsGRPC, testCase.serverIsGRPC)
	}
}

func TestFilter(t *testing.T) {
	t.Parallel()
	allTestCaseNames := []string{
		"Basic/foo/bar=baz/unary",
		"Basic/foo/bar=baz/client stream",
		"Basic/foo/bar=baz/server stream",
		"Basic/foo/bar=baz/bidi stream",
		"Basic/foo/bar=baz/(frobnitz)/unary",
		"Basic/foo/bar=baz/(frobnitz)/client stream",
		"Basic/foo/bar=baz/(frobnitz)/server stream",
		"Basic/foo/bar=baz/(frobnitz)/bidi stream",
		"Cancel/foo/bar=baz/unary",
		"Cancel/foo/bar=baz/client stream",
		"Cancel/foo/bar=baz/server stream",
		"Cancel/foo/bar=baz/bidi stream",
		"Cancel/foo/bar=baz/(frobnitz)/unary",
		"Cancel/foo/bar=baz/(frobnitz)/client stream",
		"Cancel/foo/bar=baz/(frobnitz)/server stream",
		"Cancel/foo/bar=baz/(frobnitz)/bidi stream",
		"Timeout/foo/bar=baz/unary",
		"Timeout/foo/bar=baz/client stream",
		"Timeout/foo/bar=baz/server stream",
		"Timeout/foo/bar=baz/bidi stream",
		"Timeout/foo/bar=baz/(frobnitz)/unary",
		"Timeout/foo/bar=baz/(frobnitz)/client stream",
		"Timeout/foo/bar=baz/(frobnitz)/server stream",
		"Timeout/foo/bar=baz/(frobnitz)/bidi stream",
	}
	testCases := []struct {
		name                       string
		runPatterns, noRunPatterns []string
		keepers                    []string
	}{
		{
			name:    "no patterns",
			keepers: allTestCaseNames,
		},
		{
			name: "run patterns accept nothing",
			runPatterns: []string{
				"Foo/Bar/**",
				"**/blah blah blah", //nolint:dupword
			},
			keepers: []string{},
		},
		{
			name: "no-run patterns reject all",
			noRunPatterns: []string{
				"**/bar=baz/**",
			},
			keepers: []string{},
		},
		{
			name: "combined",
			runPatterns: []string{
				"Basic/**",
				"**/unary",
			},
			noRunPatterns: []string{
				"**/(frobnitz)/**",
				"**/bidi stream",
			},
			keepers: []string{
				"Basic/foo/bar=baz/unary",
				"Basic/foo/bar=baz/client stream",
				"Basic/foo/bar=baz/server stream",
				"Cancel/foo/bar=baz/unary",
				"Timeout/foo/bar=baz/unary",
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			candidates := make([]*conformancev1.TestCase, len(allTestCaseNames))
			for i, testCaseName := range allTestCaseNames {
				candidates[i] = &conformancev1.TestCase{Request: &conformancev1.ClientCompatRequest{
					TestName: testCaseName,
				}}
			}
			filter := newFilter(parsePatterns(testCase.runPatterns), parsePatterns(testCase.noRunPatterns))
			filtered := filter.apply(candidates)
			assert.Len(t, filtered, len(testCase.keepers))
			for i, testCaseName := range testCase.keepers {
				if i >= len(filtered) {
					break
				}
				assert.Equal(t, testCaseName, filtered[i].Request.TestName, "kept test case #%d", i+1)
			}
			keptSet := make(map[string]struct{}, len(testCase.keepers))
			for _, testCaseName := range testCase.keepers {
				keptSet[testCaseName] = struct{}{}
			}
			for i, testCaseName := range allTestCaseNames {
				_, shouldKeep := keptSet[testCaseName]
				keep := filter.accept(candidates[i])
				assert.Equal(t, shouldKeep, keep, "filter.accept(%q)", testCaseName)
			}
		})
	}
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
					Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
					Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
								Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
								Message: proto.String("message"),
								Details: asAnySlice(t, header),
							},
						},
					},
				}),
			},
			expected: &conformancev1.ClientResponseResult{
				Error: &conformancev1.Error{
					Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
					Message: proto.String("message"),
					Details: asAnySlice(t, header, &conformancev1.ConformancePayload_RequestInfo{
						Requests: asAnySlice(t, &conformancev1.UnaryRequest{
							ResponseDefinition: &conformancev1.UnaryResponseDefinition{
								Response: &conformancev1.UnaryResponseDefinition_Error{
									Error: &conformancev1.Error{
										Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
					Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
					Message: proto.String("message"),
					Details: asAnySlice(t, &conformancev1.ConformancePayload_RequestInfo{
						RequestHeaders: requestHeaders,
						Requests:       asAnySlice(t, unaryErrorReq),
						ConnectGetInfo: &conformancev1.ConformancePayload_ConnectGetInfo{
							QueryParams: []*conformancev1.Header{
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
					Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
					Message: proto.String("message"),
					Details: asAnySlice(t, &conformancev1.ConformancePayload_RequestInfo{
						RequestHeaders: requestHeaders,
						Requests:       asAnySlice(t, unaryErrorReq),
						ConnectGetInfo: &conformancev1.ConformancePayload_ConnectGetInfo{
							QueryParams: []*conformancev1.Header{
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
								Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
					Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
					Message: proto.String("message"),
					Details: asAnySlice(t, &conformancev1.ConformancePayload_RequestInfo{
						RequestHeaders: requestHeaders,
						Requests: asAnySlice(t, &conformancev1.ClientStreamRequest{
							ResponseDefinition: &conformancev1.UnaryResponseDefinition{
								ResponseHeaders: responseHeaders,
								Response: &conformancev1.UnaryResponseDefinition_Error{
									Error: &conformancev1.Error{
										Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
								Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
					Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
					Message: proto.String("message"),
					Details: asAnySlice(t, header, &conformancev1.ConformancePayload_RequestInfo{
						Requests: asAnySlice(t, &conformancev1.ClientStreamRequest{
							ResponseDefinition: &conformancev1.UnaryResponseDefinition{
								Response: &conformancev1.UnaryResponseDefinition_Error{
									Error: &conformancev1.Error{
										Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
							Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
										Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
					Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
							Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
					Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
					Message: proto.String("message"),
					Details: asAnySlice(t, &conformancev1.ConformancePayload_RequestInfo{
						RequestHeaders: requestHeaders,
						Requests: asAnySlice(t, &conformancev1.ServerStreamRequest{
							ResponseDefinition: &conformancev1.StreamResponseDefinition{
								ResponseHeaders: responseHeaders,
								ResponseDelayMs: 1000,
								Error: &conformancev1.Error{
									Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
							Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
										Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
					Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
							Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
					Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
					Message: proto.String("message"),
					Details: asAnySlice(t, &conformancev1.ConformancePayload_RequestInfo{
						RequestHeaders: requestHeaders,
						Requests: asAnySlice(t, &conformancev1.BidiStreamRequest{
							ResponseDefinition: &conformancev1.StreamResponseDefinition{
								ResponseHeaders: responseHeaders,
								ResponseDelayMs: 1000,
								Error: &conformancev1.Error{
									Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
							Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
										Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
					Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
							Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
					Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
					Message: proto.String("message"),
					Details: asAnySlice(t, &conformancev1.ConformancePayload_RequestInfo{
						Requests: asAnySlice(t, &conformancev1.BidiStreamRequest{
							ResponseDefinition: &conformancev1.StreamResponseDefinition{
								ResponseHeaders: responseHeaders,
								ResponseDelayMs: 1000,
								Error: &conformancev1.Error{
									Code:    conformancev1.Code_CODE_RESOURCE_EXHAUSTED,
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
// and verifies there are no errors during the conversion.
func asAnySlice(t *testing.T, msgs ...proto.Message) []*anypb.Any {
	t.Helper()
	arr := make([]*anypb.Any, 0, len(msgs))
	for _, msg := range msgs {
		asAny, err := anypb.New(msg)
		require.NoError(t, err)
		arr = append(arr, asAny)
	}
	return arr
}
