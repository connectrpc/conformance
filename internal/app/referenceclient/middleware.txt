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
)

type WireInterceptor struct {
	Transport http.RoundTripper
}

func (w *WireInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := w.Transport.RoundTrip(req)
	wrapper, ok := req.Context().Value(wireCtxKey{}).(*wireWrapper)
	if err != nil || !ok {
		return resp, err
	}
	resp.Body = &capture{r: resp.Body, resp: resp, wrapper: wrapper}
	return resp, nil
}

type capture struct {
	r       io.ReadCloser
	resp    *http.Response
	wrapper *wireWrapper
}

func (c *capture) Read(p []byte) (int, error) {
	// Capture bytes as they are read
	c.wrapper.buf.Write(p)

	return c.r.Read(p)
}

func (c *capture) Close() error {
	return c.r.Close()
}
