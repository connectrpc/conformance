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
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	conformancev1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestRunTestCasesForServer(t *testing.T) {
	t.Parallel()

	svrResponse := &conformancev1alpha1.ServerCompatResponse{
		Host:    "127.0.0.1",
		Port:    12345,
		PemCert: []byte("some cert"),
	}
	svrResponseData, err := proto.Marshal(svrResponse)
	require.NoError(t, err)

	svrInstance := serverInstance{
		protocol:    conformancev1alpha1.Protocol_PROTOCOL_GRPC_WEB,
		httpVersion: conformancev1alpha1.HTTPVersion_HTTP_VERSION_1,
		useTLS:      false,
	}
	expectedSvrReqData, err := proto.Marshal(&conformancev1alpha1.ServerCompatRequest{
		Protocol:    conformancev1alpha1.Protocol_PROTOCOL_GRPC_WEB,
		HttpVersion: conformancev1alpha1.HTTPVersion_HTTP_VERSION_1,
		UseTls:      false,
	})
	require.NoError(t, err)

	testCaseData := []*conformancev1alpha1.TestCase{
		{
			Request: &conformancev1alpha1.ClientCompatRequest{
				TestName: "TestSuite1/testcase1",
			},
			ExpectedResponse: &conformancev1alpha1.ClientResponseResult{
				Payloads: []*conformancev1alpha1.ConformancePayload{{Data: []byte("data")}},
			},
		},
		{
			Request: &conformancev1alpha1.ClientCompatRequest{
				TestName: "TestSuite1/testcase2",
			},
			ExpectedResponse: &conformancev1alpha1.ClientResponseResult{
				Payloads: []*conformancev1alpha1.ConformancePayload{{Data: []byte("data")}},
			},
		},
		{
			Request: &conformancev1alpha1.ClientCompatRequest{
				TestName: "TestSuite2/testcase1",
			},
			ExpectedResponse: &conformancev1alpha1.ClientResponseResult{
				Error: &conformancev1alpha1.Error{Code: int32(connect.CodeAborted), Message: "ruh roh"},
			},
		},
		{
			Request: &conformancev1alpha1.ClientCompatRequest{
				TestName: "TestSuite2/testcase2",
			},
			ExpectedResponse: &conformancev1alpha1.ClientResponseResult{
				Payloads: []*conformancev1alpha1.ConformancePayload{{Data: []byte("data")}},
			},
		},
	}

	requests := make([]*conformancev1alpha1.ClientCompatRequest, len(testCaseData))
	responses := make([]*conformancev1alpha1.ClientCompatResponse, len(testCaseData))
	for i, testCase := range testCaseData {
		requests[i] = proto.Clone(testCase.Request).(*conformancev1alpha1.ClientCompatRequest) //nolint:errcheck,forcetypeassert
		requests[i].Host = svrResponse.Host
		requests[i].Port = svrResponse.Port
		requests[i].ServerTlsCert = svrResponse.PemCert

		if i == 2 {
			responses[i] = &conformancev1alpha1.ClientCompatResponse{
				TestName: testCase.Request.TestName,
				Result: &conformancev1alpha1.ClientCompatResponse_Error{
					Error: &conformancev1alpha1.ClientErrorResult{
						Message: "whoopsy daisy",
					},
				},
			}
		} else {
			responses[i] = &conformancev1alpha1.ClientCompatResponse{
				TestName: testCase.Request.TestName,
				Result: &conformancev1alpha1.ClientCompatResponse_Response{
					Response: testCase.ExpectedResponse,
				},
			}
		}
	}

	testCases := []struct {
		name              string
		isReferenceServer bool
		svrFailsToStart   bool
		svrErrorReader    io.Reader
		clientCloseAfter  int // close client after num responses read
		svrKillAfter      int // kill server process after num requests sent to client

		expectResults map[string]bool
	}{
		{
			name: "normal",
			expectResults: map[string]bool{
				"TestSuite1/testcase1": true,
				"TestSuite1/testcase2": true,
				"TestSuite2/testcase1": false,
				"TestSuite2/testcase2": true,
			},
		},
		{
			name:              "server sends sideband info",
			isReferenceServer: true,
			svrErrorReader: strings.NewReader(strings.Join([]string{
				"TestSuite1/testcase1: server didn't like this request",
				"This line is ignored because it doesn't look right",
				"Blah:Blah/blah: ignored because this isn't a valid test case name",
			}, "\n")),
			expectResults: map[string]bool{
				"TestSuite1/testcase1": false, // error due to sideband info
				"TestSuite1/testcase2": true,
				"TestSuite2/testcase1": false,
				"TestSuite2/testcase2": true,
			},
		},
		{
			name:            "server fails to start",
			svrFailsToStart: true,
			expectResults: map[string]bool{
				"TestSuite1/testcase1": false,
				"TestSuite1/testcase2": false,
				"TestSuite2/testcase1": false,
				"TestSuite2/testcase2": false,
			},
		},
		{
			name:         "server crashes",
			svrKillAfter: 1,
			expectResults: map[string]bool{
				"TestSuite1/testcase1": true,
				"TestSuite1/testcase2": false, // rest fail due to server crash
				"TestSuite2/testcase1": false,
				"TestSuite2/testcase2": false,
			},
		},
		{
			name:             "client crashes",
			clientCloseAfter: 2,
			expectResults: map[string]bool{
				"TestSuite1/testcase1": true,
				"TestSuite1/testcase2": true,
				"TestSuite2/testcase1": false, // rest fail due to client crash
				"TestSuite2/testcase2": false,
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			results := newResults(&knownFailingTrie{})

			var procAddr atomic.Pointer[process] // populated when server process created
			var actualSvrRequest bytes.Buffer
			var svrProcess processStarter
			if testCase.svrFailsToStart {
				svrProcess = newStillbornProcess(&actualSvrRequest, strings.NewReader("oops"), strings.NewReader("oops"))
			} else {
				svrProcess = newFakeProcess(&actualSvrRequest, bytes.NewReader(svrResponseData), testCase.svrErrorReader)
			}
			hookedProcess := func(ctx context.Context, pipeStderr bool) (*process, error) {
				proc, err := svrProcess(ctx, pipeStderr)
				// capture the process when it is created, so we have a way to kill it
				// after a certain amount of request messages are written.
				if err == nil && !procAddr.CompareAndSwap(nil, proc) {
					return nil, errors.New("process already created!?")
				}
				return proc, err
			}

			var client fakeClient
			expectedRequests := requests
			if testCase.svrKillAfter > 0 {
				client.requestHook = func() {
					procAddr.Load().abort()
				}
				client.requestHookCount = testCase.svrKillAfter
				expectedRequests = requests[:testCase.svrKillAfter]
			}

			responsesToSend := responses
			if testCase.clientCloseAfter > 0 {
				if len(expectedRequests) > testCase.clientCloseAfter {
					expectedRequests = expectedRequests[:testCase.clientCloseAfter]
				}
				responsesToSend = responses[:testCase.clientCloseAfter]
			}
			client.responses = make(map[string]*conformancev1alpha1.ClientCompatResponse, len(responsesToSend))
			for _, resp := range responsesToSend {
				client.responses[resp.TestName] = resp
			}

			runTestCasesForServer(
				context.Background(),
				testCase.isReferenceServer,
				svrInstance,
				testCaseData,
				hookedProcess,
				results,
				&client,
			)

			if testCase.svrFailsToStart {
				assert.Empty(t, client.actualRequests)
			} else {
				assert.Empty(t, cmp.Diff(expectedRequests, client.actualRequests, protocmp.Transform()))
			}

			assert.Empty(t, cmp.Diff(expectedSvrReqData, actualSvrRequest.Bytes()))

			actualResults := func() map[string]bool {
				res := map[string]bool{}
				results.mu.Lock()
				defer results.mu.Unlock()
				results.processServerSidebandInfoLocked()
				for name, outcome := range results.outcomes {
					res[name] = outcome.actualFailure == nil
				}
				return res
			}()
			assert.Empty(t, cmp.Diff(testCase.expectResults, actualResults))
		})
	}
}

