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
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"

	"connectrpc.com/conformance/internal"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/tracer"
	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/anypb"
)

// CI environments have been observed to have 150ms of delay between the deadline
// being set and it being serialized in a request header. This adds a little extra
// headroom to avoid flaky tests.
const timeoutCheckGracePeriodMillis = 250

// testResults represents the results of running conformance tests. It accumulates
// the state of passed and failed test cases and also reports failures to a given
// log writer. It can also incorporate data provided out-of-band by a reference
// server, when testing a client implementation.
type testResults struct {
	knownFailing *testTrie
	knownFlaky   *testTrie
	tracer       *tracer.Tracer

	traceWaitGroup sync.WaitGroup

	mu             sync.Mutex
	outcomes       map[string]testOutcome
	traces         map[string]*tracer.Trace
	serverSideband map[string]string
}

func newResults(knownFailing, knownFlaky *testTrie, tracer *tracer.Tracer) *testResults {
	return &testResults{
		knownFailing:   knownFailing,
		knownFlaky:     knownFlaky,
		tracer:         tracer,
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
		knownFlaky:    r.knownFlaky.match(strings.Split(testCase, "/")),
	}
	r.fetchTrace(testCase)
}

//nolint:contextcheck,nolintlint // intentionally using context.Background; nolintlint incorrectly complains about this
func (r *testResults) fetchTrace(testCase string) {
	if r.tracer == nil {
		return
	}
	r.traceWaitGroup.Add(1)
	go func() {
		defer r.traceWaitGroup.Done()
		ctx, cancel := context.WithTimeout(context.Background(), tracer.TraceTimeout)
		defer cancel()
		trace, err := r.tracer.Await(ctx, testCase)
		r.tracer.Clear(testCase)
		if err != nil {
			return
		}

		r.mu.Lock()
		defer r.mu.Unlock()
		outcome := r.outcomes[testCase]
		if outcome.actualFailure == nil || outcome.setupError || outcome.knownFlaky || outcome.knownFailing {
			return
		}
		if r.traces == nil {
			r.traces = map[string]*tracer.Trace{}
		}
		r.traces[testCase] = trace
	}()
}

// failedToStart marks all the given test cases with the given setup error.
// This convenience method is to mark many tests in a batch when the relevant
// server process could not be started.
func (r *testResults) failedToStart(testCases []*conformancev1.TestCase, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, testCase := range testCases {
		r.setOutcomeLocked(testCase.Request.TestName, true, err)
	}
}

