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
	conformancev1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
)

// NewConformanceServiceServer creates a new Conformance Service server.
func NewConformanceServiceServer() conformancev1alpha1.ConformanceServiceServer {
	return &conformanceServiceServer{}
}

type conformanceServiceServer struct {
	conformancev1alpha1.UnimplementedConformanceServiceServer
}
