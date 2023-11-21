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
	"errors"
	"strings"
	"testing"

	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestResults_SetOutcome(t *testing.T) {
	t.Parallel()
	results := newResults(makeKnownFailing())
	results.setOutcome("foo/bar/1", false, nil)
	results.setOutcome("foo/bar/2", true, errors.New("fail"))
	results.setOutcome("foo/bar/3", false, errors.New("fail"))
	results.setOutcome("known-to-fail/1", false, nil)
	results.setOutcome("known-to-fail/2", true, errors.New("fail"))
	results.setOutcome("known-to-fail/3", false, errors.New("fail"))

	logger := &lineWriter{}
	success, err := results.report(logger)
	require.NoError(t, err)
	require.False(t, success)
	lines := logger.errorLines()
	require.Len(t, lines, 5)
	require.Equal(t, lines[0], "FAILED: foo/bar/2: fail\n")
	require.Equal(t, lines[1], "FAILED: foo/bar/3: fail\n")
	require.Equal(t, lines[2], "FAILED: known-to-fail/1 was expected to fail but did not\n")
	require.Equal(t, lines[3], "FAILED: known-to-fail/2: fail\n")
	require.Equal(t, lines[4], "INFO: known-to-fail/3 failed (as expected): fail\n")
}

func TestResults_FailedToStart(t *testing.T) {
	t.Parallel()
	results := newResults(makeKnownFailing())
	results.failedToStart([]*conformancev1.TestCase{
		{Request: &conformancev1.ClientCompatRequest{TestName: "foo/bar/1"}},
		{Request: &conformancev1.ClientCompatRequest{TestName: "known-to-fail/1"}},
	}, errors.New("fail"))

	logger := &lineWriter{}
	success, err := results.report(logger)
	require.NoError(t, err)
	require.False(t, success)
	lines := logger.errorLines()
	require.Len(t, lines, 2)
	require.Equal(t, lines[0], "FAILED: foo/bar/1: fail\n")
	// Marked as failure even though expected to fail because it failed to start.
	require.Equal(t, lines[1], "FAILED: known-to-fail/1: fail\n")
}

func TestResults_FailRemaining(t *testing.T) {
	t.Parallel()
	results := newResults(makeKnownFailing())
	results.setOutcome("foo/bar/1", false, nil)
	results.setOutcome("known-to-fail/1", false, errors.New("fail"))
	results.failRemaining([]*conformancev1.TestCase{
		{Request: &conformancev1.ClientCompatRequest{TestName: "foo/bar/1"}},
		{Request: &conformancev1.ClientCompatRequest{TestName: "foo/bar/2"}},
		{Request: &conformancev1.ClientCompatRequest{TestName: "known-to-fail/1"}},
		{Request: &conformancev1.ClientCompatRequest{TestName: "known-to-fail/2"}},
	}, errors.New("something went wrong"))

	logger := &lineWriter{}
	success, err := results.report(logger)
	require.NoError(t, err)
	require.False(t, success)
	lines := logger.errorLines()
	require.Len(t, lines, 3)
	require.Equal(t, lines[0], "FAILED: foo/bar/2: something went wrong\n")
	require.Equal(t, lines[1], "INFO: known-to-fail/1 failed (as expected): fail\n")
	// Marked as failure even though expected to fail because failRemaining is
	// used when a process under test dies (so this error is not due to lack of
	// conformance).
	require.Equal(t, lines[2], "FAILED: known-to-fail/2: something went wrong\n")
}

func TestResults_Failed(t *testing.T) {
	t.Parallel()
	results := newResults(makeKnownFailing())
	results.failed("foo/bar/1", &conformancev1.ClientErrorResult{Message: "fail"})
	results.failed("known-to-fail/1", &conformancev1.ClientErrorResult{Message: "fail"})

	logger := &lineWriter{}
	success, err := results.report(logger)
	require.NoError(t, err)
	require.False(t, success)
	lines := logger.errorLines()
	require.Len(t, lines, 2)
	require.Equal(t, lines[0], "FAILED: foo/bar/1: fail\n")
	require.Equal(t, lines[1], "INFO: known-to-fail/1 failed (as expected): fail\n")
}

