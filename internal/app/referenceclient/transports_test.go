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

package referenceclient

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A request that may cancel must not share an HTTP/1 connection with other
// requests, or its cancellation can fail whichever request reuses the
// connection with a "context canceled" error.
func TestTransportsIsolateHTTP1RequestsThatMayCancel(t *testing.T) {
	t.Parallel()

	var connsMu sync.Mutex
	conns := map[net.Conn]struct{}{}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(respWriter http.ResponseWriter, _ *http.Request) {
		respWriter.WriteHeader(http.StatusOK)
	}))
	server.Config.ConnState = func(conn net.Conn, state http.ConnState) {
		if state == http.StateNew {
			connsMu.Lock()
			defer connsMu.Unlock()
			conns[conn] = struct{}{}
		}
	}
	server.Start()
	t.Cleanup(server.Close)
	host, portString, err := net.SplitHostPort(server.Listener.Addr().String())
	require.NoError(t, err)
	port, err := strconv.ParseUint(portString, 10, 32)
	require.NoError(t, err)

	newCompatRequest := func() *conformancev1.ClientCompatRequest {
		return &conformancev1.ClientCompatRequest{
			HttpVersion: conformancev1.HTTPVersion_HTTP_VERSION_1,
			Host:        host,
			Port:        uint32(port),
		}
	}
	cancelRequest := newCompatRequest()
	cancelRequest.Cancel = &conformancev1.ClientCompatRequest_Cancel{}
	timeoutRequest := newCompatRequest()
	timeoutMs := uint32(1000)
	timeoutRequest.TimeoutMs = &timeoutMs

	var transportCache transports
	send := func(compatRequest *conformancev1.ClientCompatRequest) {
		t.Helper()
		roundTripper, serverURL, err := transportCache.get(compatRequest)
		require.NoError(t, err)
		httpRequest, err := http.NewRequestWithContext(t.Context(), http.MethodPost, serverURL.String(), http.NoBody)
		require.NoError(t, err)
		httpResponse, err := roundTripper.RoundTrip(httpRequest)
		require.NoError(t, err)
		require.NoError(t, httpResponse.Body.Close())
	}
	connCount := func() int {
		connsMu.Lock()
		defer connsMu.Unlock()
		return len(conns)
	}

	send(newCompatRequest())
	send(newCompatRequest())
	require.Equal(t, 1, connCount(), "requests that cannot cancel should share a pooled connection")

	// Each request that may cancel dials its own connection, and the next
	// ordinary request still reuses the original pooled one.
	send(cancelRequest)
	send(newCompatRequest())
	assert.Equal(t, 2, connCount())
	send(timeoutRequest)
	send(newCompatRequest())
	assert.Equal(t, 3, connCount())

	var cachedTransports int
	transportCache.cache.Range(func(any, any) bool {
		cachedTransports++
		return true
	})
	assert.Equal(t, 2, cachedTransports, "want one pooled and one non-pooled transport")
}
