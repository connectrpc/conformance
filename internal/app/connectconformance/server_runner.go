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
	"bufio"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"connectrpc.com/conformance/internal"
	conformancev1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func populateExpectedUnaryResponse(testCase *conformancev1alpha1.TestCase) error {
	req := testCase.Request.RequestMessages[0]
	// First, find the response definition that the client instructed the server to return
	concreteReq, err := req.UnmarshalNew()
	if err != nil {
		return err
	}
	type unaryResponseDefiner interface {
		GetResponseDefinition() *conformancev1alpha1.UnaryResponseDefinition
	}

	definer, ok := concreteReq.(unaryResponseDefiner)
	if !ok {
		return fmt.Errorf("%T is not a unary test case", concreteReq)
	}

	// TODO - Need to define this better in the protos and tests as to how services should
	// behave if no responses are specified. The behavior right now differs for unary vs. streaming
	// If no responses are specified for unary, the service will still return a response with the
	// request information inside (but none of the response information since it wasn't provided)
	// But streaming endpoints don't return a single response and instead return responses via sending
	// on a stream. But if no responses are specified in the request, the streams don't send anything outbound
	// so there's no way to relay this to a client. So right now, streaming endpoints simply expect an empty
	// ClientResponseResult if no response definition is provided
	def := definer.GetResponseDefinition()
	if def == nil {
		testCase.ExpectedResponse = &conformancev1alpha1.ClientResponseResult{
			Payloads: []*conformancev1alpha1.ConformancePayload{
				{
					RequestInfo: &conformancev1alpha1.ConformancePayload_RequestInfo{
						RequestHeaders: testCase.Request.RequestHeaders,
						Requests:       testCase.Request.RequestMessages,
					},
				},
			},
		}
		return nil
	}

	// Server should have echoed back all specified headers and trailers
	expected := &conformancev1alpha1.ClientResponseResult{
		ResponseHeaders:  def.ResponseHeaders,
		ResponseTrailers: def.ResponseTrailers,
	}

	switch respType := def.Response.(type) {
	case *conformancev1alpha1.UnaryResponseDefinition_Error:
		// If an error was specified, it should be returned in the response
		expected.Error = respType.Error
	case *conformancev1alpha1.UnaryResponseDefinition_ResponseData, nil:
		// If response data was specified for the response (or nothing at all),
		// the server should echo back the request message and headers in the response
		payload := &conformancev1alpha1.ConformancePayload{
			RequestInfo: &conformancev1alpha1.ConformancePayload_RequestInfo{
				RequestHeaders: testCase.Request.RequestHeaders,
				Requests:       testCase.Request.RequestMessages,
			},
		}
		// If response data was specified for the response, it should be returned
		if respType, ok := respType.(*conformancev1alpha1.UnaryResponseDefinition_ResponseData); ok {
			payload.Data = respType.ResponseData
		}
		expected.Payloads = []*conformancev1alpha1.ConformancePayload{payload}
	default:
		return fmt.Errorf("provided UnaryRequest.Response has an unexpected type %T", respType)
	}

	testCase.ExpectedResponse = expected
	return nil
}

