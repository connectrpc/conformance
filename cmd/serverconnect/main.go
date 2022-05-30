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
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bufbuild/connect-crosstest/internal/gen/proto/connect/grpc/testing/testingconnect"
	serverpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/server/v1"
	"github.com/bufbuild/connect-crosstest/internal/interopconnect"
	"github.com/lucas-clemente/quic-go/http3"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

type flags struct {
	h1Port   string
	h2Port   string
	h3Port   string
	certFile string
	keyFile  string
}

func main() {
	flags := &flags{}
	rootCmd := &cobra.Command{
		Use:   "serverconnect",
		Short: "Starts a connect test server",
		Run: func(cmd *cobra.Command, args []string) {
			run(flags)
		},
	}
	rootCmd.Flags().StringVar(&flags.h1Port, "h1port", "", "port for HTTP/1.1 traffic")
	rootCmd.Flags().StringVar(&flags.h2Port, "h2port", "", "port for HTTP/2 traffic")
	rootCmd.Flags().StringVar(&flags.h3Port, "h3port", "", "port for HTTP/3 traffic")
	rootCmd.Flags().StringVar(&flags.certFile, "cert", "", "path to the TLS cert file")
	rootCmd.Flags().StringVar(&flags.keyFile, "key", "", "path to the TLS key file")
	_ = rootCmd.MarkFlagRequired("h1port")
	_ = rootCmd.MarkFlagRequired("h2port")
	_ = rootCmd.MarkFlagRequired("cert")
	_ = rootCmd.MarkFlagRequired("key")
	_ = rootCmd.Execute()
}

func run(flags *flags) {
	mux := http.NewServeMux()
	mux.Handle(testingconnect.NewTestServiceHandler(
		interopconnect.NewTestServiceHandler(),
	))
	corsHandler := cors.New(cors.Options{
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		// Mirror the `Origin` header value in the `Access-Control-Allow-Origin`
		// preflight response header.
		// This is equivalent to `Access-Control-Allow-Origin: *`, but allows
		// for requests with credentials.
		// Note that this effectively disables CORS and is not safe for use in
		// production environments.
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		// Note that rs/cors does not return `Access-Control-Allow-Headers: *`
		// in response to preflight requests with the following configuration.
		// It simply mirrors all headers listed in the `Access-Control-Request-Headers`
		// preflight request header.
		AllowedHeaders: []string{"*"},
		// We explicitly set the exposed header names instead of using the wildcard *,
		// because in requests with credentials, it is treated as the literal header
		// name "*" without special semantics.
		ExposedHeaders: []string{"Grpc-Status", "Grpc-Message", "Grpc-Status-Details-Bin", "X-Grpc-Test-Echo-Initial"},
	}).Handler(mux)
	tlsConfig := newTLSConfig(flags.certFile, flags.keyFile)
	h1Server := http.Server{
		Addr:      ":" + flags.h1Port,
		Handler:   corsHandler,
		TLSConfig: tlsConfig,
	}
	h2Server := http.Server{
		Addr:      ":" + flags.h2Port,
		Handler:   mux,
		TLSConfig: tlsConfig,
	}
	var h3Server http3.Server
	if flags.h3Port != "" {
		h3Server = http3.Server{
			Server: &http.Server{
				Addr:      ":" + flags.h3Port,
				Handler:   mux,
				TLSConfig: tlsConfig,
			},
		}
	}
	protocols := []*serverpb.ProtocolSupport{
		{
			Protocol: serverpb.Protocol_PROTOCOL_GRPC_WEB,
			HttpVersions: []*serverpb.HTTPVersion{
				{
					Major: int32(1),
					Minor: int32(1),
				},
			},
			Port: flags.h1Port,
		},
		{
			Protocol: serverpb.Protocol_PROTOCOL_GRPC_WEB,
			HttpVersions: []*serverpb.HTTPVersion{
				{
					Major: int32(2),
				},
			},
			Port: flags.h2Port,
		},
		{
			Protocol: serverpb.Protocol_PROTOCOL_GRPC,
			HttpVersions: []*serverpb.HTTPVersion{
				{
					Major: int32(2),
				},
			},
			Port: flags.h2Port,
		},
	}
	if flags.h3Port != "" {
		protocols = append(protocols, &serverpb.ProtocolSupport{
			Protocol: serverpb.Protocol_PROTOCOL_GRPC,
			HttpVersions: []*serverpb.HTTPVersion{
				{
					Major: int32(3),
				},
			},
			Port: flags.h3Port,
		})
	}
	bytes, err := protojson.Marshal(
		&serverpb.ServerMetadata{
			Host:      "localhost",
			Protocols: protocols,
		},
	)
	if err != nil {
		log.Fatalf("failed to marshal server metadata: %v", err)
	}
	_, _ = fmt.Fprintln(os.Stdout, string(bytes))
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := h1Server.ListenAndServeTLS(flags.certFile, flags.keyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalln(err)
		}
	}()
	go func() {
		if err := h2Server.ListenAndServeTLS(flags.certFile, flags.keyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalln(err)
		}
	}()
	if flags.h3Port != "" {
		go func() {
			if err := h3Server.ListenAndServeTLS(flags.certFile, flags.keyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalln(err)
			}
		}()
	}
	<-done
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h1Server.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
	if err := h2Server.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
	if flags.h3Port != "" {
		if err := h3Server.Shutdown(ctx); err != nil {
			log.Fatalln(err)
		}
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
