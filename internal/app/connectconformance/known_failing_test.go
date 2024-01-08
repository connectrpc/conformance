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
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseKnownFailing(t *testing.T) {
	t.Parallel()
	trie := parseKnownFailing([]byte(`
		# This is a comment
		This is a test pattern/foo/bar/baz/test-case-name
		All tests in this suite/**
		**/all-test-cases-with-this-name
		Another suite/*/*/*/another-test-case
		A suite with interior double wildcard/**/foo/bar

		# Another comment`))

	testCases := []struct {
		testName string
		matched  bool
	}{
		{
			testName: "This is a test pattern/foo/bar/baz/test-case-name",
			matched:  true,
		},
		{
			testName: "Some test that is not match",
			matched:  false,
		},
		{
			testName: "All tests in this suite",
			matched:  true,
		},
		{
			testName: "All tests in this suite/abc/xyz",
			matched:  true,
		},
		{
			testName: "All tests in this suite/a/b/c/d/e/f/g/h/i/j/k",
			matched:  true,
		},
		{
			testName: "all-test-cases-with-this-name",
			matched:  true,
		},
		{
			testName: "Abc/xyz/all-test-cases-with-this-name",
			matched:  true,
		},
		{
			testName: "Abc/1/2/3/4/5/6/7/8/9/0/all-test-cases-with-this-name",
			matched:  true,
		},
		{
			testName: "Another suite/foo/bar/baz/another-test-case",
			matched:  true,
		},
		{
			testName: "Another suite/foo/bar/baz/buzz/another-test-case",
			matched:  false, // too many path components
		},
		{
			testName: "Another suite/foo/bar/another-test-case",
			matched:  false, // too few path components
		},
		{
			testName: "A suite with interior double wildcard/foo/bar",
			matched:  true,
		},
		{
			testName: "A suite with interior double wildcard/a/1/b/2/c/3/d/4/e/5/foo/bar",
			matched:  true,
		},
		{
			testName: "A suite with interior double wildcard/a/1/b/2/c/3/d/4/e/5/foo",
			matched:  false, // wrong suffix
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()
			matched := trie.match(strings.Split(testCase.testName, "/"))
			require.Equal(t, testCase.matched, matched)
		})
	}
}

func TestKnownFailingTrie_FindUnmatched(t *testing.T) {
	t.Parallel()
	trie := parseKnownFailing([]byte(`
		Unmatched test suite/**
		Simple test suite/that/has/no/wildcards
		All tests in this suite/**
		**/all-test-cases-with-this-name
		**/unmatched-test-case
		Another suite/*/*/*/another-test-case
		A suite with interior double wildcard/**/test-case
		Another unmatched/test/*/with/*/wildcards`))

	testCaseNames := []string{
		"Simple test suite/that/has/no/wildcards",
		"All tests in this suite/foo/bar/case123",
		"All tests in this suite/foo/bar/case456",
		"All tests in this suite/foo/baz/case123",
		"All tests in this suite/foo/baz/case456",
		"Some other test suite/foo/bar/all-test-cases-with-this-name",
		"And another test suite/foo/bar/all-test-cases-with-this-name",
		"Another suite/abc/xyz/123/another-test-case",
		"A suite with interior double wildcard/test-case",
	}
	for _, testCaseName := range testCaseNames {
		require.True(t, trie.match(strings.Split(testCaseName, "/")))
	}

	unmatched := map[string]struct{}{}
	trie.findUnmatched("", unmatched)
	unmatchedSlice := make([]string, 0, len(unmatched))
	for unmatchedName := range unmatched {
		unmatchedSlice = append(unmatchedSlice, unmatchedName)
	}
	sort.Strings(unmatchedSlice)

	require.Equal(t, []string{
		"**/unmatched-test-case",
		"Another unmatched/test/*/with/*/wildcards",
		"Unmatched test suite/**",
	}, unmatchedSlice)
}
