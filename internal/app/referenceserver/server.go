// Copyright 2023-2024 The Connect Authors
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
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/conformance/internal"
	"connectrpc.com/conformance/internal/compression"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1/conformancev1connect"
	"connectrpc.com/conformance/internal/tracer"
	"connectrpc.com/connect"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Run runs the server according to server config read from the 'in' reader.
func Run(ctx context.Context, args []string, inReader io.ReadCloser, outWriter, errWriter io.WriteCloser) error {
	return run(ctx, false, args, inReader, outWriter, errWriter, nil)
}

// RunInReferenceMode is just like Run except that it performs additional checks
// that only the conformance reference server runs. These checks do not work if
// the server is run as a server under test, only when run as a reference server.
func RunInReferenceMode(ctx context.Context, args []string, inReader io.ReadCloser, outWriter, errWriter io.WriteCloser, tracer *tracer.Tracer) error {
	return run(ctx, true, args, inReader, outWriter, errWriter, tracer)
}

func run(ctx context.Context, referenceMode bool, args []string, inReader io.ReadCloser, outWriter, errWriter io.WriteCloser, tracer *tracer.Tracer) error {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	json := flags.Bool("json", false, "whether to use the JSON format for marshaling / unmarshaling messages")
	host := flags.String("bind", internal.DefaultHost, "the bind address for the conformance server")
	port := flags.Int("port", internal.DefaultPort, "the port for the conformance server")
	tlsCert := flags.String("cert", "", "the path to a PEM-encoded TLS certificate file to use instead of generating self-signed")
	tlsKey := flags.String("key", "", "the path to a PEM-encoded TLS key file to use instead of generating self-signed")
	showVersion := flags.Bool("version", false, "show version and exit")

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if *showVersion {
		_, _ = fmt.Fprintf(outWriter, "%s %s\n", filepath.Base(args[0]), internal.Version)
		return nil
	}
	if flags.NArg() != 0 {
		return errors.New("this command does not accept any positional arguments")
	}
	if (*tlsCert == "") != (*tlsKey == "") {
		return errors.New("-cert and -key must both be provided")
	}

	codec := internal.NewCodec(*json)

	// Read the server config from the in reader
	req := &conformancev1.ServerCompatRequest{}
	if err := codec.NewDecoder(inReader).DecodeNext(req); err != nil {
		return err
	}

	// Create an HTTP server based on the request
	errPrinter := internal.NewPrinter(errWriter)
	server, certBytes, err := createServer(req, net.JoinHostPort(*host, strconv.Itoa(*port)), *tlsCert, *tlsKey, referenceMode, errPrinter, tracer)
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
	if actualHost == "" || actualHost == "0.0.0.0" {
		actualHost = internal.DefaultHost
	}

	// Start the server
	var serveError error
	serveDone := make(chan struct{})
	go func() {
		defer close(serveDone)
		serveError = server.Serve()
	}()
	// Give the above goroutine a chance to start the server and potentially
	// abort if it could not be started.
	time.Sleep(200 * time.Millisecond)
	select {
	case <-serveDone:
		return serveError
	default:
	}

	resp := &conformancev1.ServerCompatResponse{
		Host:    actualHost,
		Port:    uint32(actualPort),
		PemCert: certBytes,
	}
	if err := codec.NewEncoder(outWriter).Encode(resp); err != nil {
		return err
	}

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
	if s.svr.TLSConfig != nil {
		return s.svr.ServeTLS(s.lis, "", "")
	}
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
func createServer(req *conformancev1.ServerCompatRequest, listenAddr, tlsCertFile, tlsKeyFile string, referenceMode bool, errPrinter internal.Printer, trace *tracer.Tracer) (httpServer, []byte, error) {
	mux := http.NewServeMux()
	interceptors := []connect.Interceptor{serverNameHandlerInterceptor{}}
	if referenceMode {
		interceptors = append(interceptors, rawResponseRecorder{})
	}
	opts := []connect.HandlerOption{
		connect.WithCompression(compression.Brotli, compression.NewBrotliDecompressor, compression.NewBrotliCompressor),
		connect.WithCompression(compression.Deflate, compression.NewDeflateDecompressor, compression.NewDeflateCompressor),
		connect.WithCompression(compression.Snappy, compression.NewSnappyDecompressor, compression.NewSnappyCompressor),
		connect.WithCompression(compression.Zstd, compression.NewZstdDecompressor, compression.NewZstdCompressor),
		connect.WithCodec(&internal.TextConnectCodec{}),
		connect.WithInterceptors(interceptors...),
	}
	if req.MessageReceiveLimit > 0 {
		opts = append(opts, connect.WithReadMaxBytes(int(req.MessageReceiveLimit)))
	}

	mux.Handle(conformancev1connect.NewConformanceServiceHandler(
		&conformanceServer{referenceMode: referenceMode},
		opts...,
	))
	handler := http.Handler(http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, conformancev1connect.ConformanceServiceBidiStreamProcedure) &&
			req.ProtoMajor == 1 {
			// To force support for bidirectional RPC over HTTP 1.1 (for half-duplex testing),
			// we "trick" the handler into thinking this is HTTP/2. We have to do this because
			// otherwise, connect-go refuses to handle bidi streams over HTTP 1.1.
			req.ProtoMajor, req.ProtoMinor = 2, 0
		}
		mux.ServeHTTP(respWriter, req)
	}))
	if referenceMode {
		handler = referenceServerChecks(handler, errPrinter)
		handler = rawResponder(handler)
	} else {
		// When in reference mode, checking requests from a client-under-test, we make sure that the
		// client sends a "TE: trailers" header.
		// But when NOT in reference mode, we may handle a test case defined with a "raw HTTP request",
		// sent by the reference client. We want to make sure these are all correctly formed, and neither
		// the underlying connect-go nor grpc-go implementations complain about this. So we complain
		// about it here. So if we write a raw request test case that fails to include it, we'll know
		// about it before it finds its way into a release and then is discovered when integrating with
		// the grpc-cpp implementation, for example, which has a *hard* requirement that the header is
		// present.
		orig := handler
		handler = http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
			contentType := req.Header.Get("Content-Type")
			if (contentType == grpcContentType || strings.HasPrefix(contentType, grpcContentTypePrefix)) &&
				req.Header.Get("TE") != "trailers" {
				errWriter := connect.NewErrorWriter()
				_ = errWriter.Write(respWriter, req,
					connect.NewError(connect.CodeUnknown, errors.New("missing 'TE: trailers' header")))
				return
			}
			orig.ServeHTTP(respWriter, req)
		})
	}
	if trace != nil {
		handler = tracer.TracingHandler(handler, trace)
	}
	// The server needs a lenient cors setup so that it can handle testing
	// browser clients.
	handler = cors.New(cors.Options{
		// In case TLS client certs are used.
		AllowCredentials: true,
		// If credentials are used, default "allow all origins" doesn't work since
		// it echos back "*" in the "Access-Control-Allow-Origin" header. But asterisk
		// isn't accepted by clients when credentials are used. So we have to allow
		// all this way:
		AllowOriginFunc: func(string) bool { return true },
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
	}).Handler(handler)

	// Create servers
	var tlsConf *tls.Config
	var certBytes []byte
	if req.UseTls { //nolint:nestif
		var keyBytes []byte
		var err error
		switch {
		case tlsCertFile != "":
			certBytes, err = os.ReadFile(tlsCertFile)
			if err != nil {
				return nil, nil, fmt.Errorf("could not load TLS cert: %w", err)
			}
			keyBytes, err = os.ReadFile(tlsKeyFile)
			if err != nil {
				return nil, nil, fmt.Errorf("could not load TLS key: %w", err)
			}
		case req.ServerCreds != nil:
			certBytes = req.ServerCreds.Cert
			keyBytes = req.ServerCreds.Key
		default:
			// This generally shouldn't happen. If we're using TLS, test framework should provide one we can use.
			certBytes, keyBytes, err = internal.NewServerCert()
			if err != nil {
				return nil, nil, fmt.Errorf("could not generate TLS cert: %w", err)
			}
		}
		cert, err := internal.ParseServerCert(certBytes, keyBytes)
		if err != nil {
			return nil, nil, fmt.Errorf("could not parse TLS certificate and key: %w", err)
		}
		clientCertMode := tls.NoClientCert
		if len(req.ClientTlsCert) > 0 {
			clientCertMode = tls.RequireAndVerifyClientCert
		}
		tlsConf, err = internal.NewServerTLSConfig(cert, clientCertMode, req.ClientTlsCert)
		if err != nil {
			return nil, nil, fmt.Errorf("could not create TLS configuration: %w", err)
		}
	}
	var server httpServer
	var err error
	switch req.HttpVersion {
	case conformancev1.HTTPVersion_HTTP_VERSION_1:
		server, err = newH1Server(handler, listenAddr, tlsConf)
	case conformancev1.HTTPVersion_HTTP_VERSION_2:
		server, err = newH2Server(handler, listenAddr, tlsConf)
	case conformancev1.HTTPVersion_HTTP_VERSION_3:
		server, err = newH3Server(handler, listenAddr, tlsConf)
	case conformancev1.HTTPVersion_HTTP_VERSION_UNSPECIFIED:
		err = errors.New("an HTTP version must be specified")
	}
	if err != nil {
		return nil, nil, err
	}

	return server, certBytes, nil
}

