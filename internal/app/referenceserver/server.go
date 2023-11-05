// Copyright 2023 The Connect Authors
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

package referenceserver

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"

	"connectrpc.com/conformance/internal"
	"connectrpc.com/conformance/internal/compression"
	"connectrpc.com/conformance/internal/gen/proto/connect/connectrpc/conformance/v1alpha1/conformancev1alpha1connect"
	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	connect "connectrpc.com/connect"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Run runs the server according to server config read from the 'in' reader.
func Run(_ context.Context, _ []string, inReader io.ReadCloser, outWriter io.WriteCloser) error {
	json := flag.Bool("json", false, "whether to use the JSON format for marshaling / unmarshaling messages")
	host := flag.String("host", internal.DefaultHost, "the host for the conformance server")
	port := flag.String("port", internal.DefaultPort, "the port for the conformance server ")

	flag.Parse()

	// Read the server config from  the in reader
	data, err := io.ReadAll(inReader)
	if err != nil {
		return err
	}

	codec := internal.NewCodec(*json)

	// Unmarshal into a ServerCompatRequest
	req := &v1alpha1.ServerCompatRequest{}
	if err := codec.Unmarshal(data, req); err != nil {
		return err
	}

	// Create an HTTP server based on the request
	server, err := createServer(req)
	if err != nil {
		return err
	}

	// Create a listener for the server so that we are able to obtain
	// the IP and port for publishing on the out writer
	listener, err := net.Listen("tcp", net.JoinHostPort(*host, *port))
	if err != nil {
		return err
	}
	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return errors.New("unable to determine tcp address from listener")
	}

	resp := &v1alpha1.ServerCompatResponse{
		Host: fmt.Sprint(tcpAddr.IP),
		Port: uint32(tcpAddr.Port),
	}
	bytes, err := codec.Marshal(resp)
	if err != nil {
		return err
	}
	if _, err := outWriter.Write(bytes); err != nil {
		return err
	}

	// Finally, start the server
	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

// Creates an HTTP server using the provided ServerCompatRequest.
func createServer(req *v1alpha1.ServerCompatRequest) (*http.Server, error) {
	mux := http.NewServeMux()
	mux.Handle(conformancev1alpha1connect.NewConformanceServiceHandler(
		&conformanceServer{},
		connect.WithCompression(compression.Brotli, compression.NewBrotliDecompressor, compression.NewBrotliCompressor),
		connect.WithCompression(compression.Deflate, compression.NewDeflateDecompressor, compression.NewDeflateCompressor),
		connect.WithCompression(compression.Snappy, compression.NewSnappyDecompressor, compression.NewSnappyCompressor),
		connect.WithCompression(compression.Zstd, compression.NewZstdDecompressor, compression.NewZstdCompressor),
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
		return nil, errors.New("an HTTP version must be specified")
	}

	return server, nil
}

// Create a new HTTP/1.1 server.
func newH1Server(handler http.Handler) *http.Server {
	h1Server := &http.Server{ //nolint:gosec
		Addr:    net.JoinHostPort(internal.DefaultHost, internal.DefaultPort),
		Handler: handler,
	}
	return h1Server
}

// Create a new HTTP/2 server.
func newH2Server(handler http.Handler) *http.Server {
	h2Server := &http.Server{ //nolint:gosec
		Addr: net.JoinHostPort(internal.DefaultHost, internal.DefaultPort),
	}
	h2Server.Handler = h2c.NewHandler(handler, &http2.Server{})
	return h2Server
}
