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

package grpcutil

import (
	"context"
	"strings"

	v1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"google.golang.org/grpc/metadata"
)

// ConvertMetadataToProtoHeader converts gRPC metadata into proto Headers.
func ConvertMetadataToProtoHeader(
	src metadata.MD,
) []*v1.Header {
	headerInfo := make([]*v1.Header, 0, len(src))
	for key, value := range src {
		// This fixes the problem, but is this the right solution? Doing this
		// removes anything grpc specific from the headers returned to the test
		// But maybe that's fine bc connect has already tried to parse the grpc
		// stuff by now, so as long as we know the error that grpc returned by
		// now we can discard anything we used to glean that information.
		//
		// We sort of have to do this here bc the problem is trying to write
		// this base64 encoded data to the output (utf-8? buffer length?)
		// So it needs to be omitted before we send out the response, but thats
		// ok as long as we know the status and details by now.
		if key == "grpc-status-details-bin" {
			continue
		}
		hdr := &v1.Header{
			Name:  key,
			Value: value,
		}
		headerInfo = append(headerInfo, hdr)
	}
	return headerInfo
}

// ConvertProtoHeaderToMetadata converts a slice of proto Headers into gRPC metadata.
func ConvertProtoHeaderToMetadata(
	src []*v1.Header,
) metadata.MD {
	md := make(metadata.MD, len(src))
	for _, hdr := range src {
		key := strings.ToLower(hdr.Name)
		md[key] = hdr.Value
	}
	return md
}

// Appends the given headers to the outgoing context. Used for sending metadata
// from the client side.
func AppendToOutgoingContext(ctx context.Context, src []*v1.Header) context.Context {
	for _, hdr := range src {
		for _, val := range hdr.Value {
			ctx = metadata.AppendToOutgoingContext(ctx, hdr.Name, val)
		}
	}
	return ctx
}
