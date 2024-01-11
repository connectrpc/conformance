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
	"fmt"
	"testing"

	"connectrpc.com/conformance/internal/app/connectconformance/testsuites"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Parallel()

	// Simple sanity check that the basics of run work.
	testSuiteData, err := testsuites.LoadTestSuites()
	require.NoError(t, err)
	allSuites, err := parseTestSuites(testSuiteData)
	require.NoError(t, err)
	configCases := []configCase{
		{
			Version:                conformancev1.HTTPVersion_HTTP_VERSION_1,
			Protocol:               conformancev1.Protocol_PROTOCOL_CONNECT,
			Codec:                  conformancev1.Codec_CODEC_JSON,
			Compression:            conformancev1.Compression_COMPRESSION_IDENTITY,
			StreamType:             conformancev1.StreamType_STREAM_TYPE_UNARY,
			UseTLS:                 false,
			UseTLSClientCerts:      false,
			UseConnectGET:          false,
			UseMessageReceiveLimit: false,
			ConnectVersionMode:     conformancev1.TestSuite_CONNECT_VERSION_MODE_UNSPECIFIED,
		},
		{
			Version:                conformancev1.HTTPVersion_HTTP_VERSION_2,
			Protocol:               conformancev1.Protocol_PROTOCOL_GRPC,
			Codec:                  conformancev1.Codec_CODEC_PROTO,
			Compression:            conformancev1.Compression_COMPRESSION_IDENTITY,
			StreamType:             conformancev1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
			UseTLS:                 false,
			UseTLSClientCerts:      false,
			UseConnectGET:          false,
			UseMessageReceiveLimit: false,
			ConnectVersionMode:     conformancev1.TestSuite_CONNECT_VERSION_MODE_UNSPECIFIED,
		},
	}

	// Compute expected number of test cases.
	testCaseLib, err := newTestCaseLibrary(allSuites, configCases, conformancev1.TestSuite_TEST_MODE_UNSPECIFIED)
	require.NoError(t, err)
	allTestCases := calcAllTestCases(testCaseLib.testCases, true, true)
	expectedNumCases := len(allTestCases)

	// 19 test cases as of this writing, but we will likely add more
	require.GreaterOrEqual(t, expectedNumCases, 19)

	logger := &testPrinter{t}
	results, err := run(
		configCases,
		&knownFailingTrie{},
		allSuites,
		logger,
		logger,
		&Flags{Verbose: true, MaxServers: 2, Parallelism: 4},
	)

	require.NoError(t, err)
	require.True(t, results.report(logger))
	require.Equal(t, expectedNumCases, len(results.outcomes))
}

type testPrinter struct {
	t *testing.T
}

func (t testPrinter) Printf(msg string, args ...any) {
	t.t.Logf(msg, args...)
}

func (t testPrinter) PrefixPrintf(prefix, msg string, args ...any) {
	msg = fmt.Sprintf(msg, args...)
	t.t.Logf("%s: %s", prefix, msg)
}
