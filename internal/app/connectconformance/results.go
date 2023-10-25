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
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	conformancev1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"google.golang.org/protobuf/proto"
)

type testResults struct {
	knownFailing *knownFailingTrie
	logOut       io.Writer

	mu             sync.Mutex
	outcomes       map[string]testOutcome
	serverSideband map[string]string
}

func newResults(knownFailing *knownFailingTrie, logOut io.Writer) *testResults {
	return &testResults{
		knownFailing:   knownFailing,
		logOut:         logOut,
		outcomes:       map[string]testOutcome{},
		serverSideband: map[string]string{},
	}
}
func (r *testResults) setOutcome(testCase string, setupError bool, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.setOutcomeLocked(testCase, setupError, err)
}

func (r *testResults) setOutcomeLocked(testCase string, setupError bool, err error) {
	if err != nil {
		_, _ = fmt.Fprintf(r.logOut, "FAILED: %s: %v\n", testCase, err)
	}
	r.outcomes[testCase] = testOutcome{
		actualFailure: err,
		setupError:    setupError,
		knownFailing:  r.knownFailing.match(strings.Split(testCase, "/")),
	}
}

func (r *testResults) failedToStart(testCases []*conformancev1alpha1.TestCase, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, testCase := range testCases {
		r.setOutcomeLocked(testCase.Request.TestName, true, err)
	}
}

func (r *testResults) failRemaining(testCases []*conformancev1alpha1.TestCase, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, testCase := range testCases {
		name := testCase.Request.TestName
		if _, outcomeExists := r.outcomes[name]; outcomeExists {
			continue
		}
		r.setOutcomeLocked(testCase.Request.TestName, false, err)
	}
}

func (r *testResults) recordServerSideband(testCase string, errMsg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.serverSideband[testCase] = errMsg
}

func (r *testResults) processServerSidebandInfo() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for name, err := range r.serverSideband {
		outcome, ok := r.outcomes[name]
		if ok {
			// Update outcome to include reference server's feedback
			if outcome.actualFailure == nil {
				outcome.actualFailure = errors.New(err)
			} else {
				outcome.actualFailure = fmt.Errorf("%s; %w", err, outcome.actualFailure)
			}
		} else {
			r.setOutcomeLocked(name, false, errors.New(err))
		}
	}
}

func (r *testResults) failed(testCase string, err *conformancev1alpha1.ClientErrorResult) {
	r.setOutcome(testCase, false, errors.New(err.Message))
}

func (r *testResults) assert(testCase string, expected, actual *conformancev1alpha1.ClientResponseResult) {
	// TODO: need to do smart processing of expected and actual to make sure
	//       actual *complies* with expected (doesn't necessarily have to match
	//       exactly; for example extra response headers are fine...)
	var err error
	if !proto.Equal(expected, actual) {
		err = &expectationFailedError{expected: expected, actual: actual}
	}
	r.setOutcome(testCase, false, err)
}

type testOutcome struct {
	// nil if the test case executed successfully, otherwise an error that
	// represents why the test case failed, such as an error returned by the
	// client or an "expectation failure", when the client's result does not
	// match the expected result.
	actualFailure error
	// if actualFailure != nil and setupError is true, the error occurred while
	// setting up the test, not as an outcome of running the test.
	setupError bool
	// true if this test case is known to fail
	knownFailing bool
}

type expectationFailedError struct {
	expected, actual *conformancev1alpha1.ClientResponseResult
}

func (e *expectationFailedError) Error() string {
	return "actual result did not match expected result"
}
