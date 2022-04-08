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
	"net/http"

	testrpc "github.com/bufbuild/connect-crosstest/internal/gen/proto/connect/grpc/testing/testingconnect"
	serverpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/server/v1"
	interopconnect "github.com/bufbuild/connect-crosstest/internal/interop/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	h1Port := flag.String("h1port", "", "the http1 port that the connect server will listen on")
	h2Port := flag.String("h2port", "", "the http2 port that the connect server will listen on")
	flag.Parse()
	if *h1Port == "" || *h2Port == "" {
		log.Fatal("--h1port and --h2port must both be set")
	}
	mux := http.NewServeMux()
	mux.Handle(testrpc.NewTestServiceHandler(
		interopconnect.NewTestConnectServer(),
	))
	bytes, err := protojson.Marshal(
		&serverpb.ServerMetadata{
			Host: "localhost",
			Protocols: []*serverpb.ProtocolSupport{
				{
					Protocol: serverpb.Protocol_PROTOCOL_GRPC_WEB,
					HttpVersions: []*serverpb.HTTPVersion{
						{
							Major: int32(1),
							Minor: int32(1),
						},
					},
					Port: *h1Port,
				},
				{
					Protocol: serverpb.Protocol_PROTOCOL_GRPC,
					HttpVersions: []*serverpb.HTTPVersion{
						{
							Major: int32(2),
						},
					},
					Port: *h2Port,
				},
			},
		},
	)
	if err != nil {
		log.Fatalf("failed to marshal server metadata: %v", err)
	}
	fmt.Println(string(bytes))
	go http.ListenAndServe(
		":"+*h1Port,
		mux,
	)
	http.ListenAndServe(
		":"+*h2Port,
		h2c.NewHandler(mux, &http2.Server{}),
	)
}
