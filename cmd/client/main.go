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
	"crypto/tls"
	"flag"
	"log"
	"net"
	"net/http"
	"net/url"

	"github.com/bufbuild/connect-crosstest/internal/console"
	connectpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/connect/grpc/testing/testingconnect"
	testgrpc "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/grpc/testing"
	interopconnect "github.com/bufbuild/connect-crosstest/internal/interop/connect"
	interopgrpc "github.com/bufbuild/connect-crosstest/internal/interop/grpc"
	"github.com/bufbuild/connect-go"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
)

func main() {
	host := flag.String("host", "127.0.0.1", "the host name of the test server, defaults to 127.0.0.1")
	port := flag.String("port", "", "the port of the test server")
	implementation := flag.String(
		"implementation",
		"connect",
		`the client implementation tested, accepted values are "connect" or "grpc-go", defaults to "connect"`,
	)
	flag.Parse()
	if *port == "" {
		log.Fatalf("--port must both be set")
	}
	if *implementation != "connect" && *implementation != "grpc-go" {
		log.Fatalf(`--implementation must be set to "connect" or "grpc-go": %v`)
	}
	switch *implementation {
	case "connect":
		serverURL, err := url.ParseRequestURI("http://" + net.JoinHostPort(*host, *port))
		if err != nil {
			log.Fatalf("invalid url: %s", "http://"+net.JoinHostPort(*host, *port))
		}
		client, err := connectpb.NewTestServiceClient(
			newClientH2C(),
			serverURL.String(),
			connect.WithGRPC(),
		)
		if err != nil {
			log.Fatalf("failed to create connect client: %v", err)
		}
		t := console.NewTB()
		interopconnect.DoEmptyUnaryCall(t, client)
		interopconnect.DoLargeUnaryCall(t, client)
		interopconnect.DoClientStreaming(t, client)
		interopconnect.DoServerStreaming(t, client)
		interopconnect.DoPingPong(t, client)
		interopconnect.DoEmptyStream(t, client)
		interopconnect.DoTimeoutOnSleepingServer(t, client)
		interopconnect.DoCancelAfterBegin(t, client)
		interopconnect.DoCancelAfterFirstResponse(t, client)
		interopconnect.DoCustomMetadata(t, client)
		interopconnect.DoStatusCodeAndMessage(t, client)
		interopconnect.DoSpecialStatusMessage(t, client)
		interopconnect.DoUnimplementedService(t, client)
		interopconnect.DoFailWithNonASCIIError(t, client)
	case "grpc-go":
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
}

func newClientH2C() *http.Client {
	// This is wildly insecure - don't do this in production!
	return &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(netw, addr)
			},
		},
	}
}
