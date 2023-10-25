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
	"io"
	"math"
	"os"
	"os/exec"
	"strings"
	"time"

	conformancev1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

const (
	gracefulShutdownPeriod = 5 * time.Second
	testCaseTimeout        = 20 * time.Second
)

type process interface {
	result() error
	abort()
	whenDone(func(error))
}

type processStarter func(ctx context.Context, in io.ReadCloser, out, err io.WriteCloser) (process, error)
type action func(ctx context.Context, args []string, in io.ReadCloser, out, err io.WriteCloser) error

func runCommand(command []string) processStarter {
	return func(ctx context.Context, in io.ReadCloser, out, err io.WriteCloser) (process, error) {
		cmd := exec.CommandContext(ctx, command[0], command[1:]...) //nolint:gosec
		cmd.Stdin = in
		cmd.Stdout = out
		cmd.Stderr = err
		cmd.Cancel = func() error {
			err := cmd.Process.Signal(os.Interrupt)
			if err != nil {
				// Interrupt not supported on Windows. If interrupt fails, try sending kill.
				err = cmd.Process.Signal(os.Kill)
			}
			return err
		}
		cmd.WaitDelay = gracefulShutdownPeriod
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		return (*cmdProcess)(cmd), nil
	}
}

func runInProcess(impl action) processStarter {
	return func(ctx context.Context, stdin io.ReadCloser, stdout, stderr io.WriteCloser) (process, error) {
		ctx, cancel := context.WithCancel(ctx)
		proc := &localProcess{
			cancel: cancel,
			done:   make(chan struct{}),
		}
		go func() {
			defer close(proc.done)
			defer func() {
				if stdin != nil {
					_ = stdin.Close()
				}
				if stdout != nil {
					_ = stdout.Close()
				}
				if stderr != nil {
					_ = stderr.Close()
				}
			}()
			proc.err = impl(ctx, nil, stdin, stdout, stderr)
		}()
		return proc, nil
	}
}

type cmdProcess exec.Cmd

func (c *cmdProcess) result() error {
	return (*exec.Cmd)(c).Wait()
}

func (c *cmdProcess) abort() {
	cmd := (*exec.Cmd)(c)
	err := cmd.Cancel()
	if err == nil || errors.Is(err, os.ErrProcessDone) {
		// done
		return
	}
	// Failed to cancel? Try to kill the process.
	_ = cmd.Process.Kill()
}

func (c *cmdProcess) whenDone(action func(error)) {
	go func() {
		action(c.result())
	}()
}

type localProcess struct {
	cancel context.CancelFunc
	done   chan struct{}
	err    error
}

func (l *localProcess) result() error {
	select {
	case <-l.done:
		return l.err
	case <-time.After(gracefulShutdownPeriod):
		return context.DeadlineExceeded
	}
}

func (l *localProcess) abort() {
	l.cancel()
}

func (l *localProcess) whenDone(action func(error)) {
	go func() {
		<-l.done
		action(l.err)
	}()
}

