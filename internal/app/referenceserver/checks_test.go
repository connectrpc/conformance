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
	"net/http"
	"net/url"
	"testing"

	"connectrpc.com/conformance/internal"
	"github.com/stretchr/testify/assert"
)

// TestCheckConnectGetQueryParamOrder tests the check itself.
func TestCheckConnectGetQueryParamOrder(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		method        string
		rawQuery      string
		expectedError string
	}{
		{
			name:     "non-GET request is ignored",
			method:   http.MethodPost,
			rawQuery: "message=foo&encoding=proto&connect=v1",
		},
		{
			name:     "empty query",
			method:   http.MethodGet,
			rawQuery: "",
		},
		{
			name:     "encoding and message only, correct order",
			method:   http.MethodGet,
			rawQuery: "encoding=proto&message=AAAA",
		},
		{
			name:     "all five parameters in spec order",
			method:   http.MethodGet,
			rawQuery: "connect=v1&base64=1&compression=gzip&encoding=proto&message=AAAA",
		},
		{
			name:     "version plus encoding plus message",
			method:   http.MethodGet,
			rawQuery: "connect=v1&encoding=proto&message=AAAA",
		},
		{
			name:          "alphabetical order (connect-go default) flags",
			method:        http.MethodGet,
			rawQuery:      "base64=1&compression=gzip&connect=v1&encoding=proto&message=AAAA",
			expectedError: "got [base64, compression, connect, encoding, message]",
		},
		{
			name:          "message before encoding flags",
			method:        http.MethodGet,
			rawQuery:      "message=AAAA&encoding=proto",
			expectedError: "got [message, encoding]",
		},
		{
			name:     "unknown parameters are ignored",
			method:   http.MethodGet,
			rawQuery: "x-trace=abc&connect=v1&extra=1&encoding=proto&message=AAAA",
		},
		{
			name:          "unknown parameters do not mask misorder",
			method:        http.MethodGet,
			rawQuery:      "encoding=proto&x-trace=abc&connect=v1&message=AAAA",
			expectedError: "got [encoding, connect, message]",
		},
		{
			name:     "bare parameter name without value",
			method:   http.MethodGet,
			rawQuery: "connect&encoding=proto&message=AAAA",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			printer := &internal.SimplePrinter{}
			feedback := &feedbackPrinter{p: printer, testCaseName: testCase.name}
			req := &http.Request{
				Method: testCase.method,
				URL:    &url.URL{RawQuery: testCase.rawQuery},
			}
			checkConnectGetQueryParamOrder(req, feedback)
			if testCase.expectedError == "" {
				assert.Empty(t, printer.Messages)
				return
			}
			if assert.Len(t, printer.Messages, 1) {
				assert.Contains(t, printer.Messages[0], testCase.expectedError)
			}
		})
	}
}
