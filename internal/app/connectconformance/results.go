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
	"sort"
	"strings"
	"sync"

	conformancev1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"google.golang.org/protobuf/proto"
)

// testResults represents the results of running conformance tests. It accumulates
// the state of passed and failed test cases and also reports failures to a given
// log writer. It can also incorporate data provided out-of-band by a reference
// server, when testing a client implementation.
type testResults struct {
	knownFailing *knownFailingTrie

	mu             sync.Mutex
	outcomes       map[string]testOutcome
	serverSideband map[string]string
}

func newResults(knownFailing *knownFailingTrie) *testResults {
	return &testResults{
		knownFailing:   knownFailing,
		outcomes:       map[string]testOutcome{},
		serverSideband: map[string]string{},
	}
}

// setOutcome sets the outcome for the named test case. If setupError is true,
// then err occurred before the test case could actually be run. Otherwise, err
// represents the result of issuing the RPC, which may be nil to indicate
// the test case passed.
func (r *testResults) setOutcome(testCase string, setupError bool, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.setOutcomeLocked(testCase, setupError, err)
}

func (r *testResults) setOutcomeLocked(testCase string, setupError bool, err error) {
	r.outcomes[testCase] = testOutcome{
		actualFailure: err,
		setupError:    setupError,
		knownFailing:  r.knownFailing.match(strings.Split(testCase, "/")),
	}
}

// failedToStart marks all the given test cases with the given setup error.
// This convenience method is to mark many tests in a batch when the relevant
// server process could not be started.
func (r *testResults) failedToStart(testCases []*conformancev1alpha1.TestCase, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, testCase := range testCases {
		r.setOutcomeLocked(testCase.Request.TestName, true, err)
	}
}

// failRemaining marks any of the given test cases that do not yet have an outcome
// as failing with the given error. This is typically called when the server or client
// process fails, so we can mark any pending test.
func (r *testResults) failRemaining(testCases []*conformancev1alpha1.TestCase, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, testCase := range testCases {
		name := testCase.Request.TestName
		if _, outcomeExists := r.outcomes[name]; outcomeExists {
			continue
		}
		r.setOutcomeLocked(testCase.Request.TestName, true, err)
	}
}

// failed marks the given test case as having failed with the given error
// message received from the client.
func (r *testResults) failed(testCase string, err *conformancev1alpha1.ClientErrorResult) {
	r.setOutcome(testCase, false, errors.New(err.Message))
}

// assert will examine the actual and expected RPC result and mark the test
// case as successful or failed accordingly.
func (r *testResults) assert(testCase string, expected, actual *conformancev1alpha1.ClientResponseResult) {
	// TODO: Need to do smart processing of expected and actual to make sure
	//       actual *complies* with expected (doesn't necessarily have to match
	//       exactly; for example extra response headers are fine...)
	//       Also, more bespoke checks will also result in better error messages
	//       for users, so they know what went wrong.
	var err error
	if !proto.Equal(expected, actual) {
		err = &expectationFailedError{expected: expected, actual: actual}
	}
	r.setOutcome(testCase, false, err)
}

// recordServerSideband accepts an error message for a test that was sent
// out-of-band by a reference server.
func (r *testResults) recordServerSideband(testCase string, errMsg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.serverSideband[testCase] = errMsg
}

// processServerSidebandInfoLocked merges the data recorded during calls
// to recordServerSideband into the outcomes. This is done when a report
// is created.
func (r *testResults) processServerSidebandInfoLocked() {
	for name, msg := range r.serverSideband {
		outcome, ok := r.outcomes[name]
		if ok {
			// Update outcome to include reference server's feedback
			if outcome.actualFailure == nil {
				outcome.actualFailure = errors.New(msg)
			} else {
				outcome.actualFailure = fmt.Errorf("%s; %w", msg, outcome.actualFailure)
			}
			r.outcomes[name] = outcome
		} else {
			r.setOutcomeLocked(name, false, errors.New(msg))
		}
	}
}

func (r *testResults) report(writer io.Writer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.serverSideband) > 0 {
		r.processServerSidebandInfoLocked()
		r.serverSideband = map[string]string{}
	}
	testCaseNames := make([]string, 0, len(r.outcomes))
	for testCaseName := range r.outcomes {
		testCaseNames = append(testCaseNames, testCaseName)
	}
	sort.Strings(testCaseNames)
	for _, name := range testCaseNames {
		outcome := r.outcomes[name]
		expectError := outcome.knownFailing && !outcome.setupError
		var err error
		switch {
		case !expectError && outcome.actualFailure != nil:
			_, err = fmt.Fprintf(writer, "FAILED: %s: %v\n", name, outcome.actualFailure)
		case expectError && outcome.actualFailure == nil:
			_, err = fmt.Fprintf(writer, "FAILED: %s was expected to fail but did not\n", name)
		case expectError && outcome.actualFailure != nil:
			_, err = fmt.Fprintf(writer, "INFO: %s failed (as expected): %v\n", name, outcome.actualFailure)
		}
		if err != nil {
			return err
		}
	}
	return nil
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