func TestResults_Assert(t *testing.T) {
	t.Parallel()
	results := newResults(makeKnownFailing())
	payload1 := &conformancev1.ClientResponseResult{
		Payloads: []*conformancev1.ConformancePayload{
			{Data: []byte{0, 1, 2, 3, 4}},
		},
	}
	payload2 := &conformancev1.ClientResponseResult{
		Error: &conformancev1.Error{Code: int32(connect.CodeAborted), Message: proto.String("oops")},
	}
	results.assert("foo/bar/1", payload1, payload2)
	results.assert("foo/bar/2", payload2, payload1)
	results.assert("foo/bar/3", payload1, payload1)
	results.assert("foo/bar/4", payload2, payload2)
	results.assert("known-to-fail/1", payload1, payload2)
	results.assert("known-to-fail/2", payload2, payload1)
	results.assert("known-to-fail/3", payload1, payload1)
	results.assert("known-to-fail/4", payload2, payload2)

	logger := &lineWriter{}
	success, err := results.report(logger)
	require.NoError(t, err)
	require.False(t, success)
	lines := logger.errorLines()
	require.Len(t, lines, 6)
	require.Contains(t, lines[0], "FAILED: foo/bar/1: ")
	require.Contains(t, lines[1], "FAILED: foo/bar/2: ")
	require.Contains(t, lines[2], "INFO: known-to-fail/1 failed (as expected): ")
	require.Contains(t, lines[3], "INFO: known-to-fail/2 failed (as expected): ")
	require.Equal(t, lines[4], "FAILED: known-to-fail/3 was expected to fail but did not\n")
	require.Equal(t, lines[5], "FAILED: known-to-fail/4 was expected to fail but did not\n")
}

