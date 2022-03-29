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

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	crossgrpc "github.com/bufbuild/connect-crosstest/internal/cross/grpc"
	testgrpc "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/grpc/testing"
	serverpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/server/v1"
	interopgrpc "github.com/bufbuild/connect-crosstest/internal/interop/grpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

// TODO(doria): takes in which http versions; picks the tests that are of interest based on version but also relevance to client itself
func main() {
	reader := bufio.NewReader(os.Stdin)
	serverMetadataRaw, err := reader.ReadBytes('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		log.Fatalf("failed to read server metadata: %v", err)
	}
	var serverMetadata serverpb.ServerMetadata
	if err := protojson.Unmarshal(serverMetadataRaw, &serverMetadata); err != nil {
		log.Fatalf("failed to unmarshal server metadata: %v", err)
	}
	fmt.Println("received server metadata", serverMetadata.String())
	gconn, err := grpc.Dial(
		serverMetadata.Address,
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("failed grpc dial: %v", err)
	}
	defer gconn.Close()
	client := testgrpc.NewTestServiceClient(gconn)
	interopgrpc.DoEmptyUnaryCall(client)
	interopgrpc.DoLargeUnaryCall(client)
	interopgrpc.DoClientStreaming(client)
	interopgrpc.DoServerStreaming(client)
	interopgrpc.DoPingPong(client)
	interopgrpc.DoEmptyStream(client)
	interopgrpc.DoTimeoutOnSleepingServer(client)
	interopgrpc.DoCancelAfterBegin(client)
	interopgrpc.DoCancelAfterFirstResponse(client)
	interopgrpc.DoCustomMetadata(client)
	interopgrpc.DoStatusCodeAndMessage(client)
	interopgrpc.DoSpecialStatusMessage(client)
	interopgrpc.DoUnimplementedMethod(gconn)
	interopgrpc.DoUnimplementedService(client)
	// TODO(doria): add cross tests
	crossgrpc.DoFailWithNonASCIIError(client)
}