// fakeProcess is a process starter that represents a fictitious process
// that is runs until the stop method is called.
type fakeProcess struct {
	mu           sync.Mutex
	done         bool
	err          error
	atEndActions []func(error)
}

func newFakeProcess(stdin io.Writer, stdout, stderr io.Reader) processStarter {
	return func(ctx context.Context, pipeStderr bool) (*process, error) {
		proc := &fakeProcess{}
		return &process{
			processController: proc,
			stdin:             &procWriter{w: stdin, proc: proc},
			stdout:            &procReader{r: stdout, proc: proc},
			// Allow stderr to be fully consumed, just in case we try to kill
			// the server process before we've read all sideband info.
			stderr: stderr,
		}, nil
	}
}

func newStillbornProcess(stdin io.Writer, stdout, stderr io.Reader) processStarter {
	return func(ctx context.Context, pipeStderr bool) (*process, error) {
		proc := &fakeProcess{}
		stdout = &hookReader{
			r:    stdout,
			hook: proc.abort,
		}
		return &process{
			processController: proc,
			stdin:             &procWriter{w: stdin, proc: proc},
			stdout:            &procReader{r: stdout, proc: proc},
			stderr:            &procReader{r: stderr, proc: proc},
		}, nil
	}
}

func (f *fakeProcess) stop(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.done {
		return
	}
	f.done = true
	f.err = err
	for _, fn := range f.atEndActions {
		fn(err)
	}
	f.atEndActions = nil
}

