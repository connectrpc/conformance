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
	"io"
	"net/http"

	"connectrpc.com/conformance/internal/tracer"
)

type wireInterceptor struct {
	Transport http.RoundTripper
}

// RoundTrip replaces the response body with a wireReader which captures bytes
// as they are read.
func (w *wireInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := w.Transport.RoundTrip(req)
	wrapper, ok := req.Context().Value(wireCtxKey{}).(*wireWrapper)
	if err != nil || !ok {
		return resp, err
	}
	resp.Body = &wireReader{r: resp.Body, resp: resp, wrapper: wrapper}
	return resp, nil
}

// newWireInterceptor creates a new wireInterceptor which wraps the given transport
// in a TracingRoundTripper.
func newWireInterceptor(transport http.RoundTripper, trace *tracer.Tracer) http.RoundTripper {
	return &wireInterceptor{
		Transport: tracer.TracingRoundTripper(transport, &wireTracer{
			tracer: trace,
		}),
	}
}

type wireReader struct {
	r       io.ReadCloser
	resp    *http.Response
	wrapper *wireWrapper
}

func (w *wireReader) Read(p []byte) (int, error) {
	n, err := w.r.Read(p)

	// Capture bytes as they are read
	w.wrapper.buf.Write(p[:n])

	return n, err
}

func (w *wireReader) Close() error {
	return w.r.Close()
}
