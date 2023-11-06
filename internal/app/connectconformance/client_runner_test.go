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
	"context"
	"errors"
	"io"
	"math/rand"
	"sort"
	"testing"

	"connectrpc.com/conformance/internal"
	conformancev1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunClient(t *testing.T) {
	t.Parallel()

	testReqs := []*conformancev1alpha1.ClientCompatRequest{
		{
			TestName: "TestSuite1/testcase1",
		},
		{
			TestName: "TestSuite1/testcase2",
		},
		{
			TestName: "TestSuite2/testcase1",
		},
		{
			TestName: "TestSuite2/testcase2",
		},
	}

	testCases := []struct {
		name            string
		clientFunc      func(_ context.Context, _ []string, in io.ReadCloser, out, _ io.WriteCloser) error
		expectErr       string
		failToSend      int
		expectedResults map[string]bool
	}{
		{
			name:       "simple",
			clientFunc: (&testClientProcess{}).run,
			expectedResults: map[string]bool{
				"TestSuite1/testcase1": true,
				"TestSuite1/testcase2": true,
				"TestSuite2/testcase1": true,
				"TestSuite2/testcase2": true,
			},
		},
		{
			name:       "client fails",
			clientFunc: (&testClientProcess{failAfter: 2}).run,
			failToSend: 2,
			expectErr:  "could not write request: client closed stdin",
			expectedResults: map[string]bool{
				"TestSuite1/testcase1": true,
				"TestSuite1/testcase2": true,
			},
		},
		{
			name:       "random order",
			clientFunc: testClientProcessRand,
			expectedResults: map[string]bool{
				"TestSuite1/testcase1": true,
				"TestSuite1/testcase2": true,
				"TestSuite2/testcase1": true,
				"TestSuite2/testcase2": true,
			},
		},
		{
			name:       "broken",
			clientFunc: testClientProcessBroken,
			expectErr:  "broken",
			expectedResults: map[string]bool{
				"TestSuite1/testcase1": false,
				"TestSuite1/testcase2": false,
				"TestSuite2/testcase1": false,
				"TestSuite2/testcase2": false,
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			start := runInProcess("testclient", testCase.clientFunc)
			runner, err := runClient(context.Background(), start)
			require.NoError(t, err)

			actualResults := make(map[string]bool, len(testReqs))
			var actualFailedToSend int
			for i, req := range testReqs {
				err := runner.sendRequest(req, func(name string, _ *conformancev1alpha1.ClientCompatResponse, err error) {
					if err != nil {
						t.Logf("error for %s: %v", name, err)
					}
					actualResults[name] = err == nil
				})
				if err != nil {
					actualFailedToSend = len(testReqs) - i
					break
				}
			}
			runner.closeSend()

			err = runner.waitForResponses()
			if testCase.expectErr != "" {
				assert.ErrorContains(t, err, testCase.expectErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.failToSend, actualFailedToSend)
			assert.Empty(t, cmp.Diff(testCase.expectedResults, actualResults))
		})
	}
}

// testClientProcess reads requests from in and immediately writes a corresponding response to out.
type testClientProcess struct {
	failAfter int
}

func (c *testClientProcess) run(_ context.Context, _ []string, in io.ReadCloser, out, _ io.WriteCloser) error {
	var count int
	for {
		req := &conformancev1alpha1.ClientCompatRequest{}
		if err := internal.ReadDelimitedMessage(in, req); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		resp := &conformancev1alpha1.ClientCompatResponse{
			TestName: req.TestName,
			Result: &conformancev1alpha1.ClientCompatResponse_Response{
				Response: &conformancev1alpha1.ClientResponseResult{
					Payloads: []*conformancev1alpha1.ConformancePayload{
						{Data: []byte{0, 1, 2, 3, 4}},
					},
				},
			},
		}
		if err := internal.WriteDelimitedMessage(out, resp); err != nil {
			return err
		}
		count++
		if c.failAfter > 0 && count >= c.failAfter {
			return errors.New("failed")
		}
	}
}

func testClientProcessRand(_ context.Context, _ []string, in io.ReadCloser, out, _ io.WriteCloser) error {
	var allCases []string
	for {
		req := &conformancev1alpha1.ClientCompatRequest{}
		if err := internal.ReadDelimitedMessage(in, req); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		allCases = append(allCases, req.TestName)
	}

	for {
		rand.Shuffle(len(allCases), func(i, j int) {

		})
		// just make sure we didn't shuffle to a non-random permutation
		isSorted := !sort.SliceIsSorted(allCases, func(i, j int) bool {
			return allCases[i] < allCases[j]
		})
		if isSorted {
			continue // try again
		}
		break
	}

	for _, name := range allCases {
		resp := &conformancev1alpha1.ClientCompatResponse{
			TestName: name,
			Result: &conformancev1alpha1.ClientCompatResponse_Response{
				Response: &conformancev1alpha1.ClientResponseResult{
					Payloads: []*conformancev1alpha1.ConformancePayload{
						{Data: []byte{0, 1, 2, 3, 4}},
					},
				},
			},
		}
		if err := internal.WriteDelimitedMessage(out, resp); err != nil {
			return err
		}
	}
	return nil
}

func testClientProcessBroken(_ context.Context, _ []string, in io.ReadCloser, _, _ io.WriteCloser) error {
	_, _ = io.Copy(io.Discard, in)
	return errors.New("broken")
}
