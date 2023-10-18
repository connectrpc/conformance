// Copyright 2022-2023 The Connect Authors
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

package server

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/conformance/internal/gen/proto/connect/connectrpc/conformance/v1alpha1/conformancev1alpha1connect"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func Run(ctx context.Context, args []string, in io.ReadCloser, out, err io.WriteCloser) error {
	// TODO - The ports should not be given to the run call. Instead the server should read a
	// ServerCompatRequest and start itself. It should then return the port numbers back on stdout
	// as a ServerCompatResponse
	h1Port := flag.String("h1Port", "8080", "port for HTTP/1.1 traffic")
	h2Port := flag.String("h2Port", "8081", "port for HTTP/2 traffic")

	flag.Parse()

	mux := http.NewServeMux()
	mux.Handle(conformancev1alpha1connect.NewConformanceServiceHandler(
		&conformanceServer{},
	))
	// The server needs a lenient cors setup so that it can handle testing
	// browser clients.
	corsHandler := cors.New(cors.Options{
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		// Note that rs/cors does not return `Access-Control-Allow-Headers: *`
		// in response to preflight requests with the following configuration.
		// It simply mirrors all headers listed in the `Access-Control-Request-Headers`
		// preflight request header.
		AllowedHeaders: []string{"*"},
		// Expose all headers
		ExposedHeaders: []string{"*"},
	}).Handler(mux)

	// Create servers
	h1Server := newH1Server(*h1Port, corsHandler)
	h2Server := newH2Server(*h2Port, mux)
	done := make(chan os.Signal, 1)
	errs := make(chan error, 2)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		err := h1Server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()
	go func() {
		err := h2Server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()

	fmt.Printf("HTTP/1.1 server listening on port %s\nHTTP/2 server listening on port %s", *h1Port, *h2Port)

	select {
	case err := <-errs:
		return err
	case <-done:
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h1Server.Shutdown(ctx); err != nil {
		return err
	}
	if err := h2Server.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

func newH1Server(h1Port string, handler http.Handler) *http.Server {
	h1Server := &http.Server{
		Addr:    ":" + h1Port,
		Handler: handler,
	}
	return h1Server
}

func newH2Server(h2Port string, handler http.Handler) *http.Server {
	h2Server := &http.Server{
		Addr: ":" + h2Port,
	}
	h2Server.Handler = h2c.NewHandler(handler, &http2.Server{})
	return h2Server
}