func populateExpectedStreamResponse(testCase *conformancev1alpha1.TestCase) error {
	req := testCase.Request.RequestMessages[0]
	// First, find the response definition that the client instructed the
	// server to return
	concreteReq, err := req.UnmarshalNew()
	if err != nil {
		return err
	}
	type streamResponseDefiner interface {
		GetResponseDefinition() *conformancev1alpha1.StreamResponseDefinition
	}

	definer, ok := concreteReq.(streamResponseDefiner)
	if !ok {
		return fmt.Errorf(
			"TestCase %s contains a request message of type %T, which is not a streaming request",
			testCase.Request.TestName,
			concreteReq,
		)
	}

	def := definer.GetResponseDefinition()
	if def == nil {
		testCase.ExpectedResponse = &conformancev1alpha1.ClientResponseResult{}
		return nil
	}

	// Server should have echoed back all specified headers, trailers, and errors
	expected := &conformancev1alpha1.ClientResponseResult{
		ResponseHeaders:  def.ResponseHeaders,
		ResponseTrailers: def.ResponseTrailers,
		Error:            def.Error,
	}

	// There should be one payload for every ResponseData the client specified
	expected.Payloads = make([]*conformancev1alpha1.ConformancePayload, len(def.ResponseData))

	for idx, data := range def.ResponseData {
		expected.Payloads[idx] = &conformancev1alpha1.ConformancePayload{
			Data: data,
		}
		switch testCase.Request.StreamType { //nolint:exhaustive
		case conformancev1alpha1.StreamType_STREAM_TYPE_SERVER_STREAM,
			conformancev1alpha1.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM:
			// For server streams and half duplex bidi streams, all request information
			// specified should only be echoed back in the first response
			if idx == 0 {
				expected.Payloads[idx].RequestInfo = &conformancev1alpha1.ConformancePayload_RequestInfo{
					RequestHeaders: testCase.Request.RequestHeaders,
					Requests:       testCase.Request.RequestMessages,
				}
			}
		case conformancev1alpha1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM:
			// For a full duplex stream, the first request should be echoed back in the first
			// payload. The second should be echoed back in the second payload, etc. (i.e. a ping pong interaction)
			expected.Payloads[idx].RequestInfo = &conformancev1alpha1.ConformancePayload_RequestInfo{
				// RequestHeaders: testCase.Request.RequestHeaders,
				Requests: []*anypb.Any{testCase.Request.RequestMessages[idx]},
			}
			if idx == 0 {
				expected.Payloads[idx].RequestInfo.RequestHeaders = testCase.Request.RequestHeaders
			}
		}
	}
	testCase.ExpectedResponse = expected
	return nil
}

// populateExpectedResponse populates the response we expected to get back from the server
// by examining the requests we sent.
func populateExpectedResponse(testCase *conformancev1alpha1.TestCase) error {
	// If an expected response was already provided, return and use that.
	// This allows for overriding this function with explicit values in the yaml file.
	if testCase.ExpectedResponse != nil {
		return nil
	}
	// TODO - This is just a temporary constraint to protect against panics for now.
	// Eventually, we want to be able to test client and bidi streams where there are no request messages.
	// The potential plan is for server impls to produce (and the code below to expect) a single response
	// message in this situation, where the response data value is some fixed string (such as "no response definition")
	// and whose request info will still be present, but we expect it to indicate zero request messages.
	if len(testCase.Request.RequestMessages) == 0 {
		return errors.New("at least one request is required")
	}

	switch testCase.Request.StreamType {
	case conformancev1alpha1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
		conformancev1alpha1.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM,
		conformancev1alpha1.StreamType_STREAM_TYPE_SERVER_STREAM:
		return populateExpectedStreamResponse(testCase)

	case conformancev1alpha1.StreamType_STREAM_TYPE_UNARY,
		conformancev1alpha1.StreamType_STREAM_TYPE_CLIENT_STREAM:
		return populateExpectedUnaryResponse(testCase)

	case conformancev1alpha1.StreamType_STREAM_TYPE_UNSPECIFIED:
		return errors.New("stream type is required")
	default:
		return fmt.Errorf("stream type %s is not supported", testCase.Request.StreamType)
	}
}

