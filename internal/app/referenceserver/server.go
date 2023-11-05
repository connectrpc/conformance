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
	"time"

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
<<<<<<< HEAD:internal/app/referenceserver/server.go
func Run(_ context.Context, _ []string, inReader io.ReadCloser, outWriter io.WriteCloser) error {
	json := flag.Bool("json", false, "whether to use the JSON format for marshaling / unmarshaling messages")
	host := flag.String("host", internal.DefaultHost, "the host for the conformance server")
	port := flag.String("port", internal.DefaultPort, "the port for the conformance server ")
=======
func Run(ctx context.Context, args []string, inReader io.ReadCloser, outWriter, errWriter io.WriteCloser) error {
	_ = errWriter // TODO: send out-of-band messages about test cases to this writer

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	json := flags.Bool("json", false, "whether to use the JSON format for marshaling / unmarshaling messages")
	host := flags.String("host", defaultHost, "the host for the conformance server")
	port := flags.String("port", defaultPort, "the port for the conformance server ")
>>>>>>> v2:internal/app/server/server.go

	_ = flags.Parse(args[1:])
	if flags.NArg() != 0 {
		return errors.New("this command does not accept any positional arguments")
	}

	codec := internal.NewCodec(*json)

	// Read the server config from  the in reader
	req := &v1alpha1.ServerCompatRequest{}
	if err := codec.NewDecoder(inReader).DecodeNext(req); err != nil {
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
	if err := codec.NewEncoder(outWriter).Encode(resp); err != nil {
		return err
	}

	// Finally, start the server
	var serveError error
	serveDone := make(chan struct{})
	go func() {
		defer close(serveDone)
		serveError = server.Serve(listener)
	}()
	select {
	case <-serveDone:
		return serveError
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		//nolint:contextcheck // we use context.Background() because ctx is already done
		return server.Shutdown(shutdownCtx)
	}
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
