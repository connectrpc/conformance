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

package referenceclient

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"connectrpc.com/conformance/internal"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
)

type rawRequestSender struct {
	transport  http.RoundTripper
	rawRequest *conformancev1.RawHTTPRequest
}

func (r *rawRequestSender) RoundTrip(orig *http.Request) (*http.Response, error) {
	switch r.rawRequest.Body.(type) {
	case *conformancev1.RawHTTPRequest_Unary, *conformancev1.RawHTTPRequest_Stream:
	default:
		return nil, fmt.Errorf("raw request has invalid body definition: %T", r.rawRequest.Body)
	}

	uri := r.rawRequest.Uri
	if len(r.rawRequest.RawQueryParams) > 0 || len(r.rawRequest.EncodedQueryParams) > 0 {
		reqURL, err := url.Parse(r.rawRequest.Uri)
		if err != nil {
			return nil, fmt.Errorf("raw request has invalid URI: %s: %w", r.rawRequest.Uri, err)
		}
		vals := reqURL.Query()
		for _, param := range r.rawRequest.RawQueryParams {
			vals[param.Name] = append(vals[param.Name], param.Value...)
		}
		for _, param := range r.rawRequest.EncodedQueryParams {
			var buf bytes.Buffer
			if err := internal.WriteRawMessageContents(param.Value, &buf); err != nil {
				return nil, fmt.Errorf("raw request has invalid encoded query param %s: %w", param.Name, err)
			}
			var paramVal string
			if param.Base64Encode {
				paramVal = base64.URLEncoding.EncodeToString(buf.Bytes())
			} else {
				paramVal = buf.String()
			}
			vals[param.Name] = append(vals[param.Name], paramVal)
		}
		reqURL.RawQuery = vals.Encode()
		uri = reqURL.String()
	}

	pipeReader, pipeWriter := io.Pipe()
	req, err := http.NewRequestWithContext(orig.Context(), r.rawRequest.Verb, uri, pipeReader)
	if err != nil {
		if orig.Body != nil {
			_ = orig.Body.Close()
		}
		return nil, err
	}

	// Write the request body to the pipe.
	go func() {
		defer func() {
			_ = pipeWriter.Close()
		}()
		switch contents := r.rawRequest.Body.(type) {
		case *conformancev1.RawHTTPRequest_Unary:
			_ = internal.WriteRawMessageContents(contents.Unary, pipeWriter)
		case *conformancev1.RawHTTPRequest_Stream:
			_ = internal.WriteRawStreamContents(contents.Stream, pipeWriter)
		}
	}()

	// Consume the request body and close it.
	go func() {
		defer func() {
			_ = orig.Body.Close()
		}()
		_, _ = io.Copy(io.Discard, orig.Body)
	}()

	return r.transport.RoundTrip(req)
}
