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
	"fmt"
	"log"
	"net"

	testrpc "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/grpc/testing"
	serverpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/server/v1"
	interopgrpc "github.com/bufbuild/connect-crosstest/internal/interop/grpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	port := flag.String("port", "", "the port the server will listen on")
	flag.Parse()
	if *port == "" {
		log.Fatal("--port must be set")
	}
	lis, err := net.Listen("tcp", net.JoinHostPort("localhost", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()
	host, _, err := net.SplitHostPort(lis.Addr().String())
	if err != nil {
		log.Fatalf("failed to split host port from listener addr: %s", lis.Addr().String())
	}
	bytes, err := protojson.Marshal(
		&serverpb.ServerMetadata{
			Host: host,
			Protocols: []*serverpb.ProtocolSupport{
				{
					Protocol: serverpb.Protocol_PROTOCOL_GRPC,
					HttpVersions: []*serverpb.HTTPVersion{
						{
							Major: int32(2),
						},
					},
					Port: *port,
				},
			},
		},
	)
	if err != nil {
		log.Fatalf("failed to marshal server metadata: %v", err)
	}
	fmt.Println(string(bytes))
	testrpc.RegisterTestServiceServer(server, interopgrpc.NewTestServer())

	server.Serve(lis)
	defer server.GracefulStop()
}
