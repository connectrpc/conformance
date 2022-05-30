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
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"

	"github.com/bufbuild/connect-crosstest/internal/console"
	"github.com/bufbuild/connect-crosstest/internal/gen/proto/connect/grpc/testing/testingconnect"
	testgrpc "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/grpc/testing"
	"github.com/bufbuild/connect-crosstest/internal/interopconnect"
	"github.com/bufbuild/connect-crosstest/internal/interopgrpc"
	"github.com/bufbuild/connect-go"
	"github.com/lucas-clemente/quic-go/http3"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

const (
	connectGRPCH2 = "connect-grpc-h2"
	connectGRPCH3 = "connect-grpc-h3"
	connectH2     = "connect-h2"
	connectH3     = "connect-h3"
	grpcGo        = "grpc-go"
)

type flags struct {
	host           string
	port           string
	implementation string
	certFile       string
	keyFile        string
}

func main() {
	flags := &flags{}
	rootCmd := &cobra.Command{
		Use:   "client",
		Short: "Starts a grpc or connect client, based on implementation",
		Run: func(cmd *cobra.Command, args []string) {
			run(flags)
		},
	}
	rootCmd.Flags().StringVar(&flags.host, "host", "127.0.0.1", "the host name of the test server")
	rootCmd.Flags().StringVar(&flags.port, "port", "", "the port of the test server")
	rootCmd.Flags().StringVarP(
		&flags.implementation,
		"implementation",
		"i",
		"connect",
		fmt.Sprintf(
			"the client implementation tested, accepted values are %q, %q, %q, %q, or %q",
			connectGRPCH2,
			connectGRPCH3,
			connectH2,
			connectH3,
			grpcGo,
		),
	)
	rootCmd.Flags().StringVar(&flags.certFile, "cert", "", "path to the TLS cert file")
	rootCmd.Flags().StringVar(&flags.keyFile, "key", "", "path to the TLS key file")
	_ = rootCmd.MarkFlagRequired("port")
	_ = rootCmd.MarkFlagRequired("cert")
	_ = rootCmd.MarkFlagRequired("key")
	_ = rootCmd.Execute()
}

