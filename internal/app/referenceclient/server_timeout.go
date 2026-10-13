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
	"net/http"
	"strconv"

	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
)

const (
	connectTimeoutHeader = "Connect-Timeout-Ms"
	grpcTimeoutHeader    = "Grpc-Timeout"

	// The gRPC protocol limits timeout values to 8 digits.
	grpcTimeoutMaxValue = 99_999_999
)

// timeoutHeaderTransport sends the test case's timeout to the server in the
// protocol's timeout header.
//
// In reference mode, the reference client does not set a context deadline
// for the timeout, so that only the server under test can enforce it. But
// connect-go only writes the timeout header from a context deadline, so this
// transport writes it instead.
type timeoutHeaderTransport struct {
	transport http.RoundTripper
	protocol  conformancev1.Protocol
	timeoutMs uint32
}

func (t *timeoutHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var header, value string
	switch t.protocol {
	case conformancev1.Protocol_PROTOCOL_GRPC, conformancev1.Protocol_PROTOCOL_GRPC_WEB:
		header = grpcTimeoutHeader
		if t.timeoutMs <= grpcTimeoutMaxValue {
			value = strconv.FormatUint(uint64(t.timeoutMs), 10) + "m"
		} else {
			value = strconv.FormatUint(uint64(t.timeoutMs)/1000, 10) + "S"
		}
	default:
		header = connectTimeoutHeader
		value = strconv.FormatUint(uint64(t.timeoutMs), 10)
	}
	if req.Header.Get(header) != "" {
		// A raw request specifies its headers exactly, so leave it alone.
		return t.transport.RoundTrip(req)
	}
	// RoundTrippers must not modify the request, so modify a copy.
	req = req.Clone(req.Context())
	req.Header.Set(header, value)
	return t.transport.RoundTrip(req)
}
