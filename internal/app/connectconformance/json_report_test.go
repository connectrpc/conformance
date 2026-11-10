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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"connectrpc.com/conformance/internal"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteReport(t *testing.T) {
	t.Parallel()

	t.Run("no path", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, writeReport("", nil, errors.New("ignored")))
	})

	// The expected output is literal JSON because the workflows in
	// .github/impl-conformance parse it.
	t.Run("test outcomes", func(t *testing.T) {
		t.Parallel()
		results := newResults(conformancev1.TestSuite_TEST_MODE_UNSPECIFIED, 0, makeKnownFailing(), makeKnownFlaky(), nil)
		results.setOutcome("foo/bar/passes", false, nil)
		results.setOutcome("foo/bar/fails", false, errors.New("ruh roh"))
		results.setOutcome("foo/bar/setup-error", true, errors.New("ruh roh"))
		results.setOutcome("foo/bar/not-run", true, &couldNotRunError{errors.New("client exited")})
		results.setOutcome("known-to-fail/fails", false, errors.New("ruh roh"))
		results.setOutcome("known-to-fail/passes", false, nil)
		results.setOutcome("known-to-flake/fails", false, errors.New("ruh roh"))
		results.setOutcome("known-to-flake/passes", false, nil)
		results.report(&internal.SimplePrinter{})

		assert.JSONEq(t, `{
			"tests": [
				{"name": "foo/bar/fails", "outcome": "failed"},
				{"name": "foo/bar/not-run", "outcome": "not-run"},
				{"name": "foo/bar/passes", "outcome": "passed"},
				{"name": "foo/bar/setup-error", "outcome": "failed"},
				{"name": "known-to-fail/fails", "outcome": "failed-as-expected"},
				{"name": "known-to-fail/passes", "outcome": "passed-unexpectedly"},
				{"name": "known-to-flake/fails", "outcome": "failed-as-expected"},
				{"name": "known-to-flake/passes", "outcome": "passed"}
			]
		}`, writeReportFile(t, results, nil))
	})

	t.Run("unmatched patterns", func(t *testing.T) {
		t.Parallel()
		_, matchErr := tryMatchPatterns(
			"run patterns",
			patternKindRun,
			parsePatterns([]string{"foo/**", "missing/**"}),
			[]*conformancev1.TestCase{{Request: &conformancev1.ClientCompatRequest{TestName: "foo/bar"}}},
		)
		require.Error(t, matchErr)

		assert.JSONEq(t, `{
			"error": "wrapped: run patterns: unmatched and possibly invalid patterns:\nmissing/**",
			"unmatchedPatterns": {"kind": "run", "patterns": ["missing/**"]},
			"tests": []
		}`, writeReportFile(t, nil, fmt.Errorf("wrapped: %w", matchErr)))
	})

	t.Run("other error", func(t *testing.T) {
		t.Parallel()
		assert.JSONEq(t, `{"error": "ruh roh", "tests": []}`, writeReportFile(t, nil, errors.New("ruh roh")))
	})

	t.Run("unwritable path", func(t *testing.T) {
		t.Parallel()
		err := writeReport(filepath.Join(t.TempDir(), "missing", "report.json"), nil, nil)
		require.ErrorContains(t, err, "write report")
	})
}

func TestOutcomeKindText(t *testing.T) {
	t.Parallel()

	for kind, text := range map[outcomeKind]string{
		outcomeKindNotRun:             "not-run",
		outcomeKindFailed:             "failed",
		outcomeKindPassedUnexpectedly: "passed-unexpectedly",
		outcomeKindFailedAsExpected:   "failed-as-expected",
		outcomeKindPassed:             "passed",
	} {
		marshaled, err := kind.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, text, string(marshaled))
		var unmarshaled outcomeKind
		require.NoError(t, unmarshaled.UnmarshalText([]byte(text)))
		assert.Equal(t, kind, unmarshaled)
	}

	for _, kind := range []outcomeKind{0, outcomeKindPassed + 1} {
		_, err := kind.MarshalText()
		require.ErrorIs(t, err, ErrUnknownOutcomeKind)
	}
	for _, text := range []string{"", "unknown"} {
		var unmarshaled outcomeKind
		require.ErrorIs(t, unmarshaled.UnmarshalText([]byte(text)), ErrUnknownOutcomeKind)
	}
}