func run(flags *flags) {
	serverURL, err := url.ParseRequestURI("https://" + net.JoinHostPort(flags.host, flags.port))
	if err != nil {
		log.Fatalf("invalid url: %s", "https://"+net.JoinHostPort(flags.host, flags.port))
	}
	switch flags.implementation {
	case connectGRPCH2, connectH2:
		var clientOptions []connect.ClientOption
		if flags.implementation == connectGRPCH2 {
			clientOptions = append(clientOptions, connect.WithGRPC())
		}
		transport := &http2.Transport{
			TLSClientConfig: newTLSConfig(flags.certFile, flags.keyFile),
		}
		uncompressedClient := testingconnect.NewTestServiceClient(
			&http.Client{Transport: transport},
			serverURL.String(),
			clientOptions...,
		)
		clientOptions = append(clientOptions, connect.WithSendGzip())
		compressedClient := testingconnect.NewTestServiceClient(
			&http.Client{Transport: transport},
			serverURL.String(),
			clientOptions...,
		)
		for _, client := range []testingconnect.TestServiceClient{uncompressedClient, compressedClient} {
			interopconnect.DoEmptyUnaryCall(console.NewTB(), client)
			interopconnect.DoLargeUnaryCall(console.NewTB(), client)
			interopconnect.DoClientStreaming(console.NewTB(), client)
			interopconnect.DoServerStreaming(console.NewTB(), client)
			interopconnect.DoPingPong(console.NewTB(), client)
			interopconnect.DoEmptyStream(console.NewTB(), client)
			interopconnect.DoTimeoutOnSleepingServer(console.NewTB(), client)
			interopconnect.DoCancelAfterBegin(console.NewTB(), client)
			interopconnect.DoCancelAfterFirstResponse(console.NewTB(), client)
			interopconnect.DoCustomMetadata(console.NewTB(), client)
			interopconnect.DoStatusCodeAndMessage(console.NewTB(), client)
			interopconnect.DoSpecialStatusMessage(console.NewTB(), client)
			interopconnect.DoUnimplementedService(console.NewTB(), client)
			interopconnect.DoFailWithNonASCIIError(console.NewTB(), client)
		}
		interopconnect.DoUnresolvableHost(
			console.NewTB(), testingconnect.NewTestServiceClient(
				&http.Client{Transport: transport},
				"https://unresolvable-host.some.domain",
				connect.WithGRPC(),
			),
		)
	// For tests that depend  trailers, we only run them for HTTP2, since the HTTP3 client
	// does not yet have trailers support https://github.com/lucas-clemente/quic-go/issues/2266
	case connectGRPCH3, connectH3:
		var clientOptions []connect.ClientOption
		if flags.implementation == connectGRPCH3 {
			clientOptions = append(clientOptions, connect.WithGRPC())
		}
		transport := &http3.RoundTripper{
			TLSClientConfig: newTLSConfig(flags.certFile, flags.keyFile),
		}
		uncompressedClient := testingconnect.NewTestServiceClient(
			&http.Client{Transport: transport},
			serverURL.String(),
			clientOptions...,
		)
		clientOptions = append(clientOptions, connect.WithSendGzip())
		compressedClient := testingconnect.NewTestServiceClient(
			&http.Client{Transport: transport},
			serverURL.String(),
			clientOptions...,
		)
		for _, client := range []testingconnect.TestServiceClient{uncompressedClient, compressedClient} {
			// For tests that depend  trailers, we only run them for HTTP2, since the HTTP3 client
			// does not yet have trailers support https://github.com/lucas-clemente/quic-go/issues/2266
			interopconnect.DoEmptyUnaryCall(console.NewTB(), client)
			interopconnect.DoLargeUnaryCall(console.NewTB(), client)
			interopconnect.DoClientStreaming(console.NewTB(), client)
			interopconnect.DoServerStreaming(console.NewTB(), client)
			interopconnect.DoPingPong(console.NewTB(), client)
		}
	case grpcGo:
		clientConn, err := grpc.Dial(
			net.JoinHostPort(flags.host, flags.port),
			grpc.WithTransportCredentials(credentials.NewTLS(newTLSConfig(flags.certFile, flags.keyFile))),
		)
		if err != nil {
			log.Fatalf("failed grpc dial: %v", err)
		}
		defer clientConn.Close()
		client := testgrpc.NewTestServiceClient(clientConn)
		for _, args := range [][]grpc.CallOption{
			nil,
			{grpc.UseCompressor(gzip.Name)},
		} {
			interopgrpc.DoEmptyUnaryCall(console.NewTB(), client, args...)
			interopgrpc.DoLargeUnaryCall(console.NewTB(), client, args...)
			interopgrpc.DoClientStreaming(console.NewTB(), client, args...)
			interopgrpc.DoServerStreaming(console.NewTB(), client, args...)
			interopgrpc.DoPingPong(console.NewTB(), client, args...)
			interopgrpc.DoEmptyStream(console.NewTB(), client, args...)
			interopgrpc.DoTimeoutOnSleepingServer(console.NewTB(), client, args...)
			interopgrpc.DoCancelAfterBegin(console.NewTB(), client, args...)
			interopgrpc.DoCancelAfterFirstResponse(console.NewTB(), client, args...)
			interopgrpc.DoCustomMetadata(console.NewTB(), client, args...)
			interopgrpc.DoStatusCodeAndMessage(console.NewTB(), client, args...)
			interopgrpc.DoSpecialStatusMessage(console.NewTB(), client, args...)
			interopgrpc.DoUnimplementedMethod(console.NewTB(), clientConn, args...)
			interopgrpc.DoUnimplementedService(console.NewTB(), client, args...)
			interopgrpc.DoFailWithNonASCIIError(console.NewTB(), client, args...)
		}
		unresolvableClientConn, err := grpc.Dial(
			"unresolvable-host.some.domain",
			grpc.WithTransportCredentials(credentials.NewTLS(newTLSConfig(flags.certFile, flags.keyFile))),
		)
		if err != nil {
			log.Fatalf("failed grpc dial: %v", err)
		}
		defer unresolvableClientConn.Close()
		interopgrpc.DoUnresolvableHost(
			console.NewTB(),
			testgrpc.NewTestServiceClient(unresolvableClientConn),
		)
	default:
		log.Fatalf(`must set --implementation or -i to "connect-h2", "connect-h3" or "grpc-go"`)
	}
}

func newTLSConfig(certFile, keyFile string) *tls.Config {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("Error creating x509 keypair from client cert file %s and client key file %s", certFile, keyFile)
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
