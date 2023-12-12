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
	"net/http"
	"strconv"
	"strings"
	"sync"

	"connectrpc.com/conformance/internal"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"google.golang.org/protobuf/proto"
)

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
	testCases []*conformancev1.TestCase,
	clientCreds *conformancev1.ClientCompatRequest_TLSCreds,
	startServer processStarter,
	errPrinter internal.Printer,
	results *testResults,
	client clientRunner,
) {
	expectations := make(map[string]*conformancev1.ClientResponseResult, len(testCases))
	for _, testCase := range testCases {
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
				origLine, err := r.ReadString('\n')
				str := strings.TrimSpace(origLine)
				if str != "" {
					var isSideband bool
					parts := strings.SplitN(str, ": ", 2)
					if len(parts) == 2 {
						if _, ok := expectations[parts[0]]; ok {
							// appears to be valid message in the form "test case: error message"
							isSideband = true
							results.recordSideband(parts[0], parts[1])
						}
					}
					if !isSideband {
						// Was some other message printed to stderr. Propagate to our stderr so user can see it.
						errPrinter.PrefixPrintf("referenceserver", "%s", origLine)
					}
				}
				if err != nil {
					return
				}
			}
		}()
	}

	if !meta.useTLSClientCerts {
		// don't send client cert info if these tests don't use them
		clientCreds = nil
	}

	// Write server request.
	err = internal.WriteDelimitedMessage(serverProcess.stdin, &conformancev1.ServerCompatRequest{
		Protocol:      meta.protocol,
		HttpVersion:   meta.httpVersion,
		UseTls:        meta.useTLS,
		ClientTlsCert: clientCreds.GetCert(),
		// We always set this. If server-under-test does not support it, we just
		// won't run the test cases that verify that it's enforced.
		MessageReceiveLimit: serverReceiveLimit,
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
	var resp conformancev1.ServerCompatResponse
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
		req := proto.Clone(testCase.Request).(*conformancev1.ClientCompatRequest) //nolint:errcheck,forcetypeassert
		req.Host = resp.Host
		req.Port = resp.Port
		req.ServerTlsCert = resp.PemCert
		req.ClientTlsCreds = clientCreds
		if isReferenceServer {
			httpMethod := http.MethodPost
			if req.UseGetHttpMethod {
				httpMethod = http.MethodGet
			}
			req.RequestHeaders = append(
				req.RequestHeaders,
				&conformancev1.Header{Name: "x-test-case-name", Value: []string{testCase.Request.TestName}},
				&conformancev1.Header{Name: "x-expect-http-version", Value: []string{strconv.Itoa(int(req.HttpVersion))}},
				&conformancev1.Header{Name: "x-expect-http-method", Value: []string{httpMethod}},
				&conformancev1.Header{Name: "x-expect-protocol", Value: []string{strconv.Itoa(int(req.Protocol))}},
				&conformancev1.Header{Name: "x-expect-codec", Value: []string{strconv.Itoa(int(req.Codec))}},
				&conformancev1.Header{Name: "x-expect-compression", Value: []string{strconv.Itoa(int(req.Compression))}},
				&conformancev1.Header{Name: "x-expect-tls", Value: []string{strconv.FormatBool(len(resp.PemCert) > 0)}},
			)
			if clientCreds != nil {
				req.RequestHeaders = append(
					req.RequestHeaders,
					&conformancev1.Header{Name: "x-expect-client-cert", Value: []string{internal.ClientCertName}},
				)
			}
		}

		wg.Add(1)
		err := client.sendRequest(req, func(name string, resp *conformancev1.ClientCompatResponse, err error) {
			defer wg.Done()
			switch {
			case err != nil:
				results.setOutcome(name, true, err)
			case resp.GetError() != nil:
				results.failed(name, resp.GetError())
			default:
				results.assert(name, expectations[resp.TestName], resp.GetResponse())
			}
			if isReferenceClient && resp != nil {
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
