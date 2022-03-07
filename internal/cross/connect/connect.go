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

package crossconnect

import (
	"context"
	"errors"
	"fmt"
	"log"

	"google.golang.org/grpc"

	"github.com/bufbuild/connect"
	connectpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/connect/grpc/testing/testingconnect"
	testpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/grpc/testing"
	interopconnect "github.com/bufbuild/connect-crosstest/internal/interop/connect"
)

func DoFailWithNonASCIIError(tc connectpb.TestServiceClient, args ...grpc.CallOption) {
	reply, err := tc.FailUnaryCall(
		context.Background(),
		connect.NewRequest(
			&testpb.SimpleRequest{
				ResponseType: testpb.PayloadType_COMPRESSABLE,
			},
		),
	)
	if err != nil {
		if reply != nil {
			log.Fatalf("reply should be empty: %v", reply)
		}
		var connectErr *connect.Error
		ok := errors.As(err, &connectErr)
		if !ok {
			log.Fatalf("failed to convert error to connect error: %v", err)
		}
		if connectErr.Code() != connect.CodeResourceExhausted {
			log.Fatalf("incorrect status code received: %v", connectErr.Code())
		}
		if connectErr.Error() != connect.CodeResourceExhausted.String()+": "+interopconnect.NonASCIIErrMsg {
			log.Fatalf("incorrect error message received: %s", connectErr.Error())
		}
		fmt.Println("successful fail call with non-ASCII error")
	}
}
