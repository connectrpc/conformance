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
	"connectrpc.com/conformance/internal/tracer"
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
//
//nolint:gocyclo
func runTestCasesForServer(
	ctx context.Context,
	isReferenceClient bool,
	isReferenceServer bool,
	meta serverInstance,
	testCases []*conformancev1.TestCase,
	clientCreds *conformancev1.ClientCompatRequest_TLSCreds,
	startServer processStarter,
	logPrinter internal.Printer,
	errPrinter internal.Printer,
	results *testResults,
	client clientRunner,
	tracer *tracer.Tracer,
	logEach bool,
) {
	testCaseNameSet := make(map[string]struct{}, len(testCases))
	for _, testCase := range testCases {
		testCaseNameSet[testCase.Request.TestName] = struct{}{}
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
						if _, ok := testCaseNameSet[parts[0]]; ok {
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
	if meta.useTLS && len(resp.PemCert) == 0 {
		results.failedToStart(testCases, errors.New("server config uses TLS, but server response did not indicate a certificate"))
		return
	}

	// Send all test cases to the client.
	var wg sync.WaitGroup
	for i := range testCases {
		testCase := testCases[i]
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
		if req.Host == "" {
			req.Host = internal.DefaultHost
		}
		req.Port = resp.Port
		req.ServerTlsCert = resp.PemCert
		req.ClientTlsCreds = clientCreds

		// We always include test name in request header.
		testCaseHeader := &conformancev1.Header{Name: "x-test-case-name", Value: []string{testCase.Request.TestName}}
		req.RequestHeaders = append(req.RequestHeaders, testCaseHeader)
		if req.RawRequest != nil {
			req.RawRequest.Headers = append(req.RawRequest.Headers, testCaseHeader)
		}
		if isReferenceServer {
			// The reference server wants more metadata in headers, to perform add'l validations.
			httpMethod := http.MethodPost
			if req.UseGetHttpMethod {
				httpMethod = http.MethodGet
			}
			extraHeaders := []*conformancev1.Header{
				{Name: "x-expect-http-version", Value: []string{strconv.Itoa(int(req.HttpVersion))}},
				{Name: "x-expect-http-method", Value: []string{httpMethod}},
				{Name: "x-expect-protocol", Value: []string{strconv.Itoa(int(req.Protocol))}},
				{Name: "x-expect-codec", Value: []string{strconv.Itoa(int(req.Codec))}},
				{Name: "x-expect-compression", Value: []string{strconv.Itoa(int(req.Compression))}},
				{Name: "x-expect-tls", Value: []string{strconv.FormatBool(len(resp.PemCert) > 0)}},
			}
			if clientCreds != nil {
				extraHeaders = append(
					extraHeaders,
					&conformancev1.Header{Name: "x-expect-client-cert", Value: []string{internal.ClientCertName}},
				)
			}
			req.RequestHeaders = append(req.RequestHeaders, extraHeaders...)
			if req.RawRequest != nil {
				req.RawRequest.Headers = append(req.RawRequest.Headers, extraHeaders...)
			}
		}

		tracer.Init(req.TestName)
		wg.Add(1)
		if logEach {
			logPrinter.Printf("Sending request for %q...", req.TestName)
		}
		err := client.sendRequest(req, func(name string, resp *conformancev1.ClientCompatResponse, err error) {
			defer wg.Done()
			if logEach {
				logPrinter.Printf("Received response for %q...", req.TestName)
			}
			switch {
			case err != nil:
				results.setOutcome(name, true, err)
			case resp.GetError() != nil:
				results.failed(name, resp.GetError())
			case resp.GetResponse() != nil:
				results.assert(name, testCase, resp.GetResponse())
			default:
				results.setOutcome(name, false, errors.New("client returned a response with neither an error nor result"))
			}
			if isReferenceClient && resp.GetResponse() != nil {
				for _, msg := range resp.GetResponse().Feedback {
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