// newH1Server creates a new HTTP/1.1 server.
func newH1Server(handler http.Handler, listenAddr string, tlsConf *tls.Config) (httpServer, error) {
	h1Server := &http.Server{
		Addr:              listenAddr,
		Handler:           handler,
		TLSConfig:         tlsConf,
		ReadHeaderTimeout: 5 * time.Second,
		ErrorLog:          nopLogger(),
	}
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	return &stdHTTPServer{svr: h1Server, lis: lis}, nil
}

// newH2Server creates a new HTTP/2 server.
func newH2Server(handler http.Handler, listenAddr string, tlsConf *tls.Config) (httpServer, error) {
	if tlsConf == nil {
		handler = h2c.NewHandler(handler, &http2.Server{})
	}
	h2Server := &http.Server{
		Addr:              listenAddr,
		Handler:           handler,
		TLSConfig:         tlsConf,
		ReadHeaderTimeout: 5 * time.Second,
		ErrorLog:          nopLogger(),
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
	lis, err := quic.ListenAddrEarly(listenAddr, tlsConf, &quic.Config{MaxIdleTimeout: 20 * time.Second, KeepAlivePeriod: 5 * time.Second})
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

func (s *http3Server) GracefulShutdown(_ time.Duration) error {
	// Note: http3.Server.CloseGracefully is not actually implemented!
	// https://github.com/quic-go/quic-go/issues/153
	// So we must use the non-graceful version :(
	return s.svr.Close()
}

func (s *http3Server) Addr() string {
	return s.lis.Addr().String()
}

//nolint:forbidigo // must refer to log package in order to suppress it in net/http server
func nopLogger() *log.Logger {
	// TODO: enable logging via -v option or env variable?
	return log.New(io.Discard, "", 0)
}
