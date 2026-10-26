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

	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/connect/v2"
)

// HeaderAdder is implemented by both [http.Header] and [*connect.Header].
type HeaderAdder interface {
	Add(key, value string)
}

// AddHeaders adds all header values in src to dest.
func AddHeaders(
	src []*conformancev1.Header,
	dest HeaderAdder,
) {
	for _, header := range src {
		for _, val := range header.Value {
			dest.Add(header.Name, val)
		}
	}
}

// AddTrailers adds all header values in src to dest, but
// it prefixes each header name with http.TrailerPrefix.
func AddTrailers(
	src []*conformancev1.Header,
	dest http.Header,
) {
	for _, header := range src {
		for _, val := range header.Value {
			dest.Add(http.TrailerPrefix+header.Name, val)
		}
	}
}

// ConvertToProtoHeader converts a map to a slice of proto Headers.
// Note that this can accept types of url.Values and http.Header.
func ConvertToProtoHeader(src map[string][]string) []*conformancev1.Header {
	headerInfo := make([]*conformancev1.Header, 0, len(src))
	for key, value := range src {
		hdr := &conformancev1.Header{
			Name:  key,
			Value: value,
		}
		headerInfo = append(headerInfo, hdr)
	}
	return headerInfo
}

// ConvertConnectHeaderToProto converts connect metadata to a slice of proto
// Headers. It returns nil if src is nil.
func ConvertConnectHeaderToProto(src *connect.Header) []*conformancev1.Header {
	if src == nil {
		return nil
	}
	headerInfo := make([]*conformancev1.Header, 0, src.Len())
	for key, value := range src.All() {
		headerInfo = append(headerInfo, &conformancev1.Header{
			Name:  key,
			Value: value,
		})
	}
	return headerInfo
}
