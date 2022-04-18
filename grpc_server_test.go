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
	"time"

	connectpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/connect/grpc/testing/testingconnect"
	testgrpc "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/grpc/testing"
	interopconnect "github.com/bufbuild/connect-crosstest/internal/interop/connect"
	interopgrpc "github.com/bufbuild/connect-crosstest/internal/interop/grpc"
	crosstesting "github.com/bufbuild/connect-crosstest/internal/testing"
	"github.com/bufbuild/connect-go"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
)

type T testing.T

func TestGRPCServer(t *testing.T) {
	lis, err := net.Listen("tcp", "localhost:0")
	assert.NoError(t, err)
	server := grpc.NewServer()
	testgrpc.RegisterTestServiceServer(server, interopgrpc.NewTestServer())
	go func() {
		server.Serve(lis)
	}()
	defer server.GracefulStop()
	t.Run("grpc_client", func(testingT *testing.T) {
		t := crosstesting.NewCrossTestT(testingT)
		gconn, err := grpc.Dial(
			lis.Addr().String(),
			grpc.WithInsecure(),
		)
		assert.NoError(t, err)
		defer gconn.Close()
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
		interopgrpc.DoDuplicatedCustomMetadata(t, client)
		interopgrpc.DoStatusCodeAndMessage(t, client)
		interopgrpc.DoSpecialStatusMessage(t, client)
		interopgrpc.DoUnimplementedMethod(t, gconn)
		interopgrpc.DoUnimplementedService(t, client)
		interopgrpc.DoFailWithNonASCIIError(t, client)
		interopgrpc.DoSoakTest(
			t,
			client,
			lis.Addr().String(),
			nil,
			false, /* resetChannel */
			soakIterations,
			0,
			perIterationMaxAcceptableLatency,
			time.Now().Add(10*1000*time.Millisecond), /* soakIterations * perIterationMaxAcceptableLatency */
		)
	})
	t.Run("connect_client", func(testingT *testing.T) {
		t := crosstesting.NewCrossTestT(testingT)
		client, err := connectpb.NewTestServiceClient(newClientH2C(), "http://"+lis.Addr().String(), connect.WithGRPC())
		assert.NoError(t, err)
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
		interopconnect.DoDuplicatedCustomMetadata(t, client)
		interopconnect.DoStatusCodeAndMessage(t, client)
		interopconnect.DoSpecialStatusMessage(t, client)
		interopconnect.DoUnimplementedService(t, client)
		interopconnect.DoFailWithNonASCIIError(t, client)
		interopconnect.DoSoakTest(
			t,
			client,
			"http://"+lis.Addr().String(),
			false, /* resetChannel */
			soakIterations,
			0,
			perIterationMaxAcceptableLatency,
			time.Now().Add(10*1000*time.Millisecond), /* soakIterations * perIterationMaxAcceptableLatency */
		)
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
