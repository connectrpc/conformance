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

package internal

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsWebSocketUpgrade(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name       string
		method     string
		connection []string
		upgrade    []string
		expected   bool
	}{
		{
			name:       "plain",
			method:     http.MethodGet,
			connection: []string{"Upgrade"},
			upgrade:    []string{"websocket"},
			expected:   true,
		},
		{
			name:       "case-insensitive",
			method:     http.MethodGet,
			connection: []string{"UPGRADE"},
			upgrade:    []string{"WebSocket"},
			expected:   true,
		},
		{
			name:       "among-other-tokens",
			method:     http.MethodGet,
			connection: []string{"keep-alive, Upgrade"},
			upgrade:    []string{"h2c ,websocket"},
			expected:   true,
		},
		{
			name:       "token-in-second-field-line",
			method:     http.MethodGet,
			connection: []string{"keep-alive", "Upgrade"},
			upgrade:    []string{"websocket"},
			expected:   true,
		},
		{
			name:       "not-get",
			method:     http.MethodPost,
			connection: []string{"Upgrade"},
			upgrade:    []string{"websocket"},
		},
		{
			name:    "no-connection-upgrade",
			method:  http.MethodGet,
			upgrade: []string{"websocket"},
		},
		{
			name:       "other-upgrade",
			method:     http.MethodGet,
			connection: []string{"Upgrade"},
			upgrade:    []string{"h2c"},
		},
		{
			name:       "token-substring",
			method:     http.MethodGet,
			connection: []string{"upgraded"},
			upgrade:    []string{"websocket"},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			request := &http.Request{Method: testCase.method, Header: http.Header{}}
			for _, value := range testCase.connection {
				request.Header.Add("Connection", value)
			}
			for _, value := range testCase.upgrade {
				request.Header.Add("Upgrade", value)
			}
			assert.Equal(t, testCase.expected, IsWebSocketUpgrade(request))
		})
	}
}
