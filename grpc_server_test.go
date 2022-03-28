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

package crosstest

import (
	"crypto/tls"
	"net"
	"net/http"
	"testing"

	crossgrpc "github.com/bufbuild/connect-crosstest/internal/cross/grpc"
	connectpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/connect/grpc/testing/testingconnect"
	testgrpc "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/grpc/testing"
	interopconnect "github.com/bufbuild/connect-crosstest/internal/interop/connect"
	interopgrpc "github.com/bufbuild/connect-crosstest/internal/interop/grpc"
	"github.com/bufbuild/connect-go"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
)

func TestGRPCServer(t *testing.T) {
	lis, err := net.Listen("tcp", "localhost:0")
	assert.NoError(t, err)
	server := grpc.NewServer()
	testgrpc.RegisterTestServiceServer(server, interopgrpc.NewTestServer())
	go func() {
		server.Serve(lis)
	}()
	defer server.GracefulStop()
	t.Run("grpc_client", func(t *testing.T) {
		gconn, err := grpc.Dial(
			lis.Addr().String(),
			grpc.WithInsecure(),
		)
		assert.NoError(t, err)
		defer gconn.Close()
		client := testgrpc.NewTestServiceClient(gconn)
		assert.NotPanics(t, func() { interopgrpc.DoEmptyUnaryCall(client) })
		assert.NotPanics(t, func() { interopgrpc.DoLargeUnaryCall(client) })
		assert.NotPanics(t, func() { interopgrpc.DoClientStreaming(client) })
		assert.NotPanics(t, func() { interopgrpc.DoServerStreaming(client) })
		assert.NotPanics(t, func() { interopgrpc.DoPingPong(client) })
		assert.NotPanics(t, func() { interopgrpc.DoEmptyStream(client) })
		assert.NotPanics(t, func() { interopgrpc.DoTimeoutOnSleepingServer(client) })
		assert.NotPanics(t, func() { interopgrpc.DoCancelAfterBegin(client) })
		assert.NotPanics(t, func() { interopgrpc.DoCancelAfterFirstResponse(client) })
		assert.NotPanics(t, func() { interopgrpc.DoCustomMetadata(client) })
		assert.NotPanics(t, func() { interopgrpc.DoStatusCodeAndMessage(client) })
		assert.NotPanics(t, func() { interopgrpc.DoSpecialStatusMessage(client) })
		assert.NotPanics(t, func() { interopgrpc.DoUnimplementedMethod(gconn) })
		assert.NotPanics(t, func() { interopgrpc.DoUnimplementedService(client) })
		assert.NotPanics(t, func() { crossgrpc.DoFailWithNonASCIIError(client) })
	})
	t.Run("connect_client", func(t *testing.T) {
		client, err := connectpb.NewTestServiceClient(newClientH2C(), "http://"+lis.Addr().String(), connect.WithGRPC())
		assert.NoError(t, err)
		assert.NotPanics(t, func() { interopconnect.DoEmptyUnaryCall(client) })
		assert.NotPanics(t, func() { interopconnect.DoLargeUnaryCall(client) })
		assert.NotPanics(t, func() { interopconnect.DoClientStreaming(client) })
		assert.NotPanics(t, func() { interopconnect.DoServerStreaming(client) })
		assert.NotPanics(t, func() { interopconnect.DoPingPong(client) })
		assert.NotPanics(t, func() { interopconnect.DoEmptyStream(client) })
		assert.NotPanics(t, func() { interopconnect.DoTimeoutOnSleepingServer(client) })
		assert.NotPanics(t, func() { interopconnect.DoCancelAfterBegin(client) })
		assert.NotPanics(t, func() { interopconnect.DoCancelAfterFirstResponse(client) })
		assert.NotPanics(t, func() { interopconnect.DoCustomMetadata(client) })
		// TODO(doria): fix connect client test cases
		// assert.NotPanics(t, func() { interopconnect.DoStatusCodeAndMessage(client) })
		// assert.NotPanics(t, func() { interopconnect.DoSpecialStatusMessage(client) })
		// assert.NotPanics(t, func() { interopconnect.DoUnimplementedService(client) })
		// assert.NotPanics(t, func() { crossconnect.DoFailWithNonASCIIError(client) })
	})
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
