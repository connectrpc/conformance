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
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	testrpc "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/grpc/testing"
	serverpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/server/v1"
	"github.com/bufbuild/connect-crosstest/internal/interop/interopgrpc"
	"github.com/bufbuild/connect-go/grpcadapter"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip" // this register the gzip compressor to the grpc server
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	portFlagName        = "port"
	adapterPortFlagName = "adapterport"
	certFlagName        = "cert"
	keyFlagName         = "key"
)

type flags struct {
	port        string
	adapterPort string
	certFile    string
	keyFile     string
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
	cmd.Flags().StringVar(&flagset.adapterPort, adapterPortFlagName, "", "port for gRPC adapter traffic")
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
	testrpc.RegisterTestServiceServer(server, interopgrpc.NewTestServer())

	adapterServer := newAdapterServer(flagset, server)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		err := adapterServer.ListenAndServeTLS(flagset.certFile, flagset.keyFile)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalln(err)
		}
	}()
	go func() {
		_ = server.Serve(lis)
	}()
	<-done
	server.GracefulStop()
}

func newTLSConfig(certFile, keyFile string) *tls.Config {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("Error creating x509 keypair from client cert file %s and client key file %s", certFile, keyFile)
	}
	caCert, err := os.ReadFile("cert/CrosstestCA.crt")
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

func newAdapterServer(flags *flags, server *grpc.Server) *http.Server {
	handler := grpcadapter.NewHandler(server)
	adapterServer := &http.Server{
		Addr:              ":" + flags.adapterPort,
		ReadHeaderTimeout: 3 * time.Second,
	}
	adapterServer.Handler = h2c.NewHandler(handler, &http2.Server{})
	return adapterServer
}
