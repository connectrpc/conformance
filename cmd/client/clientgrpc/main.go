// Copyright 2022 Buf Technologies, Inc.
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
	"flag"
	"log"
	"net"

	"github.com/bufbuild/connect-crosstest/internal/console"
	testgrpc "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/grpc/testing"
	interopgrpc "github.com/bufbuild/connect-crosstest/internal/interop/grpc"
	"google.golang.org/grpc"
)

func main() {
	host := flag.String("host", "127.0.0.1", "the host name of the test server, defaults to 127.0.0.1")
	port := flag.String("port", "", "the port of the test server")
	flag.Parse()
	if *port == "" {
		log.Fatalf("--port must both be set")
	}
	gconn, err := grpc.Dial(
		net.JoinHostPort(*host, *port),
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("failed grpc dial: %v", err)
	}
	defer gconn.Close()
	t := console.NewTB()
	client := testgrpc.NewTestServiceClient(gconn)
	interopgrpc.DoEmptyUnaryCall(t, client)
	interopgrpc.DoLargeUnaryCall(t, client)
	interopgrpc.DoClientStreaming(t, client)
	interopgrpc.DoServerStreaming(t, client)
	interopgrpc.DoPingPong(t, client)
	interopgrpc.DoEmptyStream(t, client)
	interopgrpc.DoTimeoutOnSleepingServer(t, client)
	interopgrpc.DoCancelAfterBegin(t, client)
	interopgrpc.DoCancelAfterFirstResponse(t, client)
	interopgrpc.DoCustomMetadata(t, client)
	interopgrpc.DoStatusCodeAndMessage(t, client)
	interopgrpc.DoSpecialStatusMessage(t, client)
	interopgrpc.DoUnimplementedMethod(t, gconn)
	interopgrpc.DoUnimplementedService(t, client)
	interopgrpc.DoFailWithNonASCIIError(t, client)
}
