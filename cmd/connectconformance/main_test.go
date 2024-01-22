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
