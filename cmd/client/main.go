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
	"crypto/x509"
	"io/ioutil"
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
	"github.com/lucas-clemente/quic-go/http3"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type flags struct {
	host           string
	port           string
	implementation string
	certFile       string
	keyFile        string
}

func main() {
	flagset := flags{}
	rootCmd := &cobra.Command{
		Use:   "client",
		Short: "Starts a grpc or connect client, based on implementation",
		Run: func(cmd *cobra.Command, args []string) {
			run(flagset)
		},
	}
	rootCmd.Flags().StringVar(&flagset.host, "host", "127.0.0.1", "the host name of the test server")
	rootCmd.Flags().StringVar(&flagset.port, "port", "", "the port of the test server")
	rootCmd.Flags().StringVarP(
		&flagset.implementation,
		"implementation",
		"i",
		"connect",
		`the client implementation tested, accepted values are "connect-h2", "connect-h3" or "grpc-go"`,
	)
	rootCmd.Flags().StringVar(&flagset.certFile, "cert", "", "path to the TLS cert file")
	rootCmd.Flags().StringVar(&flagset.keyFile, "key", "", "path to the TLS key file")
	_ = rootCmd.MarkFlagRequired("port")
	_ = rootCmd.MarkFlagRequired("cert")
	_ = rootCmd.MarkFlagRequired("key")
	_ = rootCmd.Execute()
}

func run(flagset flags) {
	switch flagset.implementation {
	case "connect-h2", "connect-h3":
		serverURL, err := url.ParseRequestURI("https://" + net.JoinHostPort(flagset.host, flagset.port))
		if err != nil {
			log.Fatalf("invalid url: %s", "https://"+net.JoinHostPort(flagset.host, flagset.port))
		}
		client := connectpb.NewTestServiceClient(
			newClient(flagset),
			serverURL.String(),
			connect.WithGRPC(),
		)
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
			net.JoinHostPort(flagset.host, flagset.port),
			grpc.WithTransportCredentials(credentials.NewTLS(newTLSConfig(flagset))),
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
	default:
		log.Fatalf(`must set --implementation or -i to "connect-h2", "connect-h3" or "grpc-go"`)
	}
}

func newClient(flagset flags) *http.Client {
	tlsConfig := newTLSConfig(flagset)
	var transport http.RoundTripper
	switch flagset.implementation {
	case "connect-h2":
		transport = &http2.Transport{
			TLSClientConfig: tlsConfig,
		}
	case "connect-h3":
		transport = &http3.RoundTripper{
			TLSClientConfig: tlsConfig,
		}
	default:
		log.Fatalf("unknown implementation flag to create client")
	}
	// This is wildly insecure - don't do this in production!
	return &http.Client{
		Transport: transport,
	}
}

func newTLSConfig(flagset flags) *tls.Config {
	cert, err := tls.LoadX509KeyPair(flagset.certFile, flagset.keyFile)
	if err != nil {
		log.Fatalf("Error creating x509 keypair from client cert file %s and client key file %s", flagset.certFile, flagset.keyFile)
	}
	caCert, err := ioutil.ReadFile("cert/CrosstestCA.crt")
	if err != nil {
		log.Fatalf("Error opening cert file")
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
}
