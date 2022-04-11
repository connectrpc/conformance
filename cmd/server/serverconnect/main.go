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
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	testrpc "github.com/bufbuild/connect-crosstest/internal/gen/proto/connect/grpc/testing/testingconnect"
	serverpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/server/v1"
	interopconnect "github.com/bufbuild/connect-crosstest/internal/interop/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	h1Port := flag.String("h1port", "", "port for HTTP/1.1 traffic")
	h2Port := flag.String("h2port", "", "port for HTTP/2 traffic")
	flag.Parse()
	if *h1Port == "" || *h2Port == "" {
		log.Fatal("--h1port and --h2port must be set")
	}
	mux := http.NewServeMux()
	mux.Handle(testrpc.NewTestServiceHandler(
		interopconnect.NewTestConnectServer(),
	))
	h1Server := http.Server{
		Addr:    ":" + *h1Port,
		Handler: mux,
	}
	h2Server := http.Server{
		Addr:    ":" + *h2Port,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}
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
					Protocol: serverpb.Protocol_PROTOCOL_GRPC_WEB,
					HttpVersions: []*serverpb.HTTPVersion{
						{
							Major: int32(2),
						},
					},
					Port: *h2Port,
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
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := h1Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	}()
	go func() {
		if err := h2Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	}()
	<-done
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h1Server.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
	if err := h2Server.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}