// failRemaining marks any of the given test cases that do not yet have an outcome
// as failing with the given error. This is typically called when the server or client
// process fails, so we can mark any pending test.
func (r *testResults) failRemaining(testCases []*conformancev1.TestCase, err error) {
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
func (r *testResults) failed(testCase string, err *conformancev1.ClientErrorResult) {
	r.setOutcome(testCase, false, errors.New(err.Message))
}

// assert will examine the actual and expected RPC result and mark the test
// case as successful or failed accordingly.
func (r *testResults) assert(
	testCase string,
	definition *conformancev1.TestCase,
	actual *conformancev1.ClientResponseResult,
) {
	expected := definition.ExpectedResponse
	var errs multiErrors

	errs = append(errs, checkError(expected.Error, actual.Error, definition.OtherAllowedErrorCodes)...)
	errs = append(errs, checkPayloads(expected.Payloads, actual.Payloads)...)

	if len(expected.Payloads) == 0 &&
		expected.Error != nil &&
		(definition.Request.StreamType == conformancev1.StreamType_STREAM_TYPE_UNARY ||
			definition.Request.StreamType == conformancev1.StreamType_STREAM_TYPE_CLIENT_STREAM) {
		// For unary and client-stream operations, a server API may not provide a way to
		// set headers and trailers separately on error but instead a way to return an
		// error with embedded metadata. In that case, we may not be able to distinguish
		// headers from trailers in the response -- they get unified into a single bag of
		// "error metadata". The conformance client should record those as trailers when
		// sending back a ClientResponseResult message.

		// So first we see if normal attribute succeeds
		metadataErrs := checkHeaders("response headers", expected.ResponseHeaders, actual.ResponseHeaders)
		metadataErrs = append(metadataErrs, checkHeaders("response trailers", expected.ResponseTrailers, actual.ResponseTrailers)...)
		if len(metadataErrs) > 0 {
			// That did not work. So we test to see if client attributed them all as trailers.
			merged := mergeHeaders(expected.ResponseHeaders, expected.ResponseTrailers)
			if allTrailersErrs := checkHeaders("response metadata", merged, actual.ResponseTrailers); len(allTrailersErrs) != 0 {
				// That check failed also. So the received headers/trailers are incorrect.
				// Report the original errors computed above.
				errs = append(errs, metadataErrs...)
			}
		}
	} else {
		errs = append(errs, checkHeaders("response headers", expected.ResponseHeaders, actual.ResponseHeaders)...)
		errs = append(errs, checkHeaders("response trailers", expected.ResponseTrailers, actual.ResponseTrailers)...)
	}

	if expected.HttpStatusCode != nil &&
		actual.HttpStatusCode != nil &&
		expected.GetHttpStatusCode() != actual.GetHttpStatusCode() {
		errs = append(errs, fmt.Errorf("actual HTTP status code does not match: wanted %d; got %d",
			expected.GetHttpStatusCode(), actual.GetHttpStatusCode()))
	}

	r.setOutcome(testCase, false, errs.Result())
}

// recordSideband accepts an error message for a test that was sent
// out-of-band by a reference server or included as feedback in the
// response from a reference client.
func (r *testResults) recordSideband(testCase string, errMsg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.serverSideband[testCase] = errMsg
}

// processSidebandInfoLocked merges the data recorded during calls
// to recordSideband into the outcomes. This is done when a report
// is created.
func (r *testResults) processSidebandInfoLocked() {
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

func (r *testResults) report(printer internal.Printer) bool {
	r.traceWaitGroup.Wait() // make sure all traces have been received
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.serverSideband) > 0 {
		r.processSidebandInfoLocked()
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
		var expectError bool
		if !outcome.setupError {
			expectError = outcome.knownFailing ||
				(outcome.knownFlaky && outcome.actualFailure != nil)
		}
		switch {
		case !expectError && outcome.actualFailure != nil:
			printer.Printf("FAILED: %s:\n%s", name, indent(outcome.actualFailure.Error()))
			trace := r.traces[name]
			if trace != nil {
				printer.Printf("---- HTTP Trace ----")
				trace.Print(printer)
				printer.Printf("--------------------")
			}
			failed++
		case expectError && outcome.actualFailure == nil:
			printer.Printf("FAILED: %s was expected to fail but did not", name)
			failed++
		case expectError && outcome.actualFailure != nil:
			printer.Printf("INFO: %s failed (as expected):\n%s", name, indent(outcome.actualFailure.Error()))
			expectedFailures++
		default:
			succeeded++
		}
	}
	if failed+expectedFailures > 0 {
		// Add a blank line to separate summary from messages above
		printer.Printf("\n")
	}
	printer.Printf("Total cases: %d\n%d passed, %d failed", len(r.outcomes), succeeded, failed)
	if expectedFailures > 0 {
		printer.Printf("(Another %d failed as expected due to being known failures/flakes.)", expectedFailures)
	}
	return failed == 0
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
	// true if this test case is known to be flaky
	knownFlaky bool
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

func mergeHeaders(a, b []*conformancev1.Header) []*conformancev1.Header {
	mergedMap := map[string][]string{}
	for _, hdr := range a {
		mergedMap[strings.ToLower(hdr.Name)] = hdr.Value
	}
	for _, hdr := range b {
		mergedMap[strings.ToLower(hdr.Name)] = append(mergedMap[strings.ToLower(hdr.Name)], hdr.Value...)
	}
	results := make([]*conformancev1.Header, 0, len(mergedMap))
	for k, v := range mergedMap {
		results = append(results, &conformancev1.Header{Name: k, Value: v})
	}
	return results
}

func checkHeaders(what string, expected, actual []*conformancev1.Header) multiErrors {
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
		expectedVals := canonicalizeHeaderVals(hdr.Value)
		actualVals = canonicalizeHeaderVals(actualVals)
		if !reflect.DeepEqual(expectedVals, actualVals) {
			errs = append(errs, fmt.Errorf("%s has incorrect values for %q: expected [%s], got [%s]",
				what, name, headerValsToString(hdr.Value), headerValsToString(actualVals)))
		}
	}
	return errs
}

func canonicalizeHeaderVals(vals []string) []string {
	canon := make([]string, 0, len(vals))
	for _, val := range vals {
		parts := strings.Split(val, ",")
		for i, last := 0, len(parts)-1; i <= last; i++ {
			part := parts[i]
			// We preserve any leading and trailing whitespace in the value.
			// But otherwise we will trim a single leading or trailing space
			// which a library may have added around a comma when combining
			// multiple header values into one.
			if i > 0 && part != "" && part[0] == ' ' {
				part = part[1:] // trim single leading space
			}
			if i < last && part != "" && part[len(part)-1] == ' ' {
				part = part[:len(part)-1] // trim single trailing space
			}
			canon = append(canon, part)
		}
	}
	return canon
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

func checkRequestInfo(expected, actual *conformancev1.ConformancePayload_RequestInfo, verifyHeaders bool) multiErrors {
	var errs multiErrors
	// If verifyHeaders is true, then verify headers, timeout, and query params. This is only needed when verifying
	// the first (or only) response received since that is what contains the header information
	if verifyHeaders { //nolint:nestif
		errs = append(errs, checkHeaders("request headers", expected.GetRequestHeaders(), actual.GetRequestHeaders())...)
		if expected != nil && expected.TimeoutMs != nil {
			if actual == nil || actual.TimeoutMs == nil {
				errs = append(errs, fmt.Errorf("server did not echo back a timeout but one was expected (%d ms)", expected.GetTimeoutMs()))
			} else {
				maxAllowed := expected.GetTimeoutMs()
				minAllowed := maxAllowed - timeoutCheckGracePeriodMillis
				if minAllowed < 0 {
					minAllowed = 0
				}
				if actual.GetTimeoutMs() > maxAllowed || actual.GetTimeoutMs() < minAllowed {
					errs = append(errs, fmt.Errorf("server echoed back a timeout (%d ms) that did not match expected (%d ms)", actual.GetTimeoutMs(), expected.GetTimeoutMs()))
				}
			}
		} else if actual != nil && actual.TimeoutMs != nil {
			errs = append(errs, fmt.Errorf("server echoed back a timeout (%d ms) but none was expected", actual.GetTimeoutMs()))
		}
		if len(expected.GetConnectGetInfo().GetQueryParams()) > 0 && len(actual.GetConnectGetInfo().GetQueryParams()) > 0 {
			errs = append(errs, checkHeaders("request query params", expected.GetConnectGetInfo().GetQueryParams(), actual.GetConnectGetInfo().GetQueryParams())...)
		}
	}

	if len(actual.GetRequests()) != len(expected.GetRequests()) {
		errs = append(errs, fmt.Errorf("expecting %d request messages to be described but instead got %d", len(expected.GetRequests()), len(actual.GetRequests())))
	}
	reqNum := 0
	for i := 0; i < len(actual.GetRequests()) && i < len(expected.GetRequests()); i++ {
		reqNum++
		actualMsg, err := anypb.UnmarshalNew(actual.GetRequests()[i], proto.UnmarshalOptions{})
		if err != nil {
			errs = append(errs, fmt.Errorf("request #%d: failed to unmarshal actual message: %w", reqNum, err))
			continue
		}
		expectedMsg, err := anypb.UnmarshalNew(expected.GetRequests()[i], proto.UnmarshalOptions{})
		if err != nil {
			errs = append(errs, fmt.Errorf("request #%d: failed to unmarshal expected message: %w", reqNum, err))
			continue
		}
		diff := cmp.Diff(expectedMsg, actualMsg, protocmp.Transform())
		if diff != "" {
			errs = append(errs, fmt.Errorf("request #%d: did not survive round-trip: - wanted, + got\n%s", reqNum, diff))
		}
	}
	return errs
}

func checkPayloads(expected, actual []*conformancev1.ConformancePayload) multiErrors {
	var errs multiErrors
	if len(actual) != len(expected) {
		errs = append(errs, fmt.Errorf("expecting %d response messages but instead got %d", len(expected), len(actual)))
	}
	for i := 0; i < len(actual) && i < len(expected); i++ {
		actualPayload := actual[i]
		expectedPayload := expected[i]
		if !bytes.Equal(actualPayload.Data, expectedPayload.Data) {
			errs = append(errs, fmt.Errorf("response #%d: expecting data %x, got %x", i+1, expectedPayload.Data, actualPayload.Data))
		}

		errs = append(errs, checkRequestInfo(expectedPayload.GetRequestInfo(), actualPayload.GetRequestInfo(), i == 0)...)
	}

	return errs
}

func checkError(expected, actual *conformancev1.Error, otherCodes []conformancev1.Code) multiErrors {
	switch {
	case expected == nil && actual == nil:
		// nothing to do
		return nil
	case expected == nil && actual != nil:
		return multiErrors{fmt.Errorf("received an unexpected error:\n%v", actual.String())}
	case expected != nil && actual == nil:
		return multiErrors{errors.New("expecting an error but received none")}
	}

	var errs multiErrors
	if expected.Code != actual.Code && !inSlice(actual.Code, otherCodes) {
		expectedCodes := expectedCodeString(expected.Code, otherCodes)
		errs = append(errs, fmt.Errorf("actual error {code: %d (%s), message: %q} does not match expected code %s",
			actual.Code, connect.Code(actual.Code).String(), actual.GetMessage(), expectedCodes))
	}
	if expected.Message != nil && expected.GetMessage() != actual.GetMessage() {
		errs = append(errs, fmt.Errorf("actual error {code: %d (%s), message: %q} does not match expected message %q",
			actual.Code, connect.Code(actual.Code).String(), actual.GetMessage(), expected.GetMessage()))
	}
	if len(expected.Details) != len(actual.Details) {
		// TODO: Should this be more lenient? Are we okay with a Connect implementation adding extra
		//       error details transparently (such that the expected details would be a *subset* of
		//       the actual details)?
		errs = append(errs, fmt.Errorf("actual error contain %d details; expecting %d",
			len(actual.Details), len(expected.Details)))
	}
	// Check as many as we can
	length := len(expected.Details)
	if len(actual.Details) < length {
		length = len(actual.Details)
	}
	actualReqInfo := &conformancev1.ConformancePayload_RequestInfo{}
	expectedReqInfo := &conformancev1.ConformancePayload_RequestInfo{}
	for i := 0; i < length; i++ {
		// TODO: Should this be more lenient? Are we okay with details getting re-ordered?
		//       An alternative might be to create a map keyed by type, and for each type
		//       remove expected messages as they are matched against actual ones.
		actualDetails := actual.Details[i]
		expectedDetails := expected.Details[i]

		// If the error details is a RequestInfo, then verify equality using the checkRequestInfo function
		// Otherwise, just do a straight diff of the two
		if actualDetails.MessageIs(actualReqInfo) && expectedDetails.MessageIs(expectedReqInfo) {
			// If unmarshalling fails for either detail, add it to the errors and move on to the next detail
			if err := actualDetails.UnmarshalTo(actualReqInfo); err != nil {
				errs = append(errs, fmt.Errorf("unable to unmarshal request info from actual error detail %s", actualDetails.MessageName()))
				continue
			}
			if err := expectedDetails.UnmarshalTo(expectedReqInfo); err != nil {
				errs = append(errs, fmt.Errorf("unable to unmarshal request info from expected error detail %s", expectedDetails.MessageName()))
				continue
			}
			errs = append(errs, checkRequestInfo(expectedReqInfo, actualReqInfo, true)...)
		} else {
			if diff := cmp.Diff(expectedDetails, actualDetails, protocmp.Transform()); diff != "" {
				errs = append(errs, fmt.Errorf("actual error detail #%d does not match expected error detail: - wanted, + got\n%s",
					i+1, diff))
			}
		}
	}
	return errs
}

func indent(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = "\t" + line
	}
	return strings.Join(lines, "\n")
}

func inSlice[T comparable](elem T, slice []T) bool {
	// TODO: delete this function when this repo is using Go 1.21
	//	     and update call sites to instead use slices.Contains
	for _, item := range slice {
		if item == elem {
			return true
		}
	}
	return false
}

func expectedCodeString(expectedCode conformancev1.Code, otherAllowedCodes []conformancev1.Code) string {
	allowedCodes := make([]string, len(otherAllowedCodes)+1)
	for i, code := range append([]conformancev1.Code{expectedCode}, otherAllowedCodes...) {
		allowedCodes[i] = fmt.Sprintf("%d (%s)", code, connect.Code(code).String())
		if i == len(allowedCodes)-1 && i != 0 {
			allowedCodes[i] = "or " + allowedCodes[i]
		}
	}
	if len(allowedCodes) < 3 {
		return strings.Join(allowedCodes, " ")
	}
	return strings.Join(allowedCodes, ", ")
}
