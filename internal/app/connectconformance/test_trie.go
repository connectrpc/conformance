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
	"strings"
	"sync/atomic"
)

// testTrie is a trie (aka prefix tree) of patterns of test case
// names that are known to fail.
type testTrie struct {
	// If true, this node represents a path that was inserted into the trie.
	// If false, this node is an intermediate component of a path.
	present  bool
	children map[string]*testTrie

	// matched is used to verify that all paths in the trie are valid
	// and correspond to at least one test case
	matched atomic.Int32
}

func parsePatterns(patterns []string) *testTrie {
	if len(patterns) == 0 {
		return nil
	}
	var result testTrie
	for _, pattern := range patterns {
		result.addPattern(pattern)
	}
	return &result
}

func (tt *testTrie) addPattern(pattern string) {
	tt.add(strings.Split(pattern, "/"))
}

func (tt *testTrie) add(components []string) {
	if len(components) == 0 {
		tt.present = true
		return
	}
	if tt.children == nil {
		tt.children = map[string]*testTrie{}
	}
	first, rest := components[0], components[1:]
	child := tt.children[first]
	if child == nil {
		child = &testTrie{}
		tt.children[first] = child
	}
	child.add(rest)
}

func (tt *testTrie) matchPattern(pattern string) bool {
	return tt.match(strings.Split(pattern, "/"))
}

func (tt *testTrie) match(components []string) bool {
	if len(components) == 0 {
		if tt.present {
			tt.matched.Add(1)
			return true
		}
		// See if there's a double-wildcard that may match the empty remaining components.
		child := tt.children["**"]
		if child != nil && child.present {
			child.matched.Add(1)
			return true
		}
		return false
	}
	first, rest := components[0], components[1:]
	child := tt.children[first]
	if child != nil && child.match(rest) {
		return true
	}
	child = tt.children["*"]
	if child != nil && child.match(rest) {
		return true
	}

	// ** can match zero or more components
	child = tt.children["**"]
	if child == nil {
		return false
	}
	for {
		if child.match(components) {
			return true
		}
		if len(components) == 0 {
			return child.present
		}
		components = components[1:]
	}
}

func (tt *testTrie) allUnmatched() map[string]struct{} {
	unmatched := map[string]struct{}{}
	tt.findUnmatched("", unmatched)
	return unmatched
}

func (tt *testTrie) findUnmatched(prefix string, unmatched map[string]struct{}) {
	if tt.present && tt.matched.Load() == 0 {
		unmatched[prefix] = struct{}{}
	}
	for next, child := range tt.children {
		var childPrefix string
		if prefix == "" {
			childPrefix = next
		} else {
			childPrefix = prefix + "/" + next
		}
		child.findUnmatched(childPrefix, unmatched)
	}
}

func (tt *testTrie) length() int {
	var result int
	if tt.present {
		result++
	}
	for _, child := range tt.children {
		result += child.length()
	}
	return result
}
