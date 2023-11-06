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
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"

	conformancev1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/anypb"
)

const timeoutCheckGracePeriodMillis = 500

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
	var errs multiErrors

	if len(expected.Payloads) == 0 && expected.Error != nil && (len(actual.ResponseHeaders) == 0 || len(actual.ResponseTrailers) == 0) {
		// When there are no messages in the body, only an error, the server may send a
		// trailers-only response. In that case, it is acceptable for the client the
		// expected headers and trailers to be merged into one set, and it is acceptable
		// for the client to interpret them as either headers or trailers.
		merged := mergeHeaders(expected.ResponseHeaders, expected.ResponseTrailers)
		var actualHeaders []*conformancev1alpha1.Header
		if len(actual.ResponseHeaders) == 0 {
			actualHeaders = actual.ResponseTrailers
		} else {
			actualHeaders = actual.ResponseHeaders
		}
		errs = append(errs, checkHeaders("response metadata", merged, actualHeaders)...)
	} else {
		errs = append(errs, checkHeaders("response headers", expected.ResponseHeaders, actual.ResponseHeaders)...)
		errs = append(errs, checkHeaders("response trailers", expected.ResponseTrailers, actual.ResponseTrailers)...)
	}

	errs = append(errs, checkPayloads(expected.Payloads, actual.Payloads)...)

	if diff := cmp.Diff(expected.Error, actual.Error, protocmp.Transform()); diff != "" {
		errs = append(errs, fmt.Errorf("actual error does not match expected error: - wanted, + got\n%s", diff))
	}

	// If client didn't provide actual raw error, we skip this check.
	if expected.ConnectErrorRaw != nil && actual.ConnectErrorRaw != nil {
		diff := cmp.Diff(expected.ConnectErrorRaw, actual.ConnectErrorRaw, protocmp.Transform())
		if diff != "" {
			errs = append(errs, fmt.Errorf("raw Connect error does not match: - wanted, + got\n%s", diff))
		}
	}

	r.setOutcome(testCase, false, errs.Result())
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
	var succeeded, failed, expectedFailures int
	sort.Strings(testCaseNames)
	for _, name := range testCaseNames {
		outcome := r.outcomes[name]
		expectError := outcome.knownFailing && !outcome.setupError
		var err error
		switch {
		case !expectError && outcome.actualFailure != nil:
			_, err = fmt.Fprintf(writer, "FAILED: %s: %v\n", name, outcome.actualFailure)
			failed++
		case expectError && outcome.actualFailure == nil:
			_, err = fmt.Fprintf(writer, "FAILED: %s was expected to fail but did not\n", name)
			failed++
		case expectError && outcome.actualFailure != nil:
			_, err = fmt.Fprintf(writer, "INFO: %s failed (as expected): %v\n", name, outcome.actualFailure)
			expectedFailures++
		default:
			succeeded++
		}
		if err != nil {
			return err
		}
	}
	if failed+expectedFailures > 0 {
		// Add a blank line to separate summary from messages above
		_, err := writer.Write([]byte{'\n'})
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(writer, "Total cases: %d\n%d passed, %d failed\n", len(r.outcomes), succeeded, failed)
	if err != nil {
		return err
	}
	if expectedFailures > 0 {
		_, err := fmt.Fprintf(writer, "(%d failed as expected due to being known failures.)\n", expectedFailures)
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

type multiErrors []error

func (e multiErrors) Error() string {
	var buf bytes.Buffer
	for i, err := range e {
		if i > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(err.Error())
	}
	return buf.String()
}

func (e multiErrors) Result() error {
	switch len(e) {
	case 0:
		return nil
	case 1:
		return e[0]
	default:
		return e
	}
}

func mergeHeaders(a, b []*conformancev1alpha1.Header) []*conformancev1alpha1.Header {
	mergedMap := map[string][]string{}
	for _, hdr := range a {
		mergedMap[strings.ToLower(hdr.Name)] = hdr.Value
	}
	for _, hdr := range b {
		mergedMap[strings.ToLower(hdr.Name)] = append(mergedMap[strings.ToLower(hdr.Name)], hdr.Value...)
	}
	results := make([]*conformancev1alpha1.Header, 0, len(mergedMap))
	for k, v := range mergedMap {
		results = append(results, &conformancev1alpha1.Header{Name: k, Value: v})
	}
	return results
}

func checkHeaders(what string, expected, actual []*conformancev1alpha1.Header) multiErrors {
	var errs multiErrors
	actualHeaders := map[string][]string{}
	for _, hdr := range actual {
		actualHeaders[strings.ToLower(hdr.Name)] = hdr.Value
	}
	for _, hdr := range expected {
		name := strings.ToLower(hdr.Name)
		actualVals, ok := actualHeaders[name]
		if !ok {
			errs = append(errs, fmt.Errorf("actual %s missing %q", what, name))
			continue
		}
		actualStr := headerValsToString(actualVals)
		expectedStr := headerValsToString(hdr.Value)
		if actualStr != expectedStr {
			errs = append(errs, fmt.Errorf("%s has incorrect values for %q: expected [%s], got [%s]", what, name, expectedStr, actualStr))
		}
	}
	return errs
}

func headerValsToString(vals []string) string {
	var buf bytes.Buffer
	for i, val := range vals {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(strconv.Quote(val))
	}
	return buf.String()
}

func checkPayloads(expected, actual []*conformancev1alpha1.ConformancePayload) multiErrors {
	var errs multiErrors
	if len(actual) != len(expected) {
		errs = append(errs, fmt.Errorf("expecting %d response messages but instead got %d", len(expected), len(actual)))
	}
	reqNum := 0
	for i := 0; i < len(actual) && i < len(expected); i++ {
		actualPayload := actual[i]
		expectedPayload := expected[i]
		if !bytes.Equal(actualPayload.Data, expectedPayload.Data) {
			errs = append(errs, fmt.Errorf("response #%d: expecting data %x, got %x", i+1, expectedPayload.Data, actualPayload.Data))
		}
		actualReq := actualPayload.GetRequestInfo()
		expectedReq := expectedPayload.GetRequestInfo()

		if i == 0 { //nolint:nestif
			// Validate headers, timeout, and query params. We only need to do this once, for first payload.
			errs = append(errs, checkHeaders("request headers", expectedReq.GetRequestHeaders(), actualReq.GetRequestHeaders())...)
			if expectedReq != nil && expectedReq.TimeoutMs != nil {
				if actualReq == nil || actualReq.TimeoutMs == nil {
					errs = append(errs, fmt.Errorf("server did not echo back a timeout but one was expected (%d ms)", expectedReq.GetTimeoutMs()))
				} else {
					max := expectedReq.GetTimeoutMs()
					min := max - timeoutCheckGracePeriodMillis
					if min < 0 {
						min = 0
					}
					if actualReq.GetTimeoutMs() > max || actualReq.GetTimeoutMs() < min {
						errs = append(errs, fmt.Errorf("server echoed back a timeout (%d ms) that did not match expected (%d ms)", actualReq.GetTimeoutMs(), expectedReq.GetTimeoutMs()))
					}
				}
			} else if actualReq != nil && actualReq.TimeoutMs != nil {
				errs = append(errs, fmt.Errorf("server echoed back a timeout (%d ms) but none was expected", actualReq.GetTimeoutMs()))
			}
			if len(expectedReq.GetConnectGetInfo().GetQueryParams()) > 0 && len(actualReq.GetConnectGetInfo().GetQueryParams()) > 0 {
				errs = append(errs, checkHeaders("request query params", expectedReq.GetConnectGetInfo().GetQueryParams(), actualReq.GetConnectGetInfo().GetQueryParams())...)
			}
		}

		if len(actualReq.GetRequests()) != len(expectedReq.GetRequests()) {
			errs = append(errs, fmt.Errorf("response #%d: expecting %d request messages to be described but instead got %d", i+1, len(expectedReq.GetRequests()), len(actualReq.GetRequests())))
		}
		for i := 0; i < len(actualReq.GetRequests()) && i < len(expectedReq.GetRequests()); i++ {
			reqNum++
			actualMsg, err := anypb.UnmarshalNew(actualReq.GetRequests()[i], proto.UnmarshalOptions{})
			if err != nil {
				errs = append(errs, fmt.Errorf("request #%d: failed to unmarshal actual message: %w", reqNum, err))
				continue
			}
			expectedMsg, err := anypb.UnmarshalNew(expectedReq.GetRequests()[i], proto.UnmarshalOptions{})
			if err != nil {
				errs = append(errs, fmt.Errorf("request #%d: failed to unmarshal expected message: %w", reqNum, err))
				continue
			}
			diff := cmp.Diff(expectedMsg, actualMsg, protocmp.Transform())
			if diff != "" {
				errs = append(errs, fmt.Errorf("request #%d: did not survive round-trip: - wanted, + got\n%s", reqNum, diff))
			}
		}
	}

	return errs
}
