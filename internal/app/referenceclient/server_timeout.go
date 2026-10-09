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
	"errors"
	"net/http"
	"strconv"
	"time"
)

const (
	connectTimeoutHeader = "Connect-Timeout-Ms"
	grpcTimeoutHeader    = "Grpc-Timeout"

	// serverTimeoutGracePeriod is how much longer than the requested timeout
	// the reference client waits for a server under test to enforce it.
	serverTimeoutGracePeriod = 10 * time.Second
)

// errServerIgnoredTimeout is the cause of the reference client's context
// being canceled when the server under test fails to enforce a timeout.
var errServerIgnoredTimeout = errors.New("server did not enforce timeout")

// serverTimeoutTransport overwrites the timeout header computed by
// connect-go with the timeout requested by the test case.
//
// When testing a server, the reference client's context deadline is longer
// than the requested timeout so that the client doesn't report "deadline
// exceeded" on behalf of a server that never enforces the timeout. But
// connect-go derives the timeout header from that context deadline, so this
// transport restores the requested value on the wire.
type serverTimeoutTransport struct {
	transport http.RoundTripper
	timeoutMs uint32
}

func (t *serverTimeoutTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var header, value string
	switch {
	case req.Header.Get(connectTimeoutHeader) != "":
		header, value = connectTimeoutHeader, strconv.FormatUint(uint64(t.timeoutMs), 10)
	case req.Header.Get(grpcTimeoutHeader) != "":
		header, value = grpcTimeoutHeader, strconv.FormatUint(uint64(t.timeoutMs), 10)+"m"
	default:
		return t.transport.RoundTrip(req)
	}
	// RoundTrippers must not modify the request, so modify a copy.
	req = req.Clone(req.Context())
	req.Header.Set(header, value)
	return t.transport.RoundTrip(req)
}
