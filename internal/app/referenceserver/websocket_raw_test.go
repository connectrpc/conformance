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

package referenceserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckClientClose(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		expected []uint32
		code     int
		reason   string
		problem  string
	}{
		{
			name:     "error with reason",
			expected: []uint32{3102},
			code:     3102,
			reason:   "unexpected marker",
		},
		{
			name:     "error without reason",
			expected: []uint32{3102},
			code:     3102,
			problem:  "client closed the connection with code 3102 but no reason",
		},
		{
			name:     "unexpected code",
			expected: []uint32{3102},
			code:     3103,
			reason:   "body in wrong frame type",
			problem:  "client closed the connection with code 3103; expected one of [3102]",
		},
		{
			// A browser closes with 1002 or 1007 itself, and may give no reason.
			name:     "browser code without reason",
			expected: []uint32{3102, 1007, 1002},
			code:     1007,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, testCase.problem, checkClientClose(testCase.expected, testCase.code, testCase.reason))
		})
	}
}
