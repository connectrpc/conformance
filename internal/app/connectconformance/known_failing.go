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
	"os"
	"strings"
)

type knownFailingTrie struct {
	present  bool
	matched  int
	children map[string]*knownFailingTrie
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
		k.matched++
		return k.present
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
	child = k.children["**"]
	return child != nil && child.match(nil)
}

func (k *knownFailingTrie) findUnmatched(prefix string, unmatched map[string]struct{}) {
	if k.present && k.matched == 0 {
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

func loadKnownFailing(knownFailingFileName string) (*knownFailingTrie, error) {
	if knownFailingFileName == "" {
		return nil, nil //nolint:nilnil
	}
	data, err := os.ReadFile(knownFailingFileName)
	if err != nil {
		return nil, err
	}
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
	return &knownFailing, nil
}
