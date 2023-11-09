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
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"connectrpc.com/conformance/internal"
	"connectrpc.com/conformance/internal/compression"
	"connectrpc.com/conformance/internal/gen/proto/connect/connectrpc/conformance/v1alpha1/conformancev1alpha1connect"
	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	connect "connectrpc.com/connect"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Run runs the server according to server config read from the 'in' reader.
func Run(ctx context.Context, args []string, inReader io.ReadCloser, outWriter, errWriter io.WriteCloser) error {
	_ = errWriter // TODO: send out-of-band messages about test cases to this writer

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	json := flags.Bool("json", false, "whether to use the JSON format for marshaling / unmarshaling messages")
	host := flags.String("host", internal.DefaultHost, "the host for the conformance server")
	port := flags.Int("port", internal.DefaultPort, "the port for the conformance server ")

	_ = flags.Parse(args[1:])
	if flags.NArg() != 0 {
		return errors.New("this command does not accept any positional arguments")
	}

	codec := internal.NewCodec(*json)

	// Read the server config from the in reader
	req := &v1alpha1.ServerCompatRequest{}
	if err := codec.NewDecoder(inReader).DecodeNext(req); err != nil {
		return err
	}

	// Create an HTTP server based on the request
	server, certBytes, err := createServer(req, net.JoinHostPort(*host, strconv.Itoa(*port)))
	if err != nil {
		return err
	}

	actualHost, actualPortStr, err := net.SplitHostPort(server.Addr())
	if err != nil {
		return err
	}
	actualPort, err := strconv.Atoi(actualPortStr)
	if err != nil {
		return err
	}

	resp := &v1alpha1.ServerCompatResponse{
		Host:    actualHost,
		Port:    uint32(actualPort),
		PemCert: certBytes,
	}
	if err := codec.NewEncoder(outWriter).Encode(resp); err != nil {
		return err
	}

	// Finally, start the server
	var serveError error
	serveDone := make(chan struct{})
	go func() {
		defer close(serveDone)
		serveError = server.Serve()
	}()
	select {
	case <-serveDone:
		return serveError
	case <-ctx.Done():
		return server.GracefulShutdown(5 * time.Second)
	}
}

type httpServer interface {
	Serve() error
	GracefulShutdown(time.Duration) error
	Addr() string
}

type stdHTTPServer struct {
	svr *http.Server
	lis net.Listener
}

func (s *stdHTTPServer) Serve() error {
	return s.svr.Serve(s.lis)
}

func (s *stdHTTPServer) GracefulShutdown(period time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), period)
	defer cancel()
	return s.svr.Shutdown(ctx)
}

func (s *stdHTTPServer) Addr() string {
	return s.lis.Addr().String()
}

// Creates an HTTP server using the provided ServerCompatRequest.
func createServer(req *v1alpha1.ServerCompatRequest, listenAddr string) (httpServer, []byte, error) {
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
	var tlsConf *tls.Config
	var certBytes []byte
	if req.UseTls {
		var cert tls.Certificate
		var err error
		cert, certBytes, err = internal.NewServerCert()
		if err != nil {
			return nil, nil, fmt.Errorf("could not generate TLS cert: %w", err)
		}
		tlsConf, err = internal.NewServerTLSConfig(cert, tls.NoClientCert, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("could not create TLS configuration: %w", err)
		}
	}
	var server httpServer
	var err error
	switch req.HttpVersion {
	case v1alpha1.HTTPVersion_HTTP_VERSION_1:
		server, err = newH1Server(corsHandler, listenAddr, tlsConf)
	case v1alpha1.HTTPVersion_HTTP_VERSION_2:
		// TODO: Should we support CORS over HTTP/2, too?
		server, err = newH2Server(mux, listenAddr, tlsConf)
	case v1alpha1.HTTPVersion_HTTP_VERSION_3:
		server, err = newH3Server(mux, listenAddr, tlsConf)
	case v1alpha1.HTTPVersion_HTTP_VERSION_UNSPECIFIED:
		err = errors.New("an HTTP version must be specified")
	}
	if err != nil {
		return nil, nil, err
	}

	return server, certBytes, nil
}

// newH1Server creates a new HTTP/1.1 server.
func newH1Server(handler http.Handler, listenAddr string, tlsConf *tls.Config) (httpServer, error) {
	if tlsConf != nil {
		tlsConf.NextProtos = []string{"http/1.1"}
	}
	h1Server := &http.Server{
		Addr:              listenAddr,
		Handler:           handler,
		TLSConfig:         tlsConf,
		ReadHeaderTimeout: 5 * time.Second,
	}
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	return &stdHTTPServer{svr: h1Server, lis: lis}, nil
}

// newH2Server creates a new HTTP/2 server.
func newH2Server(handler http.Handler, listenAddr string, tlsConf *tls.Config) (httpServer, error) {
	if tlsConf != nil {
		tlsConf.NextProtos = []string{"h2"}
	} else {
		handler = h2c.NewHandler(handler, &http2.Server{})
	}
	h2Server := &http.Server{
		Addr:              listenAddr,
		Handler:           handler,
		TLSConfig:         tlsConf,
		ReadHeaderTimeout: 5 * time.Second,
	}
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	return &stdHTTPServer{svr: h2Server, lis: lis}, nil
}

// Create a new HTTP/3 server.
func newH3Server(handler http.Handler, listenAddr string, tlsConf *tls.Config) (httpServer, error) {
	if tlsConf == nil {
		return nil, errors.New("request indicated HTTP/3 without TLS, which is not possible")
	}
	tlsConf = http3.ConfigureTLSConfig(tlsConf)
	h3Server := &http3.Server{
		Addr:      listenAddr,
		Handler:   handler,
		TLSConfig: tlsConf,
	}
	lis, err := quic.ListenAddrEarly(listenAddr, tlsConf, &quic.Config{Allow0RTT: true, EnableDatagrams: true})
	if err != nil {
		return nil, err
	}
	return &http3Server{svr: h3Server, lis: lis}, nil
}

type http3Server struct {
	svr *http3.Server
	lis http3.QUICEarlyListener
}

func (s *http3Server) Serve() error {
	return s.svr.ServeListener(s.lis)
}

func (s *http3Server) GracefulShutdown(period time.Duration) error {
	return s.svr.CloseGracefully(period)
}

func (s *http3Server) Addr() string {
	return s.lis.Addr().String()
}