func TestRunWritesReport(t *testing.T) {
	t.Parallel()

	configFile := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte(`features:
  versions: [HTTP_VERSION_1]
  protocols: [PROTOCOL_CONNECT]
  codecs: [CODEC_PROTO]
  compressions: [COMPRESSION_IDENTITY]
  streamTypes: [STREAM_TYPE_UNARY]
  supportsTls: false
`), 0o600))
	const runPattern = "Basic/**/unary/success"
	newFlags := func(t *testing.T) *Flags {
		t.Helper()
		return &Flags{
			ConfigFile:  configFile,
			RunPatterns: []string{runPattern},
			MaxServers:  2,
			Parallelism: 4,
			ServerBind:  "127.0.0.1",
			ReportFile:  filepath.Join(t.TempDir(), "report.json"),
		}
	}

	t.Run("passed", func(t *testing.T) {
		t.Parallel()
		flags := newFlags(t)
		succeeded, err := Run(flags, &testPrinter{t}, &testPrinter{t})
		require.NoError(t, err)
		assert.True(t, succeeded)
		report := readReport(t, flags.ReportFile)
		assert.Empty(t, report.Error)
		assertOutcomes(t, report, outcomeKindPassed)
	})

	t.Run("passed unexpectedly", func(t *testing.T) {
		t.Parallel()
		flags := newFlags(t)
		flags.KnownFailingPatterns = []string{runPattern}
		succeeded, err := Run(flags, &testPrinter{t}, &testPrinter{t})
		require.NoError(t, err)
		assert.False(t, succeeded)
		assertOutcomes(t, readReport(t, flags.ReportFile), outcomeKindPassedUnexpectedly)
	})

	t.Run("unmatched pattern", func(t *testing.T) {
		t.Parallel()
		flags := newFlags(t)
		flags.KnownFlakyPatterns = []string{"Missing/**"}
		succeeded, err := Run(flags, &testPrinter{t}, &testPrinter{t})
		require.Error(t, err)
		assert.False(t, succeeded)
		report := readReport(t, flags.ReportFile)
		assert.Equal(t, err.Error(), report.Error)
		assert.Equal(t, &jsonUnmatchedPatterns{Kind: patternKindKnownFlaky, Patterns: []string{"Missing/**"}}, report.UnmatchedPatterns)
		assert.Empty(t, report.Tests)
	})

	// A run whose tests pass must still fail if its report can't be written.
	t.Run("unwritable report", func(t *testing.T) {
		t.Parallel()
		flags := newFlags(t)
		flags.ReportFile = filepath.Join(t.TempDir(), "missing", "report.json")
		succeeded, err := Run(flags, &testPrinter{t}, &testPrinter{t})
		require.ErrorContains(t, err, "write report")
		assert.False(t, succeeded)
	})
}

func writeReportFile(t *testing.T, results *testResults, runErr error) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "report.json")
	require.NoError(t, writeReport(path, results, runErr))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}

func readReport(t *testing.T, path string) *jsonReport {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var report jsonReport
	require.NoError(t, json.Unmarshal(data, &report))
	return &report
}

func assertOutcomes(t *testing.T, report *jsonReport, want outcomeKind) {
	t.Helper()
	require.NotEmpty(t, report.Tests)
	for _, result := range report.Tests {
		assert.True(t, strings.HasPrefix(result.Name, "Basic/"), result.Name)
		assert.True(t, strings.HasSuffix(result.Name, "/unary/success"), result.Name)
		assert.Equal(t, want, result.Outcome, result.Name)
	}
}
