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
	"net"
	"net/http"

	"connectrpc.com/conformance/internal/app"
	"connectrpc.com/conformance/internal/gen/proto/connect/connectrpc/conformance/v1alpha1/conformancev1alpha1connect"
	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const (
	// The default host to use for the server
	defaultHost = "127.0.0.1"
	// The default port to use for the server. We choose 0 so that
	// an ephemeral port is selected by the OS
	defaultPort = "0"
)

// Run runs the server according to server config read from the 'in' reader.
func Run(ctx context.Context, args []string, in io.ReadCloser, out io.WriteCloser) error {
	json := flag.Bool("json", false, "whether to use the JSON format for marshaling / unmarshaling messages")

	flag.Parse()

	// Read the server config from  the in reader
	data, err := io.ReadAll(in)
	if err != nil {
		return err
	}

	marshaler := app.NewMarshaler(*json)

	// Unmarshal into a ServerCompatRequest
	req := &v1alpha1.ServerCompatRequest{}
	if err := marshaler.Unmarshal(data, req); err != nil {
		return err
	}

	// Create an HTTP server based on the request
	server, err := createServer(req)
	if err != nil {
		return err
	}

	// Create a listener for the server so that we are able to obtain
	// the IP and port for publishing on the out writer
	ln, err := net.Listen("tcp", net.JoinHostPort(defaultHost, defaultPort))
	if err != nil {
		return err
	}
	resp := &v1alpha1.ServerCompatResponse{
		Result: &v1alpha1.ServerCompatResponse_Listening{
			Listening: &v1alpha1.ServerListeningResult{
				Host: fmt.Sprint(ln.Addr().(*net.TCPAddr).IP),
				Port: fmt.Sprint(ln.Addr().(*net.TCPAddr).Port),
			},
		},
	}
	bytes, err := marshaler.Marshal(resp)
	if err != nil {
		return err
	}
	if _, err := out.Write(bytes); err != nil {
		return err
	}

	// Finally, start the server
	if err := server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

// Creates an HTTP server using the provided ServerCompatRequest
func createServer(req *v1alpha1.ServerCompatRequest) (*http.Server, error) {
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
	var server *http.Server
	switch req.HttpVersion {
	case v1alpha1.HTTPVersion_HTTP_VERSION_1:
		server = newH1Server(corsHandler)
	case v1alpha1.HTTPVersion_HTTP_VERSION_2:
		server = newH2Server(mux)
	case v1alpha1.HTTPVersion_HTTP_VERSION_3:
		return nil, errors.New("HTTP/3 is not yet supported")
	case v1alpha1.HTTPVersion_HTTP_VERSION_UNSPECIFIED:
		return nil, errors.New("an HTTP version must be specified.")
	}

	return server, nil
}

// Create a new HTTP/1.1 server
func newH1Server(handler http.Handler) *http.Server {
	h1Server := &http.Server{
		Addr:    net.JoinHostPort(defaultHost, defaultPort),
		Handler: handler,
	}
	return h1Server
}

// Create a new HTTP/2 server
func newH2Server(handler http.Handler) *http.Server {
	h2Server := &http.Server{
		Addr: net.JoinHostPort(defaultHost, defaultPort),
	}
	h2Server.Handler = h2c.NewHandler(handler, &http2.Server{})
	return h2Server
}
