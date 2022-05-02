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
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	testrpc "github.com/bufbuild/connect-crosstest/internal/gen/proto/connect/grpc/testing/testingconnect"
	serverpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/server/v1"
	interopconnect "github.com/bufbuild/connect-crosstest/internal/interop/connect"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/encoding/protojson"
)

type flags struct {
	h1Port string
	h2Port string
}

func main() {
	flagset := flags{}
	rootCmd := &cobra.Command{
		Use:   "serverconnect",
		Short: "Starts a connect test server",
		Run: func(cmd *cobra.Command, args []string) {
			run(flagset)
		},
	}
	rootCmd.Flags().StringVar(&flagset.h1Port, "h1port", "", "port for HTTP/1.1 traffic")
	rootCmd.Flags().StringVar(&flagset.h2Port, "h2port", "", "port for HTTP/2 traffic")
	_ = rootCmd.MarkFlagRequired("h1port")
	_ = rootCmd.MarkFlagRequired("h2port")
	_ = rootCmd.Execute()
}

func run(flagset flags) {
	mux := http.NewServeMux()
	mux.Handle(testrpc.NewTestServiceHandler(
		interopconnect.NewTestConnectServer(),
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
		ExposedHeaders: []string{"Grpc-Status", "Grpc-Message", "Grpc-Status-Details-Bin"},
	}).Handler(mux)
	h1Server := http.Server{
		Addr:    ":" + flagset.h1Port,
		Handler: corsHandler,
	}
	h2Server := http.Server{
		Addr:    ":" + flagset.h2Port,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}
	bytes, err := protojson.Marshal(
		&serverpb.ServerMetadata{
			Host: "localhost",
			Protocols: []*serverpb.ProtocolSupport{
				{
					Protocol: serverpb.Protocol_PROTOCOL_GRPC_WEB,
					HttpVersions: []*serverpb.HTTPVersion{
						{
							Major: int32(1),
							Minor: int32(1),
						},
					},
					Port: flagset.h1Port,
				},
				{
					Protocol: serverpb.Protocol_PROTOCOL_GRPC_WEB,
					HttpVersions: []*serverpb.HTTPVersion{
						{
							Major: int32(2),
						},
					},
					Port: flagset.h2Port,
				},
				{
					Protocol: serverpb.Protocol_PROTOCOL_GRPC,
					HttpVersions: []*serverpb.HTTPVersion{
						{
							Major: int32(2),
						},
					},
					Port: flagset.h2Port,
				},
			},
		},
	)
	if err != nil {
		log.Fatalf("failed to marshal server metadata: %v", err)
	}
	_, _ = fmt.Fprintln(os.Stdout, string(bytes))
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := h1Server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Fatalln(err)
		}
	}()
	go func() {
		if err := h2Server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Fatalln(err)
		}
	}()
	<-done
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h1Server.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
	if err := h2Server.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}
