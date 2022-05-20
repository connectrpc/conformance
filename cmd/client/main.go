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
	"google.golang.org/grpc/encoding/gzip"
)

const (
	connectH2 = "connect-h2"
	connectH3 = "connect-h3"
	grpcGo    = "grpc-go"
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
	case connectH2:
		serverURL, err := url.ParseRequestURI("https://" + net.JoinHostPort(flagset.host, flagset.port))
		if err != nil {
			log.Fatalf("invalid url: %s", "https://"+net.JoinHostPort(flagset.host, flagset.port))
		}
		client := connectpb.NewTestServiceClient(
			&http.Client{
				Transport: &http2.Transport{
					TLSClientConfig: newTLSConfig(flagset.certFile, flagset.keyFile),
				},
			},
			serverURL.String(),
			connect.WithGRPC(),
		)
		unresolvableClient := connectpb.NewTestServiceClient(
			&http.Client{
				Transport: &http2.Transport{
					TLSClientConfig: newTLSConfig(flagset.certFile, flagset.keyFile),
				},
			},
			"https://unresolvable-host.some.domain",
			connect.WithGRPC(),
		)
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
		interopconnect.DoUnresolvableHost(console.NewTB(), unresolvableClient)
		compressedClient := connectpb.NewTestServiceClient(
			&http.Client{
				Transport: &http2.Transport{
					TLSClientConfig: newTLSConfig(flagset.certFile, flagset.keyFile),
				},
			},
			serverURL.String(),
			connect.WithGRPC(),
			connect.WithSendGzip(),
		)
		interopconnect.DoEmptyUnaryCall(console.NewTB(), compressedClient)
		interopconnect.DoLargeUnaryCall(console.NewTB(), compressedClient)
		interopconnect.DoClientStreaming(console.NewTB(), compressedClient)
		interopconnect.DoServerStreaming(console.NewTB(), compressedClient)
		interopconnect.DoPingPong(console.NewTB(), compressedClient)
		interopconnect.DoEmptyStream(console.NewTB(), compressedClient)
		interopconnect.DoTimeoutOnSleepingServer(console.NewTB(), compressedClient)
		interopconnect.DoCancelAfterBegin(console.NewTB(), compressedClient)
		interopconnect.DoCancelAfterFirstResponse(console.NewTB(), compressedClient)
		interopconnect.DoCustomMetadata(console.NewTB(), compressedClient)
		interopconnect.DoStatusCodeAndMessage(console.NewTB(), compressedClient)
		interopconnect.DoSpecialStatusMessage(console.NewTB(), compressedClient)
		interopconnect.DoUnimplementedService(console.NewTB(), compressedClient)
		interopconnect.DoFailWithNonASCIIError(console.NewTB(), compressedClient)
	case connectH3:
		serverURL, err := url.ParseRequestURI("https://" + net.JoinHostPort(flagset.host, flagset.port))
		if err != nil {
			log.Fatalf("invalid url: %s", "https://"+net.JoinHostPort(flagset.host, flagset.port))
		}
		client := connectpb.NewTestServiceClient(
			&http.Client{
				Transport: &http3.RoundTripper{
					TLSClientConfig: newTLSConfig(flagset.certFile, flagset.keyFile),
				},
			},
			serverURL.String(),
			connect.WithGRPC(),
		)
		// For tests that depend  trailers, we only run them for HTTP2, since the HTTP3 client
		// does not yet have trailers support https://github.com/lucas-clemente/quic-go/issues/2266
		interopconnect.DoEmptyUnaryCall(console.NewTB(), client)
		interopconnect.DoLargeUnaryCall(console.NewTB(), client)
		interopconnect.DoClientStreaming(console.NewTB(), client)
		interopconnect.DoServerStreaming(console.NewTB(), client)
		interopconnect.DoPingPong(console.NewTB(), client)
		compressedClient := connectpb.NewTestServiceClient(
			&http.Client{
				Transport: &http3.RoundTripper{
					TLSClientConfig: newTLSConfig(flagset.certFile, flagset.keyFile),
				},
			},
			serverURL.String(),
			connect.WithGRPC(),
			connect.WithSendGzip(),
		)
		interopconnect.DoEmptyUnaryCall(console.NewTB(), compressedClient)
		interopconnect.DoLargeUnaryCall(console.NewTB(), compressedClient)
		interopconnect.DoClientStreaming(console.NewTB(), compressedClient)
		interopconnect.DoServerStreaming(console.NewTB(), compressedClient)
		interopconnect.DoPingPong(console.NewTB(), compressedClient)
	case grpcGo:
		gconn, err := grpc.Dial(
			net.JoinHostPort(flagset.host, flagset.port),
			grpc.WithTransportCredentials(credentials.NewTLS(newTLSConfig(flagset.certFile, flagset.keyFile))),
		)
		if err != nil {
			log.Fatalf("failed grpc dial: %v", err)
		}
		defer gconn.Close()
		client := testgrpc.NewTestServiceClient(gconn)
		unresolvableGconn, err := grpc.Dial(
			"unresolvable-host.some.domain",
			grpc.WithTransportCredentials(credentials.NewTLS(newTLSConfig(flagset.certFile, flagset.keyFile))),
		)
		if err != nil {
			log.Fatalf("failed grpc dial: %v", err)
		}
		defer unresolvableGconn.Close()
		unresolvableClient := testgrpc.NewTestServiceClient(unresolvableGconn)
		interopgrpc.DoEmptyUnaryCall(console.NewTB(), client)
		interopgrpc.DoLargeUnaryCall(console.NewTB(), client)
		interopgrpc.DoClientStreaming(console.NewTB(), client)
		interopgrpc.DoServerStreaming(console.NewTB(), client)
		interopgrpc.DoPingPong(console.NewTB(), client)
		interopgrpc.DoEmptyStream(console.NewTB(), client)
		interopgrpc.DoTimeoutOnSleepingServer(console.NewTB(), client)
		interopgrpc.DoCancelAfterBegin(console.NewTB(), client)
		interopgrpc.DoCancelAfterFirstResponse(console.NewTB(), client)
		interopgrpc.DoCustomMetadata(console.NewTB(), client)
		interopgrpc.DoStatusCodeAndMessage(console.NewTB(), client)
		interopgrpc.DoSpecialStatusMessage(console.NewTB(), client)
		interopgrpc.DoUnimplementedMethod(console.NewTB(), gconn)
		interopgrpc.DoUnimplementedService(console.NewTB(), client)
		interopgrpc.DoFailWithNonASCIIError(console.NewTB(), client)
		interopgrpc.DoUnresolvableHost(console.NewTB(), unresolvableClient)
		interopgrpc.DoEmptyUnaryCall(console.NewTB(), client, grpc.UseCompressor(gzip.Name))
		interopgrpc.DoLargeUnaryCall(console.NewTB(), client, grpc.UseCompressor(gzip.Name))
		interopgrpc.DoClientStreaming(console.NewTB(), client, grpc.UseCompressor(gzip.Name))
		interopgrpc.DoServerStreaming(console.NewTB(), client, grpc.UseCompressor(gzip.Name))
		interopgrpc.DoPingPong(console.NewTB(), client, grpc.UseCompressor(gzip.Name))
		interopgrpc.DoEmptyStream(console.NewTB(), client, grpc.UseCompressor(gzip.Name))
		interopgrpc.DoTimeoutOnSleepingServer(console.NewTB(), client, grpc.UseCompressor(gzip.Name))
		interopgrpc.DoCancelAfterBegin(console.NewTB(), client, grpc.UseCompressor(gzip.Name))
		interopgrpc.DoCancelAfterFirstResponse(console.NewTB(), client, grpc.UseCompressor(gzip.Name))
		interopgrpc.DoCustomMetadata(console.NewTB(), client, grpc.UseCompressor(gzip.Name))
		interopgrpc.DoStatusCodeAndMessage(console.NewTB(), client, grpc.UseCompressor(gzip.Name))
		interopgrpc.DoSpecialStatusMessage(console.NewTB(), client, grpc.UseCompressor(gzip.Name))
		interopgrpc.DoUnimplementedMethod(console.NewTB(), gconn, grpc.UseCompressor(gzip.Name))
		interopgrpc.DoUnimplementedService(console.NewTB(), client, grpc.UseCompressor(gzip.Name))
		interopgrpc.DoFailWithNonASCIIError(console.NewTB(), client, grpc.UseCompressor(gzip.Name))
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
