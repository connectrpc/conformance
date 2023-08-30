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
	"log"
	"net"
	"os"

	conformance "github.com/connectrpc/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	serverpb "github.com/connectrpc/conformance/internal/gen/proto/go/server/v1"
	"github.com/connectrpc/conformance/internal/interop/interopgrpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip" // this register the gzip compressor to the grpc server
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	portFlagName = "port"
	certFlagName = "cert"
	keyFlagName  = "key"
)

type flags struct {
	port     string
	certFile string
	keyFile  string
}

func main() {
	flagset := &flags{}
	rootCmd := &cobra.Command{
		Use:   "servergrpc",
		Short: "Starts a grpc test server",
		Run: func(cmd *cobra.Command, args []string) {
			run(flagset)
		},
	}
	if err := bind(rootCmd, flagset); err != nil {
		os.Exit(1)
	}
	_ = rootCmd.Execute()
}

func bind(cmd *cobra.Command, flagset *flags) error {
	cmd.Flags().StringVar(&flagset.port, portFlagName, "", "the port the server will listen on")
	cmd.Flags().StringVar(&flagset.certFile, certFlagName, "", "path to the TLS cert file")
	cmd.Flags().StringVar(&flagset.keyFile, keyFlagName, "", "path to the TLS key file")
	for _, requiredFlag := range []string{portFlagName, certFlagName, keyFlagName} {
		if err := cmd.MarkFlagRequired(requiredFlag); err != nil {
			return err
		}
	}
	return nil
}

func run(flagset *flags) {
	lis, err := net.Listen("tcp", ":"+flagset.port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(newTLSConfig(flagset.certFile, flagset.keyFile))),
	)
	bytes, err := protojson.Marshal(
		&serverpb.ServerMetadata{
			Host: "localhost",
			Protocols: []*serverpb.ProtocolSupport{
				{
					Protocol: serverpb.Protocol_PROTOCOL_GRPC,
					HttpVersions: []*serverpb.HTTPVersion{
						{
							Major: int32(2),
						},
					},
					Port: flagset.port,
				},
			},
		},
	)
	if err != nil {
		log.Fatalf("failed to marshal server metadata: %v", err)
	}
	_, _ = fmt.Fprintln(os.Stdout, string(bytes))
	conformance.RegisterTestServiceServer(server, interopgrpc.NewTestServer())
	_ = server.Serve(lis)
	defer server.GracefulStop()
}

func newTLSConfig(certFile, keyFile string) *tls.Config {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("Error creating x509 keypair from client cert file %s and client key file %s", certFile, keyFile)
	}
	caCert, err := os.ReadFile("cert/ConformanceCA.crt")
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
