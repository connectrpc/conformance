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
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"testing"

	connectpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/connect/grpc/testing/testingconnect"
	testgrpc "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/grpc/testing"
	interopconnect "github.com/bufbuild/connect-crosstest/internal/interop/connect"
	interopgrpc "github.com/bufbuild/connect-crosstest/internal/interop/grpc"
	"github.com/bufbuild/connect-go"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestConnectServer(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle(connectpb.NewTestServiceHandler(
		interopconnect.NewTestConnectServer(),
	))
	server := httptest.NewUnstartedServer(mux)
	server.EnableHTTP2 = true
	// TODO(doria): Can I do this without using TLS?
	server.StartTLS()
	defer server.Close()
	t.Run("grpc_client", func(t *testing.T) {
		pool := x509.NewCertPool()
		pool.AddCert(server.Certificate())
		gconn, err := grpc.Dial(
			server.Listener.Addr().String(),
			grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(pool, "")),
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
		interopgrpc.DoStatusCodeAndMessage(t, client)
		interopgrpc.DoSpecialStatusMessage(t, client)
		interopgrpc.DoUnimplementedMethod(t, gconn)
		interopgrpc.DoUnimplementedService(t, client)
		interopgrpc.DoFailWithNonASCIIError(t, client)
	})
	t.Run("connect_client", func(t *testing.T) {
		client, err := connectpb.NewTestServiceClient(server.Client(), server.URL, connect.WithGRPC())
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
		interopconnect.DoStatusCodeAndMessage(t, client)
		interopconnect.DoSpecialStatusMessage(t, client)
		interopconnect.DoUnimplementedService(t, client)
		interopconnect.DoFailWithNonASCIIError(t, client)
	})
}