// runTestCasesForServer runs starts a server process and runs the given test cases while
// it is active. The test cases are executed by serializing the request, writing to the
// given requestWriter, and then awaiting a corresponding response to be read from the
// given responseReader. The given results are used to record the state of all cases and
// their actual responses.
//
// The given cancel function should be invoked if an I/O error occurs while interacting
// with the requestWriter or responseReader.
//
// If isReferenceServer is true, then the server's stderr will be examined as well, to
// record out-of-band feedback about the client requests.
func runTestCasesForServer(
	ctx context.Context,
	isReferenceClient bool,
	isReferenceServer bool,
	meta serverInstance,
	testCases []*conformancev1alpha1.TestCase,
	startServer processStarter,
	results *testResults,
	client clientRunner,
) {
	expectations := make(map[string]*conformancev1alpha1.ClientResponseResult, len(testCases))
	for _, testCase := range testCases {
		err := populateExpectedResponse(testCase)
		if err != nil {
			results.recordSideband(testCase.Request.TestName, err.Error())
		}
		expectations[testCase.Request.TestName] = testCase.ExpectedResponse
	}

	procCtx, procCancel := context.WithCancel(ctx)
	defer procCancel()
	serverProcess, err := startServer(procCtx, isReferenceServer)
	if err != nil {
		results.failedToStart(testCases, fmt.Errorf("error starting server: %w", err))
		return
	}
	defer serverProcess.abort()
	serverProcess.whenDone(func(_ error) {
		procCancel()
	})

	var refServerFinished chan struct{}
	if isReferenceServer { //nolint:nestif
		refServerFinished = make(chan struct{})
		go func() {
			defer close(refServerFinished)
			r := bufio.NewReader(serverProcess.stderr)
			for {
				str, err := r.ReadString('\n')
				str = strings.TrimSpace(str)
				if str != "" {
					parts := strings.SplitN(str, ": ", 2)
					if len(parts) == 2 {
						if _, ok := expectations[parts[0]]; ok {
							// appears to be valid message in the form "test case: error message"
							results.recordSideband(parts[0], parts[1])
						}
					}
				}
				if err != nil {
					return
				}
			}
		}()
	}

	// Write server request.
	err = internal.WriteDelimitedMessage(serverProcess.stdin, &conformancev1alpha1.ServerCompatRequest{
		Protocol:    meta.protocol,
		HttpVersion: meta.httpVersion,
		UseTls:      meta.useTLS,
	})
	if err != nil {
		results.failedToStart(testCases, fmt.Errorf("error writing server request: %w", err))
		return
	}
	if err := serverProcess.stdin.Close(); err != nil {
		results.failedToStart(testCases, fmt.Errorf("error writing server request: %w", err))
		return
	}

	// Read response.
	var resp conformancev1alpha1.ServerCompatResponse
	err = internal.ReadDelimitedMessage(serverProcess.stdout, &resp)
	if err != nil {
		results.failedToStart(testCases, fmt.Errorf("error reading server response: %w", err))
		return
	}

	// Send all test cases to the client.
	var wg sync.WaitGroup
	for i, testCase := range testCases {
		if procCtx.Err() != nil {
			// server crashed: mark remaining tests
			err := errors.New("server process terminated unexpectedly")
			for j := i; j < len(testCases); j++ {
				results.setOutcome(testCases[j].Request.TestName, true, err)
			}
			return
		}
		req := proto.Clone(testCase.Request).(*conformancev1alpha1.ClientCompatRequest) //nolint:errcheck,forcetypeassert
		req.Host = resp.Host
		req.Port = resp.Port
		req.ServerTlsCert = resp.PemCert
		if isReferenceServer {
			req.RequestHeaders = append(
				req.RequestHeaders,
				&conformancev1alpha1.Header{Name: "x-test-case-name", Value: []string{testCase.Request.TestName}},
				&conformancev1alpha1.Header{Name: "x-expect-http-version", Value: []string{strconv.Itoa(int(req.HttpVersion))}},
				&conformancev1alpha1.Header{Name: "x-expect-protocol", Value: []string{strconv.Itoa(int(req.Protocol))}},
				&conformancev1alpha1.Header{Name: "x-expect-codec", Value: []string{strconv.Itoa(int(req.Codec))}},
				&conformancev1alpha1.Header{Name: "x-expect-compression", Value: []string{strconv.Itoa(int(req.Compression))}},
			)
		}

		wg.Add(1)
		err := client.sendRequest(req, func(name string, resp *conformancev1alpha1.ClientCompatResponse, err error) {
			defer wg.Done()
			switch {
			case err != nil:
				results.setOutcome(name, true, err)
			case resp.GetError() != nil:
				results.failed(name, resp.GetError())
			default:
				results.assert(name, expectations[resp.TestName], resp.GetResponse())
			}
			if isReferenceClient {
				for _, msg := range resp.Feedback {
					results.recordSideband(resp.TestName, msg)
				}
			}
		})
		if err != nil {
			wg.Done() // call it explicitly since callback above won't be invoked
			// client pipe broken: mark remaining tests, including this one, as failed
			for j := i; j < len(testCases); j++ {
				results.setOutcome(testCases[j].Request.TestName, true, err)
			}
		}
	}

	// Wait for all responses.
	wg.Wait()

	serverProcess.abort()
	_ = serverProcess.result() // wait for server process to end
	if isReferenceServer {
		<-refServerFinished
	}

	// If there are any tests without outcomes, mark them now.
	results.failRemaining(testCases, errors.New("no outcome received from the client"))
}
