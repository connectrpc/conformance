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
	"errors"
	"io"
	"net/http"
	"testing"

	"connectrpc.com/conformance/internal"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1/conformancev1connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Browser clients under test connect from the test page's origin, so only the
// reference server accepts a cross-origin handshake. A server under test
// keeps connect-go's origin check, which the handshake suite exercises.
func TestCrossOriginWebSocketHandshake(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name           string
		referenceMode  bool
		expectedStatus int
	}{
		{name: "reference-mode", referenceMode: true, expectedStatus: http.StatusSwitchingProtocols},
		{name: "server-under-test", referenceMode: false, expectedStatus: http.StatusForbidden},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			server, _, err := createServer(
				t.Context(),
				&conformancev1.ServerCompatRequest{HttpVersion: conformancev1.HTTPVersion_HTTP_VERSION_1},
				"127.0.0.1:0", "", "",
				testCase.referenceMode,
				internal.NewPrinter(io.Discard),
				nil,
			)
			require.NoError(t, err)
			serveErr := make(chan error, 1)
			go func() { serveErr <- server.Serve() }()
			t.Cleanup(func() {
				require.NoError(t, server.Close())
				if err := <-serveErr; !errors.Is(err, http.ErrServerClosed) {
					assert.NoError(t, err)
				}
			})

			url := "http://" + server.Addr() + conformancev1connect.ConformanceServiceUnaryProcedure
			request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
			require.NoError(t, err)
			request.Header.Set("Connection", "Upgrade")
			request.Header.Set("Upgrade", "websocket")
			request.Header.Set("Sec-WebSocket-Version", "13")
			request.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
			request.Header.Set("Sec-WebSocket-Protocol", "connectrpc.1+json")
			request.Header.Set("Origin", "https://attacker.example")
			response, err := http.DefaultClient.Do(request)
			require.NoError(t, err)
			require.NoError(t, response.Body.Close())
			assert.Equal(t, testCase.expectedStatus, response.StatusCode)
		})
	}
}
