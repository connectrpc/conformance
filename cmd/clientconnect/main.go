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
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/bufbuild/connect"
	crossconnect "github.com/bufbuild/connect-crosstest/internal/cross/connect"
	connectpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/connect/grpc/testing/testingconnect"
	serverpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/server/v1"
	interopconnect "github.com/bufbuild/connect-crosstest/internal/interop/connect"
	"golang.org/x/net/http2"
	"google.golang.org/protobuf/encoding/protojson"
)

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
	client, err := connectpb.NewTestServiceClient(newClientH2C(), "http://"+serverMetadata.Address, connect.WithGRPC())
	if err != nil {
		log.Fatalf("failed to create connect client: %v", err)
	}
	interopconnect.DoEmptyUnaryCall(client)
	interopconnect.DoLargeUnaryCall(client)
	interopconnect.DoClientStreaming(client)
	interopconnect.DoServerStreaming(client)
	interopconnect.DoPingPong(client)
	interopconnect.DoEmptyStream(client)
	interopconnect.DoTimeoutOnSleepingServer(client)
	interopconnect.DoCancelAfterBegin(client)
	interopconnect.DoCancelAfterFirstResponse(client)
	// interopconnect.DoCustomMetadata(client)
	interopconnect.DoStatusCodeAndMessage(client)
	interopconnect.DoSpecialStatusMessage(client)
	interopconnect.DoUnimplementedService(client)
	// TODO(doria): add cross tests
	crossconnect.DoFailWithNonASCIIError(client)
}
