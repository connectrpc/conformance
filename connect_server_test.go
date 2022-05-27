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
	"time"

	"github.com/bufbuild/connect-crosstest/internal/crosstesting"
	"github.com/bufbuild/connect-crosstest/internal/gen/proto/connect/grpc/testing/testingconnect"
	testgrpc "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/grpc/testing"
	"github.com/bufbuild/connect-crosstest/internal/interopconnect"
	"github.com/bufbuild/connect-crosstest/internal/interopgrpc"
	"github.com/bufbuild/connect-go"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	// gRPC interop test defaults, per https://github.com/grpc/grpc/blob/master/doc/interop-test-descriptions.md#rpc_soak
	soakIterations                   = 10
	perIterationMaxAcceptableLatency = 1000 * time.Millisecond
)

func TestConnectServer(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.Handle(testingconnect.NewTestServiceHandler(
		interopconnect.NewTestServiceHandler(),
	))
	t.Run("grpc_client", func(t *testing.T) {
		t.Parallel()
		server := newUnstartedServer(mux)
		defer server.Close()
		pool := x509.NewCertPool()
		pool.AddCert(server.Certificate())
		gconn, err := grpc.Dial(
			server.Listener.Addr().String(),
			grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(pool, "")),
		)
		assert.NoError(t, err)
		defer gconn.Close()
		client := testgrpc.NewTestServiceClient(gconn)
		interopgrpc.DoEmptyUnaryCall(crosstesting.NewTB(t), client)
		interopgrpc.DoLargeUnaryCall(crosstesting.NewTB(t), client)
		interopgrpc.DoClientStreaming(crosstesting.NewTB(t), client)
		interopgrpc.DoServerStreaming(crosstesting.NewTB(t), client)
		interopgrpc.DoPingPong(crosstesting.NewTB(t), client)
		interopgrpc.DoEmptyStream(crosstesting.NewTB(t), client)
		interopgrpc.DoTimeoutOnSleepingServer(crosstesting.NewTB(t), client)
		interopgrpc.DoCancelAfterBegin(crosstesting.NewTB(t), client)
		interopgrpc.DoCancelAfterFirstResponse(crosstesting.NewTB(t), client)
		interopgrpc.DoCustomMetadata(crosstesting.NewTB(t), client)
		interopgrpc.DoDuplicatedCustomMetadata(crosstesting.NewTB(t), client)
		interopgrpc.DoStatusCodeAndMessage(crosstesting.NewTB(t), client)
		interopgrpc.DoSpecialStatusMessage(crosstesting.NewTB(t), client)
		interopgrpc.DoUnimplementedMethod(crosstesting.NewTB(t), gconn)
		interopgrpc.DoUnimplementedService(crosstesting.NewTB(t), client)
		interopgrpc.DoFailWithNonASCIIError(crosstesting.NewTB(t), client)
	})
	t.Run("grpc_client soak test", func(t *testing.T) {
		t.Parallel()
		if testing.Short() {
			t.Skip("skipping test in short mode")
		}
		server := newUnstartedServer(mux)
		defer server.Close()
		pool := x509.NewCertPool()
		pool.AddCert(server.Certificate())
		gconn, err := grpc.Dial(
			server.Listener.Addr().String(),
			grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(pool, "")),
		)
		assert.NoError(t, err)
		defer gconn.Close()
		client := testgrpc.NewTestServiceClient(gconn)
		interopgrpc.DoSoakTest(
			crosstesting.NewTB(t),
			client,
			server.Listener.Addr().String(),
			nil,
			false, /* resetChannel */
			soakIterations,
			0,
			perIterationMaxAcceptableLatency,
			time.Now().Add(10*1000*time.Millisecond), /* soakIterations * perIterationMaxAcceptableLatency */
		)
	})
	t.Run("connect_client", func(t *testing.T) {
		t.Parallel()
		server := newUnstartedServer(mux)
		defer server.Close()
		client := testingconnect.NewTestServiceClient(server.Client(), server.URL, connect.WithGRPC())
		interopconnect.DoEmptyUnaryCall(crosstesting.NewTB(t), client)
		interopconnect.DoLargeUnaryCall(crosstesting.NewTB(t), client)
		interopconnect.DoClientStreaming(crosstesting.NewTB(t), client)
		interopconnect.DoServerStreaming(crosstesting.NewTB(t), client)
		interopconnect.DoPingPong(crosstesting.NewTB(t), client)
		interopconnect.DoEmptyStream(crosstesting.NewTB(t), client)
		interopconnect.DoTimeoutOnSleepingServer(crosstesting.NewTB(t), client)
		interopconnect.DoCancelAfterBegin(crosstesting.NewTB(t), client)
		interopconnect.DoCancelAfterFirstResponse(crosstesting.NewTB(t), client)
		interopconnect.DoCustomMetadata(crosstesting.NewTB(t), client)
		interopconnect.DoDuplicatedCustomMetadata(crosstesting.NewTB(t), client)
		interopconnect.DoStatusCodeAndMessage(crosstesting.NewTB(t), client)
		interopconnect.DoSpecialStatusMessage(crosstesting.NewTB(t), client)
		interopconnect.DoUnimplementedService(crosstesting.NewTB(t), client)
		interopconnect.DoFailWithNonASCIIError(crosstesting.NewTB(t), client)
	})
	t.Run("connect_client soak test", func(t *testing.T) {
		t.Parallel()
		if testing.Short() {
			t.Skip("skipping test in short mode")
		}
		server := newUnstartedServer(mux)
		defer server.Close()
		client := testingconnect.NewTestServiceClient(server.Client(), server.URL, connect.WithGRPC())
		interopconnect.DoSoakTest(
			crosstesting.NewTB(t),
			client,
			server.Listener.Addr().String(),
			false, /* resetChannel */
			soakIterations,
			0,
			perIterationMaxAcceptableLatency,
			time.Now().Add(10*1000*time.Millisecond), /* soakIterations * perIterationMaxAcceptableLatency */
		)
	})
}

func newUnstartedServer(handler http.Handler) *httptest.Server {
	server := httptest.NewUnstartedServer(handler)
	server.EnableHTTP2 = true
	server.StartTLS()
	return server
}
