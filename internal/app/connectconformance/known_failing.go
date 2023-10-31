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
	"strings"
	"sync/atomic"
)

// knownFailingTrie is a trie (aka prefix tree) of patterns of test case
// names that are known to fail.
type knownFailingTrie struct {
	// If true, this node represents a path that was inserted into the trie.
	// If false, this node is an intermediate component of a path.
	present  bool
	children map[string]*knownFailingTrie

	// matched is used to verify that all paths in the trie are valid
	// and correspond to at least one test case
	matched atomic.Int32
}

// parseKnownFailing creates a knownFailingTrie from the given configuration
// data for known failing test cases.
func parseKnownFailing(data []byte) *knownFailingTrie {
	lines := bytes.Split(data, []byte{'\n'})
	var knownFailing knownFailingTrie
	for _, line := range lines {
		line := bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if line[0] == '#' {
			// comment line
			continue
		}
		knownFailing.add(strings.Split(string(line), "/"))
	}
	return &knownFailing
}

func (k *knownFailingTrie) add(components []string) {
	if len(components) == 0 {
		k.present = true
		return
	}
	if k.children == nil {
		k.children = map[string]*knownFailingTrie{}
	}
	first, rest := components[0], components[1:]
	child := k.children[first]
	if child == nil {
		child = &knownFailingTrie{}
		k.children[first] = child
	}
	child.add(rest)
}

func (k *knownFailingTrie) match(components []string) bool {
	if len(components) == 0 {
		if k.present {
			k.matched.Add(1)
			return true
		}
		// See if there's a double-wildcard that may match the empty remaining components.
		child := k.children["**"]
		if child != nil && child.present {
			child.matched.Add(1)
			return true
		}
		return false
	}
	first, rest := components[0], components[1:]
	child := k.children[first]
	if child != nil && child.match(rest) {
		return true
	}
	child = k.children["*"]
	if child != nil && child.match(rest) {
		return true
	}

	// ** can match zero or more components
	child = k.children["**"]
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

func (k *knownFailingTrie) findUnmatched(prefix string, unmatched map[string]struct{}) {
	if k.present && k.matched.Load() == 0 {
		unmatched[prefix] = struct{}{}
	}
	for next, child := range k.children {
		var childPrefix string
		if prefix == "" {
			childPrefix = next
		} else {
			childPrefix = prefix + "/" + next
		}
		child.findUnmatched(childPrefix, unmatched)
	}
}
