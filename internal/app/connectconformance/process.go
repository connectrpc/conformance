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
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
)

const (
	gracefulShutdownPeriod = 5 * time.Second
)

// process represents some asynchronous execution unit. It may be in another
// OS process or it may actually be another goroutine in the current OS process.
type process struct {
	processController

	stdin  io.WriteCloser
	stdout io.Reader
	stderr io.Reader
}

type processController interface {
	// result returns nil if the process exits normally and an error otherwise.
	// If the process is a separate OS process that exits with a non-zero code,
	// this will return an instance of *[exec.ExitError].
	result() error
	abort()
	whenDone(func(error))
}

type processStarter func(ctx context.Context, pipeStderr bool) (*process, error)

// runCommand returns a process starter that invokes the given command-line in
// a separate OS process.
func runCommand(command []string) processStarter {
	return makeProcess(func(ctx context.Context, stdin io.ReadCloser, stdout, stderr io.WriteCloser) (processController, error) {
		ctx, cancel := context.WithCancel(ctx)
		cmd := exec.CommandContext(ctx, command[0], command[1:]...) //nolint:gosec
		cmd.Stdin = stdin
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		cmd.Cancel = func() error {
			err := cmd.Process.Signal(syscall.SIGTERM)
			if err != nil {
				// Signals like above are not supported on Windows. So if signal fails, try killing.
				err = cmd.Process.Kill()
			}
			return err
		}
		cmd.WaitDelay = gracefulShutdownPeriod
		if err := cmd.Start(); err != nil {
			cancel()
			return nil, err
		}
		cmdProc := &cmdProcess{
			cmd:    cmd,
			cancel: cancel,
			done:   make(chan struct{}),
		}
		go func() {
			// cmd.Wait can only be called once. So we call it from this goroutine
			// and then publish the result so it can be read via cmdProc.result()
			defer close(cmdProc.done)
			cmdProc.cmdResult = cmd.Wait()
			// Also close pipes when the process exits, just so any goroutines that
			// are blocked reading/writing can wake up and observe EOF.
			if stdin != os.Stdin {
				_ = stdin.Close()
			}
			if stdout != os.Stdout {
				_ = stdout.Close()
			}
			if stderr != os.Stderr {
				_ = stderr.Close()
			}
		}()
		return cmdProc, nil
	})
}

// runInProcess returns a process starter that invokes the given function
// in another goroutine.
func runInProcess(args []string, impl func(ctx context.Context, args []string, in io.ReadCloser, out, err io.WriteCloser) error) processStarter {
	return makeProcess(func(ctx context.Context, stdin io.ReadCloser, stdout, stderr io.WriteCloser) (processController, error) {
		ctx, cancel := context.WithCancel(ctx)
		proc := &localProcess{
			cancel: cancel,
			done:   make(chan struct{}),
		}
		if stdin == nil {
			stdin = os.Stdin
		}
		if stdout == nil {
			stdout = os.Stdout
		}
		if stderr == nil {
			stderr = os.Stderr
		}
		go func() {
			defer close(proc.done)
			defer func() {
				if stdin != os.Stdin {
					_ = stdin.Close()
				}
				if stdout != os.Stdout {
					_ = stdout.Close()
				}
				if stderr != os.Stderr {
					_ = stderr.Close()
				}
			}()
			proc.err = impl(ctx, args, stdin, stdout, stderr)
			if proc.err != nil {
				_, _ = fmt.Fprintf(stderr, "%v\n", proc.err)
			}
		}()
		return proc, nil
	})
}

func makeProcess(procFunc func(ctx context.Context, stdin io.ReadCloser, stdout, stderr io.WriteCloser) (processController, error)) processStarter {
	return func(ctx context.Context, pipeStderr bool) (*process, error) {
		stdinReader, stdinWriter := io.Pipe()
		stdoutReader, stdoutWriter := io.Pipe()
		var stderrReader io.ReadCloser
		var stderrWriter io.WriteCloser
		if pipeStderr {
			stderrReader, stderrWriter = io.Pipe()
		} else {
			stderrReader = io.NopCloser(bytes.NewReader(nil)) // empty
			stderrWriter = os.Stderr
		}
		proc, err := procFunc(ctx, stdinReader, stdoutWriter, stderrWriter)
		if err != nil {
			return nil, err
		}
		return &process{
			processController: proc,
			stdin:             stdinWriter,
			stdout:            stdoutReader,
			stderr:            stderrReader,
		}, nil
	}
}

type cmdProcess struct {
	cmd       *exec.Cmd
	cancel    context.CancelFunc
	done      chan struct{}
	cmdResult error
}

func (c *cmdProcess) result() error {
	<-c.done
	return c.cmdResult
}

func (c *cmdProcess) abort() {
	c.cancel()
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
