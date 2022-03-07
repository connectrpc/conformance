// Copyright 2020-2022 Buf Technologies, Inc.
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

package crossgrpc

import (
	"context"
	"fmt"
	"log"

	testpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/grpc/testing"
	interopconnect "github.com/bufbuild/connect-crosstest/internal/interop/connect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func DoFailWithNonASCIIError(tc testpb.TestServiceClient, args ...grpc.CallOption) {
	reply, err := tc.FailUnaryCall(context.Background(), &testpb.SimpleRequest{
		ResponseType: testpb.PayloadType_COMPRESSABLE,
	},
	)
	if err != nil {
		if reply != nil {
			log.Fatalf("reply should be empty: %v", reply)
		}
		s, ok := status.FromError(err)
		if !ok {
			log.Fatalf("unable to get grpc status from error")
		}
		if s.Code() != codes.ResourceExhausted {
			log.Fatalf("incorrect status code received: %v", s.Code())
		}
		if s.Message() != interopconnect.NonASCIIErrMsg {
			log.Fatalf("incorrect error message received: %s", s.Message())
		}
		fmt.Println("successful fail call with non-ASCII error")
	}
}