func TestResults_Assert_ReportsAllErrors(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name             string
		expected, actual string
		expectedErrors   []string
	}{
		{
			name: "identical",
			expected: `{
				"response_headers": [
					{"name": "abc", "value": ["xyz","123"]}
				],
				"error": {
					"code": 5,
					"message": "foobar",
					"details": [
						{"@type":"/google.protobuf.Empty", "value":{}}
					]
				},
				"payloads": [
					{
						"data": "abcdefgh",
						"request_info": {
							"request_headers": [
								{"name": "foo", "value": ["bar", "baz"]}
							],
							"timeout_ms": 12345,
							"requests": [
								{"@type": "/google.protobuf.Int32Value", "value": 123}
							]
						}
					}
				],
				"response_trailers": [
					{"name": "xyz", "value": ["value1"]}
				]
			}`,
		},
		{
			name: "superset request headers allowed",
			expected: `{
				"payloads": [
					{
						"request_info": {
							"request_headers": [
								{"name": "abc", "value": ["xyz", "123"]},
								{"name": "xyz", "value": ["value1"]},
								{"name": "Case-Does-Not-Matter-For-Name", "value": ["value2"]}
							]
						}
					}
				]
			}`,
			actual: `{
				"payloads": [
					{
						"request_info": {
							"request_headers": [
								{"name": "User-Agent", "value": ["blah blah blah"]},
								{"name": "abc", "value": ["xyz", "123"]},
								{"name": "xyz", "value": ["value1"]},
								{"name": "case-does-not-matter-for-name", "value": ["value2"]},
								{"name": "Content-Type", "value": ["application/json"]}
							]
						}
					}
				]
			}`,
		},
		{
			name: "superset response headers allowed",
			expected: `{
				"response_headers": [
					{"name": "abc", "value": ["xyz", "123"]},
					{"name": "xyz", "value": ["value1"]},
					{"name": "Case-Does-Not-Matter-For-Name", "value": ["value2"]}
				]
			}`,
			actual: `{
				"response_headers": [
					{"name": "User-Agent", "value": ["blah blah blah"]},
					{"name": "abc", "value": ["xyz", "123"]},
					{"name": "xyz", "value": ["value1"]},
					{"name": "case-does-not-matter-for-name", "value": ["value2"]},
					{"name": "Content-Type", "value": ["application/json"]}
				]
			}`,
		},
		{
			name: "superset response trailers allowed",
			expected: `{
				"response_trailers": [
					{"name": "abc", "value": ["xyz", "123"]},
					{"name": "xyz", "value": ["value1"]},
					{"name": "Case-Does-Not-Matter-For-Name", "value": ["value2"]}
				]
			}`,
			actual: `{
				"response_trailers": [
					{"name": "User-Agent", "value": ["blah blah blah"]},
					{"name": "abc", "value": ["xyz", "123"]},
					{"name": "xyz", "value": ["value1"]},
					{"name": "case-does-not-matter-for-name", "value": ["value2"]},
					{"name": "Content-Type", "value": ["application/json"]}
				]
			}`,
		},
		{
			name: "response headers or trailers missing/misattributed",
			expected: `{
				"error": {"code": 5},
				"response_headers": [
					{"name": "abc", "value": ["xyz", "123"]},
					{"name": "xyz", "value": ["value1"]}
				],
				"response_trailers": [
					{"name": "Case-Does-Not-Matter-For-Name", "value": ["value2"]}
				]
			}`,
			actual: `{
				"error": {"code": 5},
				"response_headers": [
					{"name": "Case-Does-Not-Matter-For-Name", "value": ["value2"]}
				],
				"response_trailers": [
					{"name": "abc", "value": ["xyz", "123"]},
					{"name": "xyz", "value": ["value1"]}
				]
			}`,
			expectedErrors: []string{
				`actual response headers missing "abc"`,
				`actual response headers missing "xyz"`,
				`actual response trailers missing "case-does-not-matter-for-name"`,
			},
		},
		{
			name: "response meta misattributed allowed for trailers-only response (all in headers)",
			expected: `{
				"error": {"code": 5},
				"response_headers": [
					{"name": "abc", "value": ["xyz", "123"]},
					{"name": "xyz", "value": ["value1"]}
				],
				"response_trailers": [
					{"name": "Case-Does-Not-Matter-For-Name", "value": ["value2"]}
				]
			}`,
			actual: `{
				"error": {"code": 5},
				"response_headers": [
					{"name": "abc", "value": ["xyz", "123"]},
					{"name": "xyz", "value": ["value1"]},
					{"name": "Case-Does-Not-Matter-For-Name", "value": ["value2"]}
				]
			}`,
		},
		{
			name: "response meta misattributed allowed for trailers-only response (all in trailers)",
			expected: `{
				"error": {"code": 5},
				"response_headers": [
					{"name": "abc", "value": ["xyz", "123"]},
					{"name": "xyz", "value": ["value1"]}
				],
				"response_trailers": [
					{"name": "Case-Does-Not-Matter-For-Name", "value": ["value2"]}
				]
			}`,
			actual: `{
				"error": {"code": 5},
				"response_trailers": [
					{"name": "abc", "value": ["xyz", "123"]},
					{"name": "xyz", "value": ["value1"]},
					{"name": "Case-Does-Not-Matter-For-Name", "value": ["value2"]}
				]
			}`,
		},
		{
			name: "error code mismatch",
			expected: `{
				"error": {
					"code": 5,
					"message": "foobar"
				}
			}`,
			actual: `{
				"error": {
					"code": 11,
					"message": "foobar"
				}
			}`,
			expectedErrors: []string{
				"actual error code 11 (out_of_range) does not match expected code 5 (not_found)",
			},
		},
		{
			name: "error message mismatch",
			expected: `{
				"error": {
					"code": 5,
					"message": "foobar"
				}
			}`,
			actual: `{
				"error": {
					"code": 5,
					"message": "oof!"
				}
			}`,
			expectedErrors: []string{
				`actual error message "oof!" does not match expected message "foobar"`,
			},
		},
		{
			name: "error detail mismatch",
			expected: `{
				"error": {
					"code": 5,
					"message": "foobar",
					"details": [
						{
							"@type": "/google.protobuf.Int32Value",
							"value": 123
						},
						{
							"@type": "/google.protobuf.StringValue",
							"value": "foobar"
						}
					]
				}
			}`,
			actual: `{
				"error": {
					"code": 5,
					"message": "foobar",
					"details": [
						{
							"@type": "/google.protobuf.Int32Value",
							"value": 456
						},
						{
							"@type": "/google.protobuf.StringValue",
							"value": "bobloblaw"
						}
					]
				}
			}`,
			expectedErrors: []string{
				"actual error detail #1 does not match expected error detail",
				"actual error detail #2 does not match expected error detail",
			},
		},
		{
			name: "missing error",
			expected: `{
				"payloads": [
					{
						"data": "abcdefgh",
						"request_info": {
							"request_headers": [
								{"name": "abc", "value": ["xyz", "123"]},
								{"name": "xyz", "value": ["value1"]},
								{"name": "Case-Does-Not-Matter-For-Name", "value": ["value2"]}
							]
						}
					}
				],
				"error": {
					"code": 5,
					"message": "foobar"
				}
			}`,
			actual: `{
				"payloads": [
					{
						"data": "abcdefgh",
						"request_info": {
							"request_headers": [
								{"name": "abc", "value": ["xyz", "123"]},
								{"name": "xyz", "value": ["value1"]},
								{"name": "Case-Does-Not-Matter-For-Name", "value": ["value2"]}
							]
						}
					}
				]
			}`,
			expectedErrors: []string{
				"expecting an error but received none",
			},
		},
		{
			name: "unexpected error",
			expected: `{
				"payloads": [
					{
						"data": "abcdefgh",
						"request_info": {
							"request_headers": [
								{"name": "abc", "value": ["xyz", "123"]},
								{"name": "xyz", "value": ["value1"]},
								{"name": "Case-Does-Not-Matter-For-Name", "value": ["value2"]}
							]
						}
					}
				]
			}`,
			actual: `{
				"payloads": [
					{
						"data": "abcdefgh",
						"request_info": {
							"request_headers": [
								{"name": "abc", "value": ["xyz", "123"]},
								{"name": "xyz", "value": ["value1"]},
								{"name": "Case-Does-Not-Matter-For-Name", "value": ["value2"]}
							]
						}
					}
				],
				"error": {
					"code": 5,
					"message": "foobar"
				}
			}`,
			expectedErrors: []string{
				"received an unexpected error",
			},
		},
		{
			name: "mismatch response count",
			expected: `{
				"payloads": [
					{
						"data": "abcdefgh",
						"request_info": {
							"request_headers": [
								{"name": "abc", "value": ["xyz", "123"]},
								{"name": "xyz", "value": ["value1"]},
								{"name": "Case-Does-Not-Matter-For-Name", "value": ["value2"]}
							]
						}
					}
				]
			}`,
			actual: `{
				"payloads": [
					{
						"data": "abcdefgh",
						"request_info": {
							"request_headers": [
								{"name": "abc", "value": ["xyz", "123"]},
								{"name": "xyz", "value": ["value1"]},
								{"name": "Case-Does-Not-Matter-For-Name", "value": ["value2"]}
							]
						}
					},
					{
						"data": "abcdefgh"
					}
				]
			}`,
			expectedErrors: []string{
				"expecting 1 response messages but instead got 2",
			},
		},
		{
			name: "mismatch response data",
			expected: `{
				"payloads": [
					{
						"data": "abcdefgh",
						"request_info": {
							"request_headers": [
								{"name": "abc", "value": ["xyz", "123"]},
								{"name": "xyz", "value": ["value1"]},
								{"name": "Case-Does-Not-Matter-For-Name", "value": ["value2"]}
							]
						}
					},
					{
						"data": "12345678"
					}
				]
			}`,
			actual: `{
				"payloads": [
					{
						"data": "abcdefgh",
						"request_info": {
							"request_headers": [
								{"name": "abc", "value": ["xyz", "123"]},
								{"name": "xyz", "value": ["value1"]},
								{"name": "Case-Does-Not-Matter-For-Name", "value": ["value2"]}
							]
						}
					},
					{
						"data": "abcdefgh"
					}
				]
			}`,
			expectedErrors: []string{
				"response #2: expecting data d76df8e7aefc, got 69b71d79f821",
			},
		},
		{
			name: "mismatch request count",
			expected: `{
				"payloads": [
					{
						"data": "abcdefgh",
						"request_info": {
							"requests": [
								{"@type": "/google.protobuf.Int32Value", "value": 123},
								{"@type": "/google.protobuf.Int32Value", "value": 456},
								{"@type": "/google.protobuf.Int32Value", "value": 789}
							]
						}
					}
				]
			}`,
			actual: `{
				"payloads": [
					{
						"data": "abcdefgh",
						"request_info": {
							"requests": [
								{"@type": "/google.protobuf.Int32Value", "value": 123},
								{"@type": "/google.protobuf.Int32Value", "value": 456}
							]
						}
					}
				]
			}`,
			expectedErrors: []string{
				"expecting 3 request messages to be described but instead got 2",
			},
		},
		{
			name: "mismatch request data",
			expected: `{
				"payloads": [
					{
						"data": "abcdefgh",
						"request_info": {
							"requests": [
								{"@type": "/google.protobuf.Int32Value", "value": 123},
								{"@type": "/google.protobuf.Int32Value", "value": 456}
							]
						}
					}
				]
			}`,
			actual: `{
				"payloads": [
					{
						"data": "abcdefgh",
						"request_info": {
							"requests": [
								{"@type": "/google.protobuf.Int32Value", "value": 101},
								{"@type": "/google.protobuf.Int32Value", "value": 404}
							]
						}
					}
				]
			}`,
			expectedErrors: []string{
				"request #1: did not survive round-trip",
				"request #2: did not survive round-trip",
			},
		},
		{
			name: "everything is wrong 😱",
			expected: `{
				"response_headers": [
					{"name": "abc", "value": ["xyz","123"]}
				],
				"error": {
					"code": 5,
					"message": "foobar",
					"details": [
						{"@type":"/google.protobuf.Empty", "value":{}}
					]
				},
				"payloads": [
					{
						"data": "abcdefgh",
						"request_info": {
							"request_headers": [
								{"name": "foo", "value": ["bar", "baz"]}
							],
							"timeout_ms": 12345,
							"requests": [
								{"@type": "/google.protobuf.Int32Value", "value": 123}
							]
						}
					},
					{
						"data": "abcdefgh"
					},
					{
						"data": "abcdefgh"
					}
				],
				"response_trailers": [
					{"name": "xyz", "value": ["value1"]}
				]
			}`,
			actual: `{
				"payloads": [
					{
						"data": "1234",
						"request_info": {
							"requests": [
								{"@type": "/google.protobuf.Int32Value", "value": 999},
								{"@type": "/google.protobuf.Int32Value", "value": 123}
							]
						}
					}
				]
			}`,
			// It tries to describe everything wrong, all in one shot.
			expectedErrors: []string{
				`actual response headers missing "abc"`,
				`actual response trailers missing "xyz"`,
				`expecting 3 response messages but instead got 1`,
				`response #1: expecting data 69b71d79f821, got d76df8`,
				`actual request headers missing "foo"`,
				`server did not echo back a timeout but one was expected (12345 ms)`,
				`expecting 1 request messages to be described but instead got 2`,
				`request #1: did not survive round-trip`,
				`expecting an error but received none`,
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			results := newResults(&knownFailingTrie{})

			expected := &conformancev1.ClientResponseResult{}
			err := protojson.Unmarshal(([]byte)(testCase.expected), expected)
			require.NoError(t, err)

			actual := &conformancev1.ClientResponseResult{}
			actualJSON := testCase.actual
			if actualJSON == "" {
				actualJSON = testCase.expected
			}
			err = protojson.Unmarshal(([]byte)(actualJSON), actual)
			require.NoError(t, err)

			results.assert(testCase.name, expected, actual)
			err = results.outcomes[testCase.name].actualFailure
			if len(testCase.expectedErrors) == 0 {
				require.NoError(t, err)
			} else {
				var errs multiErrors
				if !errors.As(err, &errs) {
					errs = multiErrors{err}
				}
				assert.Len(t, errs, len(testCase.expectedErrors))
				for i := 0; i < len(errs) && i < len(testCase.expectedErrors); i++ {
					assert.ErrorContains(t, errs[i], testCase.expectedErrors[i])
				}
			}
		})
	}
}

