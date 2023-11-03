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
	"strings"
	"testing"

	conformancev1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
)

func TestResults_SetOutcome(t *testing.T) {
	t.Parallel()
	results := newResults(makeKnownFailing())
	results.setOutcome("foo/bar/1", false, nil)
	results.setOutcome("foo/bar/2", true, errors.New("fail"))
	results.setOutcome("foo/bar/3", false, errors.New("fail"))
	results.setOutcome("known-to-fail/1", false, nil)
	results.setOutcome("known-to-fail/2", true, errors.New("fail"))
	results.setOutcome("known-to-fail/3", false, errors.New("fail"))

	logger := &lineWriter{}
	err := results.report(logger)
	require.NoError(t, err)
	require.Len(t, logger.lines, 5)
	require.Equal(t, logger.lines[0], "FAILED: foo/bar/2: fail\n")
	require.Equal(t, logger.lines[1], "FAILED: foo/bar/3: fail\n")
	require.Equal(t, logger.lines[2], "FAILED: known-to-fail/1 was expected to fail but did not\n")
	require.Equal(t, logger.lines[3], "FAILED: known-to-fail/2: fail\n")
	require.Equal(t, logger.lines[4], "INFO: known-to-fail/3 failed (as expected): fail\n")
}

func TestResults_FailedToStart(t *testing.T) {
	t.Parallel()
	results := newResults(makeKnownFailing())
	results.failedToStart([]*conformancev1alpha1.TestCase{
		{Request: &conformancev1alpha1.ClientCompatRequest{TestName: "foo/bar/1"}},
		{Request: &conformancev1alpha1.ClientCompatRequest{TestName: "known-to-fail/1"}},
	}, errors.New("fail"))

	logger := &lineWriter{}
	err := results.report(logger)
	require.NoError(t, err)
	require.Len(t, logger.lines, 2)
	require.Equal(t, logger.lines[0], "FAILED: foo/bar/1: fail\n")
	// Marked as failure even though expected to fail because it failed to start.
	require.Equal(t, logger.lines[1], "FAILED: known-to-fail/1: fail\n")
}

func TestResults_FailRemaining(t *testing.T) {
	t.Parallel()
	results := newResults(makeKnownFailing())
	results.setOutcome("foo/bar/1", false, nil)
	results.setOutcome("known-to-fail/1", false, errors.New("fail"))
	results.failRemaining([]*conformancev1alpha1.TestCase{
		{Request: &conformancev1alpha1.ClientCompatRequest{TestName: "foo/bar/1"}},
		{Request: &conformancev1alpha1.ClientCompatRequest{TestName: "foo/bar/2"}},
		{Request: &conformancev1alpha1.ClientCompatRequest{TestName: "known-to-fail/1"}},
		{Request: &conformancev1alpha1.ClientCompatRequest{TestName: "known-to-fail/2"}},
	}, errors.New("something went wrong"))

	logger := &lineWriter{}
	err := results.report(logger)
	require.NoError(t, err)
	require.Len(t, logger.lines, 3)
	require.Equal(t, logger.lines[0], "FAILED: foo/bar/2: something went wrong\n")
	require.Equal(t, logger.lines[1], "INFO: known-to-fail/1 failed (as expected): fail\n")
	// Marked as failure even though expected to fail because failRemaining is
	// used when a process under test dies (so this error is not due to lack of
	// conformance).
	require.Equal(t, logger.lines[2], "FAILED: known-to-fail/2: something went wrong\n")
}

func TestResults_Failed(t *testing.T) {
	t.Parallel()
	results := newResults(makeKnownFailing())
	results.failed("foo/bar/1", &conformancev1alpha1.ClientErrorResult{Message: "fail"})
	results.failed("known-to-fail/1", &conformancev1alpha1.ClientErrorResult{Message: "fail"})

	logger := &lineWriter{}
	err := results.report(logger)
	require.NoError(t, err)
	require.Len(t, logger.lines, 2)
	require.Equal(t, logger.lines[0], "FAILED: foo/bar/1: fail\n")
	require.Equal(t, logger.lines[1], "INFO: known-to-fail/1 failed (as expected): fail\n")
}

