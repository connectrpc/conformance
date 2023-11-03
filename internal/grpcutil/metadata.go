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

	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func ConvertMetadataToProtoHeader(
	src metadata.MD,
) []*v1alpha1.Header {
	headerInfo := make([]*v1alpha1.Header, 0, len(src))
	for key, value := range src {
		hdr := &v1alpha1.Header{
			Name:  key,
			Value: value,
		}
		headerInfo = append(headerInfo, hdr)
	}
	return headerInfo
}

func ConvertProtoHeaderToMetadata(
	src []*v1alpha1.Header,
) metadata.MD {
	md := make(metadata.MD, len(src))
	for _, hdr := range src {
		key := strings.ToLower(hdr.Name)
		md[key] = hdr.Value
	}
	return md
}

func AddHeaderMetadata(
	ctx context.Context,
	hdrs []*v1alpha1.Header,
) error {
	md := ConvertProtoHeaderToMetadata(hdrs)
	return grpc.SetHeader(ctx, md)
}

func AddTrailerMetadata(
	ctx context.Context,
	hdrs []*v1alpha1.Header,
) error {
	md := ConvertProtoHeaderToMetadata(hdrs)
	return grpc.SetTrailer(ctx, md)
}
