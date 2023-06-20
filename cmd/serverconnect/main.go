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
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bufbuild/connect-crosstest/internal/gen/proto/connect/grpc/testing/testingconnect"
	serverpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/server/v1"
	"github.com/bufbuild/connect-crosstest/internal/interop/interopconnect"
	"github.com/quic-go/quic-go/http3"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	h1PortFlagName   = "h1port"
	h2PortFlagName   = "h2port"
	h3PortFlagName   = "h3port"
	certFlagName     = "cert"
	keyFlagName      = "key"
	insecureFlagName = "insecure"
)

type flags struct {
	h1Port   string
	h2Port   string
	h3Port   string
	certFile string
	keyFile  string
	insecure bool
}

func main() {
	flagset := &flags{}
	rootCmd := &cobra.Command{
		Use:   "serverconnect",
		Short: "Starts a connect test server",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			insecure, _ := cmd.Flags().GetBool(insecureFlagName)
			certFile, _ := cmd.Flags().GetString(certFlagName)
			keyFile, _ := cmd.Flags().GetString(keyFlagName)
			h3Port, _ := cmd.Flags().GetString(h3PortFlagName)
			if !insecure && (certFile == "" || keyFile == "") {
				return errors.New("either a 'cert' and 'key' combination or 'insecure' must be specified")
			}
			if h3Port != "" && (certFile == "" || keyFile == "") {
				return errors.New("a 'cert' and 'key' combination is required when an HTTP/3 port is specified")
			}
			return nil
		},
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
	cmd.Flags().StringVar(&flagset.h1Port, h1PortFlagName, "", "port for HTTP/1.1 traffic")
	cmd.Flags().StringVar(&flagset.h2Port, h2PortFlagName, "", "port for HTTP/2 traffic")
	cmd.Flags().StringVar(&flagset.h3Port, h3PortFlagName, "", "port for HTTP/3 traffic")
	cmd.Flags().StringVar(&flagset.certFile, certFlagName, "", "path to the TLS cert file")
	cmd.Flags().StringVar(&flagset.keyFile, keyFlagName, "", "path to the TLS key file")
	cmd.Flags().BoolVar(&flagset.insecure, insecureFlagName, false, "whether to serve cleartext or TLS. HTTP/3 requires TLS.")
	for _, requiredFlag := range []string{h1PortFlagName, h2PortFlagName} {
		if err := cmd.MarkFlagRequired(requiredFlag); err != nil {
			return err
		}
	}
	return nil
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
		ExposedHeaders: []string{
			"Grpc-Status", "Grpc-Message", "Grpc-Status-Details-Bin", "X-Grpc-Test-Echo-Initial",
			"Trailer-X-Grpc-Test-Echo-Trailing-Bin", "Request-Protocol", "Get-Request"},
	}).Handler(mux)

	// Create servers
	h1Server := newH1Server(flags, corsHandler)
	h2Server := newH2Server(flags, mux)
	var h3Server http3.Server
	if flags.h3Port != "" {
		h3Server = newH3Server(flags, mux)
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
					Major: int32(1),
					Minor: int32(1),
				},
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
					Major: int32(1),
					Minor: int32(1),
				},
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
		var err error
		if flags.insecure {
			err = h1Server.ListenAndServe()
		} else {
			err = h1Server.ListenAndServeTLS(flags.certFile, flags.keyFile)
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalln(err)
		}
	}()
	go func() {
		var err error
		if flags.insecure {
			err = h2Server.ListenAndServe()
		} else {
			err = h2Server.ListenAndServeTLS(flags.certFile, flags.keyFile)
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
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
		if err := h3Server.Close(); err != nil {
			log.Fatalln(err)
		}
	}
}

func newH1Server(flags *flags, handler http.Handler) *http.Server {
	h1Server := &http.Server{
		Addr:    ":" + flags.h1Port,
		Handler: handler,
	}
	if !flags.insecure {
		h1Server.TLSConfig = newTLSConfig(flags.certFile, flags.keyFile)
	}
	return h1Server
}

func newH2Server(flags *flags, handler http.Handler) *http.Server {
	h2Server := &http.Server{
		Addr: ":" + flags.h2Port,
	}
	if !flags.insecure {
		h2Server.TLSConfig = newTLSConfig(flags.certFile, flags.keyFile)
		h2Server.Handler = handler
	} else {
		h2Server.Handler = h2c.NewHandler(handler, &http2.Server{})
	}
	return h2Server
}

func newH3Server(flags *flags, handler http.Handler) http3.Server {
	return http3.Server{
		Addr:      ":" + flags.h3Port,
		Handler:   handler,
		TLSConfig: newTLSConfig(flags.certFile, flags.keyFile),
	}
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
