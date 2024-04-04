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
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"connectrpc.com/conformance/internal"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
)

const (
	clientResponseTimeout = 20 * time.Second
	maxClientResponseSize = 16 * 1024 * 1024 // 16 MB
)

var (
	errClosed    = errors.New("could not send request: client stdin is closed")
	errDuplicate = errors.New("duplicate test case")
	errNoOutcome = errors.New("no outcome ever received")
)

type clientRunner interface {
	sendRequest(req *conformancev1.ClientCompatRequest, whenDone func(string, *conformancev1.ClientCompatResponse, error)) error
	closeSend()
	waitForResponses() error

	isRunning() bool
	stop()
}

func runClient(ctx context.Context, start processStarter) (clientRunner, error) {
	proc, err := start(ctx, false)
	if err != nil {
		return nil, err
	}
	result := &clientProcessRunner{
		proc:       proc,
		done:       make(chan struct{}),
		pendingOps: map[string]func(string, *conformancev1.ClientCompatResponse, error){},
	}
	proc.whenDone(func(_ error) {
		result.terminated.Store(false)
	})
	go result.consumeOutput()
	return result, nil
}

type clientProcessRunner struct {
	proc       *process
	terminated atomic.Bool

	err  atomic.Pointer[error]
	done chan struct{}

	sendMu     sync.Mutex // serialize calls to sendRequest
	closedSend bool

	// pendingMu needs to be separate so that consumeOutput function can acquire
	// it without being blocked by another goroutine writing a request (which can
	// trivially lead to deadlock).
	// If acquiring both sendMu and pendingMu, *always* acquire sendMu first.
	pendingMu  sync.Mutex
	pendingOps map[string]func(string, *conformancev1.ClientCompatResponse, error)
}

func (c *clientProcessRunner) sendRequest(req *conformancev1.ClientCompatRequest, whenDone func(string, *conformancev1.ClientCompatResponse, error)) (err error) {
	if err := c.err.Load(); err != nil && *err != nil {
		return *err
	}

	c.sendMu.Lock()
	defer c.sendMu.Unlock()

	if c.closedSend {
		return errClosed
	}

	// We have to eagerly add to pending set. If we waited until after
	// a successful write, it's possible that process could process the
	// request and reply to it and consumeOutput goroutine could read
	// the response, all concurrently before we've gotten a chance to
	// add it. That would result in consumeOutput failing due to receiving
	// a response for an unknown test case.
	c.pendingMu.Lock()
	_, exists := c.pendingOps[req.TestName]
	if !exists {
		c.pendingOps[req.TestName] = whenDone
	}
	c.pendingMu.Unlock()
	if exists {
		return fmt.Errorf("%w: %q", errDuplicate, req.TestName)
	}

	if err := internal.WriteDelimitedMessage(c.proc.stdin, req); err != nil {
		// Since we eagerly added to pending set but failed to write,
		// we now need to remove it to clean up.
		c.pendingMu.Lock()
		_, exists := c.pendingOps[req.TestName]
		if exists {
			delete(c.pendingOps, req.TestName)
		}
		c.pendingMu.Unlock()

		if !exists {
			// It was concurrently removed, which means the client *did* get the
			// request and already replied, and consumeOutput already handled it.
			return nil
		}

		if errors.Is(err, io.ErrClosedPipe) {
			err = errClosed
		}
		c.err.CompareAndSwap(nil, &err)
		return err
	}

	return nil
}

func (c *clientProcessRunner) closeSend() {
	c.sendMu.Lock()
	_ = c.proc.stdin.Close()
	c.closedSend = true
	c.sendMu.Unlock()
}

func (c *clientProcessRunner) waitForResponses() error {
	<-c.done

	// Allow process some time to close on its own.
	procErrChan := make(chan error, 1)
	go func() {
		procErrChan <- c.proc.result()
	}()
	var procErr error
	select {
	case procErr = <-procErrChan:
	case <-time.After(3 * time.Second):
		// Not closing fast enough. Let's prod it along
		c.proc.abort()
		procErr = <-procErrChan
	}

	err := c.err.Load()
	if err != nil {
		return *err
	}
	return procErr
}

func (c *clientProcessRunner) isRunning() bool {
	return !c.terminated.Load()
}

func (c *clientProcessRunner) stop() {
	c.proc.abort()
	c.terminated.Store(true)
	_ = c.proc.result() // wait for process to stop
}

func (c *clientProcessRunner) consumeOutput() {
	defer close(c.done)
	var reasonForReturn error
	defer func() {
		if reasonForReturn != nil && !errors.Is(reasonForReturn, io.EOF) {
			c.err.CompareAndSwap(nil, &reasonForReturn)
			c.terminated.Store(true)
			c.proc.abort()
		}
		c.closeSend() // stop the send side now that we're done with receive side

		c.pendingMu.Lock()
		defer c.pendingMu.Unlock()
		for key, action := range c.pendingOps {
			err := reasonForReturn
			if err == nil || errors.Is(err, io.EOF) {
				err = errNoOutcome
			}
			action(key, nil, &failedToGetResultError{err})
			delete(c.pendingOps, key)
		}
	}()

	testCaseNames := map[string]struct{}{}
	for {
		resp := &conformancev1.ClientCompatResponse{}
		readErr := internal.ReadDelimitedMessage(c.proc.stdout, resp, "client", clientResponseTimeout, maxClientResponseSize)
		if readErr != nil {
			reasonForReturn = readErr
			return
		}

		c.pendingMu.Lock()
		action, ok := c.pendingOps[resp.TestName]
		if ok {
			delete(c.pendingOps, resp.TestName)
		}
		c.pendingMu.Unlock()
		if !ok {
			if _, ok := testCaseNames[resp.TestName]; ok {
				// already processed this one
				reasonForReturn = fmt.Errorf("duplicate response received for test case name %q", resp.TestName)
			} else {
				reasonForReturn = fmt.Errorf("received response for unrecognized test case name %q", resp.TestName)
			}
			return
		}
		testCaseNames[resp.TestName] = struct{}{}
		action(resp.TestName, resp, nil)
	}
}

type failedToGetResultError struct {
	err error
}

func (e *failedToGetResultError) Error() string {
	return fmt.Sprintf("failed to get result from client: %v", e.err)
}

func (e *failedToGetResultError) Unwrap() error {
	return e.err
}
