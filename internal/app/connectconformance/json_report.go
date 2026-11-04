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
)

type outcomeKind int

const (
	_ outcomeKind = iota
	outcomeKindNotRun
	outcomeKindFailed
	outcomeKindPassedUnexpectedly
	outcomeKindFailedAsExpected
	outcomeKindPassed
)

// ErrUnknownOutcomeKind is returned for an outcome kind with no text form.
var ErrUnknownOutcomeKind = errors.New("unknown outcome kind")

// String returns the kind's name in the JSON report, or "" if it has none.
func (kind outcomeKind) String() string {
	switch kind {
	case outcomeKindNotRun:
		return "not-run"
	case outcomeKindFailed:
		return "failed"
	case outcomeKindPassedUnexpectedly:
		return "passed-unexpectedly"
	case outcomeKindFailedAsExpected:
		return "failed-as-expected"
	case outcomeKindPassed:
		return "passed"
	default:
		return ""
	}
}

func (kind outcomeKind) MarshalText() ([]byte, error) {
	text := kind.String()
	if text == "" {
		return nil, ErrUnknownOutcomeKind
	}
	return []byte(text), nil
}

func (kind *outcomeKind) UnmarshalText(text []byte) error {
	for candidate := outcomeKindNotRun; candidate <= outcomeKindPassed; candidate++ {
		if string(text) == candidate.String() {
			*kind = candidate
			return nil
		}
	}
	return ErrUnknownOutcomeKind
}

const (
	patternKindKnownFailing = "known-failing"
	patternKindKnownFlaky   = "known-flaky"
	patternKindRun          = "run"
	patternKindSkip         = "skip"
)

// jsonReport is the file written for the --report-json flag, for tools that
// need per-test results instead of the console output.
type jsonReport struct {
	Error             string                 `json:"error,omitempty"`
	UnmatchedPatterns *jsonUnmatchedPatterns `json:"unmatchedPatterns,omitempty"`
	Tests             []jsonTestResult       `json:"tests"`
}

type jsonUnmatchedPatterns struct {
	Kind     string   `json:"kind"`
	Patterns []string `json:"patterns"`
}

type jsonTestResult struct {
	Name    string      `json:"name"`
	Outcome outcomeKind `json:"outcome"`
}

// writeReport writes a JSON report to path, if it isn't empty. results is nil
// when the run stopped before any test case ran.
func writeReport(path string, results *testResults, runErr error) error {
	if path == "" {
		return nil
	}
	report := jsonReport{Tests: []jsonTestResult{}}
	if runErr != nil {
		report.Error = runErr.Error()
		if unmatchedErr, ok := errors.AsType[*unmatchedPatternsError](runErr); ok {
			report.UnmatchedPatterns = &jsonUnmatchedPatterns{
				Kind:     unmatchedErr.kind,
				Patterns: unmatchedErr.patterns,
			}
		}
	}
	if results != nil {
		report.Tests = results.jsonTestResults()
	}
	data, err := json.MarshalIndent(&report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	return nil
}