func (f *fakeProcess) tryResult() (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.done, f.err
}

func (f *fakeProcess) result() error {
	ch := make(chan struct{})
	f.whenDone(func(_ error) { close(ch) })
	<-ch
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.err
}

func (f *fakeProcess) abort() {
	f.stop(errors.New("process killed by call to abort"))
}

func (f *fakeProcess) whenDone(action func(error)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.done {
		action(f.err)
	}
	f.atEndActions = append(f.atEndActions, action)
}

// procWriter delegates to the given writer but will instead
// immediately return an error if the given process has stopped.
type procWriter struct {
	w      io.Writer
	proc   *fakeProcess
	closed atomic.Bool
}

func (p *procWriter) Write(data []byte) (n int, err error) {
	if p.closed.Load() {
		return 0, errors.New("closed")
	}
	if done, _ := p.proc.tryResult(); done {
		return 0, errors.New("closed")
	}
	return p.w.Write(data)
}

func (p *procWriter) Close() error {
	p.closed.Store(true)
	return nil
}

// procReader delegates to the given reader but will instead
// immediately return an error if the given process has stopped.
type procReader struct {
	r      io.Reader
	proc   *fakeProcess
	closed atomic.Bool
}

func (p *procReader) Read(data []byte) (n int, err error) {
	if p.closed.Load() {
		return 0, io.EOF
	}
	if done, _ := p.proc.tryResult(); done {
		return 0, io.EOF
	}
	return p.r.Read(data)
}

func (p *procReader) Close() error {
	p.closed.Store(true)
	return nil
}

// hookReader calls the given hook function upon reaching EOF.
type hookReader struct {
	r    io.Reader
	hook func()
}

func (h *hookReader) Read(data []byte) (n int, err error) {
	n, err = h.r.Read(data)
	if err != nil && h.hook != nil {
		h.hook()
		// don't run hook 2x
		h.hook = nil
	}
	return n, err
}

// fakeClient immediately calls whenDone when responses that have been prepared
// in the responses field. If requestHookCount is greater than zero and requestHook
// is non-nil, requestHook will be invoked after sendRequest is called requestHookCount
// times.
type fakeClient struct {
	actualRequests   []*conformancev1alpha1.ClientCompatRequest
	responses        map[string]*conformancev1alpha1.ClientCompatResponse
	requestHookCount int
	requestHook      func()
}

func (f *fakeClient) sendRequest(req *conformancev1alpha1.ClientCompatRequest, whenDone func(string, *conformancev1alpha1.ClientCompatResponse, error)) error {
	if len(f.responses) == 0 {
		return errors.New("no more")
	}

	f.actualRequests = append(f.actualRequests, req)
	resp := f.responses[req.TestName]
	delete(f.responses, req.TestName)
	if resp != nil {
		whenDone(req.TestName, resp, nil)
	} else {
		whenDone(req.TestName, nil, errors.New("no configured response"))
	}

	f.requestHookCount--
	if f.requestHookCount == 0 && f.requestHook != nil {
		f.requestHook()
	}

	return nil
}

func (f *fakeClient) closeSend() {
	f.responses = nil
}

func (f *fakeClient) waitForResponses() error {
	return nil
}