func TestResults_Assert(t *testing.T) {
	t.Parallel()
	results := newResults(makeKnownFailing())
	payload1 := &conformancev1alpha1.ClientResponseResult{
		Payloads: []*conformancev1alpha1.ConformancePayload{
			{Data: []byte{0, 1, 2, 3, 4}},
		},
	}
	payload2 := &conformancev1alpha1.ClientResponseResult{
		Error: &conformancev1alpha1.Error{Code: int32(connect.CodeAborted), Message: "oops"},
	}
	results.assert("foo/bar/1", payload1, payload2)
	results.assert("foo/bar/2", payload2, payload1)
	results.assert("foo/bar/3", payload1, payload1)
	results.assert("foo/bar/4", payload2, payload2)
	results.assert("known-to-fail/1", payload1, payload2)
	results.assert("known-to-fail/2", payload2, payload1)
	results.assert("known-to-fail/3", payload1, payload1)
	results.assert("known-to-fail/4", payload2, payload2)

	logger := &lineWriter{}
	err := results.report(logger)
	require.NoError(t, err)
	// only keep the summary lines:
	var lines []string
	for _, line := range logger.lines {
		if strings.HasPrefix(line, "FAILED: ") || strings.HasPrefix(line, "INFO: ") {
			lines = append(lines, line)
		}
	}
	require.Len(t, lines, 6)
	require.Contains(t, lines[0], "FAILED: foo/bar/1: ")
	require.Contains(t, lines[1], "FAILED: foo/bar/2: ")
	require.Contains(t, lines[2], "INFO: known-to-fail/1 failed (as expected): ")
	require.Contains(t, lines[3], "INFO: known-to-fail/2 failed (as expected): ")
	require.Equal(t, lines[4], "FAILED: known-to-fail/3 was expected to fail but did not\n")
	require.Equal(t, lines[5], "FAILED: known-to-fail/4 was expected to fail but did not\n")
}

func TestResults_ServerSideband(t *testing.T) {
	t.Parallel()
	results := newResults(makeKnownFailing())
	results.setOutcome("foo/bar/1", false, nil)
	results.setOutcome("foo/bar/2", false, errors.New("fail"))
	results.setOutcome("foo/bar/3", false, nil)
	results.setOutcome("known-to-fail/1", false, nil)
	results.setOutcome("known-to-fail/2", false, errors.New("fail"))
	results.recordServerSideband("foo/bar/2", "something awkward in wire format")
	results.recordServerSideband("foo/bar/3", "something awkward in wire format")
	results.recordServerSideband("known-to-fail/1", "something awkward in wire format")

	logger := &lineWriter{}
	err := results.report(logger)
	require.NoError(t, err)
	require.Len(t, logger.lines, 4)
	require.Equal(t, logger.lines[0], "FAILED: foo/bar/2: something awkward in wire format; fail\n")
	require.Equal(t, logger.lines[1], "FAILED: foo/bar/3: something awkward in wire format\n")
	require.Equal(t, logger.lines[2], "INFO: known-to-fail/1 failed (as expected): something awkward in wire format\n")
	require.Equal(t, logger.lines[3], "INFO: known-to-fail/2 failed (as expected): fail\n")
}

func makeKnownFailing() *knownFailingTrie {
	var trie knownFailingTrie
	trie.add([]string{"known-to-fail", "**"})
	return &trie
}

type lineWriter struct {
	current []byte
	lines   []string
}

func (l *lineWriter) Write(data []byte) (n int, err error) {
	for {
		if len(data) == 0 {
			return n, nil
		}
		var hasLF bool
		pos := bytes.IndexByte(data, '\n')
		if pos == -1 {
			pos = len(data)
		} else {
			pos++ // include LF
			hasLF = true
		}
		l.current = append(l.current, data[:pos]...)
		if hasLF {
			l.lines = append(l.lines, string(l.current))
			l.current = nil
		}
		data = data[pos:]
	}
}