func TestResults_ServerSideband(t *testing.T) {
	t.Parallel()
	results := newResults(makeKnownFailing())
	results.setOutcome("foo/bar/1", false, nil)
	results.setOutcome("foo/bar/2", false, errors.New("fail"))
	results.setOutcome("foo/bar/3", false, nil)
	results.setOutcome("known-to-fail/1", false, nil)
	results.setOutcome("known-to-fail/2", false, errors.New("fail"))
	results.recordSideband("foo/bar/2", "something awkward in wire format")
	results.recordSideband("foo/bar/3", "something awkward in wire format")
	results.recordSideband("known-to-fail/1", "something awkward in wire format")

	logger := &lineWriter{}
	success, err := results.report(logger)
	require.NoError(t, err)
	require.False(t, success)
	lines := logger.errorLines()
	require.Len(t, lines, 4)
	require.Equal(t, lines[0], "FAILED: foo/bar/2: something awkward in wire format; fail\n")
	require.Equal(t, lines[1], "FAILED: foo/bar/3: something awkward in wire format\n")
	require.Equal(t, lines[2], "INFO: known-to-fail/1 failed (as expected): something awkward in wire format\n")
	require.Equal(t, lines[3], "INFO: known-to-fail/2 failed (as expected): fail\n")
}

func TestResults_Report(t *testing.T) {
	t.Parallel()
	results := newResults(makeKnownFailing())
	logger := &lineWriter{}

	// No test cases? Report success.
	success, err := results.report(logger)
	require.NoError(t, err)
	require.True(t, success)

	// Only successful outcomes? Report success.
	results = newResults(makeKnownFailing())
	results.setOutcome("foo/bar/1", false, nil)
	success, err = results.report(logger)
	require.NoError(t, err)
	require.True(t, success)

	// Unexpected failure? Report failure.
	results = newResults(makeKnownFailing())
	results.setOutcome("foo/bar/1", false, errors.New("ruh roh"))
	success, err = results.report(logger)
	require.NoError(t, err)
	require.False(t, success)

	// Unexpected failure during setup? Report failure.
	results = newResults(makeKnownFailing())
	results.setOutcome("foo/bar/1", true, errors.New("ruh roh"))
	success, err = results.report(logger)
	require.NoError(t, err)
	require.False(t, success)

	// Expected failure? Report success.
	results = newResults(makeKnownFailing())
	results.setOutcome("known-to-fail/1", false, errors.New("ruh roh"))
	success, err = results.report(logger)
	require.NoError(t, err)
	require.True(t, success)

	// Setup error from expected failure? Report failure (setup errors never acceptable).
	results = newResults(makeKnownFailing())
	results.setOutcome("known-to-fail/1", true, errors.New("ruh roh"))
	success, err = results.report(logger)
	require.NoError(t, err)
	require.False(t, success)
}

func makeKnownFailing() *knownFailingTrie {
	var trie knownFailingTrie
	trie.add([]string{"known-to-fail", "**"})
	return &trie
}

type lineWriter struct {
	current []byte
	lines   []string
}

func (l *lineWriter) Write(data []byte) (n int, err error) {
	for {
		if len(data) == 0 {
			return n, nil
		}
		var hasLF bool
		pos := bytes.IndexByte(data, '\n')
		if pos == -1 {
			pos = len(data)
		} else {
			pos++ // include LF
			hasLF = true
		}
		l.current = append(l.current, data[:pos]...)
		if hasLF {
			l.lines = append(l.lines, string(l.current))
			l.current = nil
		}
		data = data[pos:]
	}
}

func (l *lineWriter) errorLines() []string {
	var lines []string
	for _, line := range l.lines {
		if strings.HasPrefix(line, "FAILED: ") || strings.HasPrefix(line, "INFO: ") {
			lines = append(lines, line)
		}
	}
	return lines
}