func runTestCasesForServer(
	ctx context.Context,
	isReferenceServer bool,
	cancel func(error),
	meta serverInstance,
	testCases []*conformancev1alpha1.TestCase,
	startServer processStarter,
	requestWriter io.Writer,
	responseReader io.Reader,
	results *testResults,
) bool {
	expectations := make(map[string]*conformancev1alpha1.ClientResponseResult, len(testCases))
	for _, testCase := range testCases {
		expectations[testCase.Request.TestName] = testCase.ExpectedResponse
	}

	serverInputReader, serverInputWriter := io.Pipe()
	serverOutputReader, serverOutputWriter := io.Pipe()
	var serverErrorReader io.ReadCloser
	var serverErrorWriter io.WriteCloser
	if isReferenceServer {
		serverErrorReader, serverErrorWriter = io.Pipe()
	}
	procCtx, procCancel := context.WithCancel(ctx)
	defer procCancel()
	serverProcess, err := startServer(procCtx, serverInputReader, serverOutputWriter, serverErrorWriter)
	if err != nil {
		results.failedToStart(testCases, fmt.Errorf("error starting server: %w", err))
		return ctx.Err() != nil
	}
	defer serverProcess.abort()
	serverProcess.whenDone(func(_ error) {
		procCancel()
	})

	var refServerDone chan struct{}
	if isReferenceServer { //nolint:nestif
		refServerDone = make(chan struct{})
		go func() {
			defer close(refServerDone)
			r := bufio.NewReader(serverErrorReader)
			for {
				str, err := r.ReadString('\n')
				str = strings.TrimSpace(str)
				if str != "" {
					parts := strings.SplitN(str, ":", 2)
					if len(parts) == 2 {
						if _, ok := expectations[parts[0]]; ok {
							// appears to be valid message in the form "test case: error message"
							results.recordServerSideband(parts[0], parts[1])
						}
					}
				}
				if err != nil {
					return
				}
			}
		}()
	}

	// Write request
	serverRequestData, err := proto.Marshal(&conformancev1alpha1.ServerCompatRequest{
		Protocol:    meta.protocol,
		HttpVersion: meta.httpVersion,
		UseTls:      meta.useTLS,
	})
	if err != nil {
		results.failedToStart(testCases, fmt.Errorf("error writing server request: %w", err))
		return ctx.Err() != nil
	}
	if _, err := serverInputWriter.Write(serverRequestData); err != nil {
		results.failedToStart(testCases, fmt.Errorf("error writing server request: %w", err))
		return ctx.Err() != nil
	}
	if err := serverInputWriter.Close(); err != nil {
		results.failedToStart(testCases, fmt.Errorf("error writing server request: %w", err))
		return ctx.Err() != nil
	}

	// Read response
	serverResponseData, err := io.ReadAll(serverOutputReader)
	if err != nil {
		results.failedToStart(testCases, fmt.Errorf("error reading server response: %w", err))
		return ctx.Err() != nil
	}
	if err := serverInputWriter.Close(); err != nil {
		results.failedToStart(testCases, fmt.Errorf("error reading server response: %w", err))
		return ctx.Err() != nil
	}
	var resp conformancev1alpha1.ServerCompatResponse
	if err := proto.Unmarshal(serverResponseData, &resp); err != nil {
		results.failedToStart(testCases, fmt.Errorf("error reading server response: %w", err))
		return ctx.Err() != nil
	}

	// Now we can write all the test cases to the client.
	// We'll spin up a goroutine to write requests and another to read responses.
	clientFinished := make(chan struct{})
	go func() {
		for i, testCase := range testCases {
			if procCtx.Err() != nil {
				// server crashed: mark remaining tests
				err := errors.New("server process terminated unexpectedly")
				for j := i; j < len(testCases); j++ {
					results.setOutcome(testCases[j].Request.TestName, true, err)
				}
				return
			}
			req := testCase.Request
			req.Host = resp.Host
			req.Port = resp.Port
			req.ServerTlsCert = resp.PemCert
			reqData, err := proto.Marshal(req)
			if err != nil {
				results.setOutcome(req.TestName, true, err)
				continue
			}
			n, err := requestWriter.Write(reqData)
			if n < len(reqData) {
				// client pipe broken: mark remaining tests, including this one as failed
				for j := i; j < len(testCases); j++ {
					results.setOutcome(testCases[j].Request.TestName, true, err)
				}
				return
			}
		}
	}()

	var responseErr error
	go func() {
		defer close(clientFinished)
		pending := len(testCases)
		for {
			if pending == 0 {
				return
			}

			// Read varint length
			val, err := readTestCaseResponseLen(ctx, responseReader)
			if err != nil {
				responseErr = err
				return
			}
			buf := make([]byte, val)
			_, err = io.ReadFull(responseReader, buf)
			if err != nil {
				responseErr = err
				return
			}

			var resp conformancev1alpha1.ClientCompatResponse
			err = proto.Unmarshal(buf, &resp)
			switch {
			case err != nil:
				// Oof, can't even read test name. We'll wait until the end and any
				// test cases that have no outcome can be marked as "no response
				// received from client".
			case resp.GetError() != nil:
				results.failed(resp.TestName, resp.GetError())
			default:
				results.assert(resp.TestName, expectations[resp.TestName], resp.GetResponse())
			}

			pending--
		}
	}()

	<-clientFinished
	serverProcess.abort()
	_ = serverProcess.result() // wait for server process to end
	if isReferenceServer {
		<-refServerDone
		results.processServerSidebandInfo()
	}

	// if there are any tests without outcomes, mark them
	results.failRemaining(testCases, errors.New("no outcome received from the client"))

	if responseErr != nil {
		cancel(responseErr)
	}
	return ctx.Err() != nil
}

func readTestCaseResponseLen(ctx context.Context, responseIn io.Reader) (uint32, error) {
	ctx, cancel := context.WithTimeout(ctx, testCaseTimeout)
	defer cancel()

	var result uint32
	var readErr error
	readDone := make(chan struct{})

	go func() {
		defer close(readDone)
		var buf [10]byte // capacious enough for maximum-length varint
		for n := 0; n < len(buf); n++ {
			// read one byte at a time until we encounter end varint
			numRead, err := responseIn.Read(buf[n : n+1])
			if numRead == 0 {
				readErr = err
				return
			}
			if buf[numRead]&0x80 == 0 {
				// high bit clear? this is the last byte of varint
				varint, length := protowire.ConsumeVarint(buf[:numRead+1])
				if length < 0 {
					readErr = errors.New("invalid varint-encoded length")
					return
				}
				if varint > math.MaxUint32 {
					readErr = fmt.Errorf("varint-encoded length out of range: %d > %d", varint, math.MaxUint32)
					return
				}
				result = uint32(varint)
				return
			}
			if err != nil {
				readErr = err
				return
			}
		}
		// consumed 16 bytes and never encountered end of varint? can't be valid
		readErr = errors.New("invalid varint-encoded length")
	}()

	select {
	case <-readDone:
		return result, readErr
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return 0, errors.New("test timed out: client too long to write outcome")
		}
		return 0, ctx.Err()
	}
}
