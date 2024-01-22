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

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePatterns(t *testing.T) {
	t.Parallel()
	patterns := parsePatternFile([]byte(`
		# This is a comment
		This is a test pattern/foo/bar/baz/test-case-name
		All tests in this suite/**
		**/all-test-cases-with-this-name
		Another suite/*/*/*/another-test-case
		A suite with interior double wildcard/**/foo/bar

		# Another comment`))
	expectedResult := []string{
		"This is a test pattern/foo/bar/baz/test-case-name",
		"All tests in this suite/**",
		"**/all-test-cases-with-this-name",
		"Another suite/*/*/*/another-test-case",
		"A suite with interior double wildcard/**/foo/bar",
	}
	assert.Equal(t, expectedResult, patterns)
}
