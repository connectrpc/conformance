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

package grpcserver

import (
	"context"

	conformancev1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewConformanceServiceServer creates a new Conformance Service server.
func NewConformanceServiceServer() conformancev1alpha1.ConformanceServiceServer {
	return &conformanceServiceServer{}
}

type conformanceServiceServer struct {
	conformancev1alpha1.UnimplementedConformanceServiceServer
}

func (c *conformanceServiceServer) Unary(
	ctx context.Context,
	req *conformancev1alpha1.UnaryRequest,
) (*conformancev1alpha1.UnaryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Unary not implemented")
}

func (c *conformanceServiceServer) ClientStream(
	stream conformancev1alpha1.ConformanceService_ClientStreamServer,
) error {
	return status.Errorf(codes.Unimplemented, "method ClientStream not implemented")
}

func (c *conformanceServiceServer) ServerStream(
	req *conformancev1alpha1.ServerStreamRequest,
	stream conformancev1alpha1.ConformanceService_ServerStreamServer,
) error {
	return status.Errorf(codes.Unimplemented, "method ServerStream not implemented")
}

func (c *conformanceServiceServer) BidiStream(
	stream conformancev1alpha1.ConformanceService_BidiStreamServer,
) error {
	return status.Errorf(codes.Unimplemented, "method BidiStream not implemented")
}
