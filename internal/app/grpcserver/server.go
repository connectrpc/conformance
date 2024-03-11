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

package grpcserver

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"connectrpc.com/conformance/internal"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/tracer"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip" // enables GZIP compression w/ gRPC
)

// Run runs the server according to server config read from the 'in' reader.
func Run(ctx context.Context, args []string, inReader io.ReadCloser, outWriter io.WriteCloser, errWriter io.WriteCloser) error {
	return RunWithTrace(ctx, args, inReader, outWriter, errWriter, nil)
}

// RunWithTrace is just like Run except that it can collect trace info and
// report it via the given tracer.
func RunWithTrace(ctx context.Context, args []string, inReader io.ReadCloser, outWriter io.WriteCloser, _ io.WriteCloser, trace *tracer.Tracer) error {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	json := flags.Bool("json", false, "whether to use the JSON format for marshaling / unmarshaling messages")
	host := flags.String("bind", internal.DefaultHost, "the bind address for the conformance server")
	port := flags.Int("port", internal.DefaultPort, "the port for the conformance server ")
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

	codec := internal.NewCodec(*json)

	// Read the server config from  the in reader
	req := &conformancev1.ServerCompatRequest{}
	if err := codec.NewDecoder(inReader).DecodeNext(req); err != nil {
		return err
	}
	if req.UseTls {
		return fmt.Errorf("%s: TLS is not supported", args[0])
	}
	if req.Protocol != conformancev1.Protocol_PROTOCOL_GRPC && req.Protocol != conformancev1.Protocol_PROTOCOL_GRPC_WEB {
		return fmt.Errorf("%s: protocol %s is not supported", args[0], req.Protocol)
	}
	if req.Protocol == conformancev1.Protocol_PROTOCOL_GRPC && req.HttpVersion != conformancev1.HTTPVersion_HTTP_VERSION_2 {
		return fmt.Errorf("%s: HTTP version %s is not supported with protocol %s", args[0], req.HttpVersion, req.Protocol)
	}

	// Create a gRPC server based on the request
	server, err := createServer(req.MessageReceiveLimit)
	if err != nil {
		return err
	}

	// Create a listener for the server so that we are able to obtain
	// the IP and port for publishing on the out writer
	listener, err := net.Listen("tcp", net.JoinHostPort(*host, strconv.Itoa(*port)))
	if err != nil {
		return err
	}
	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return errors.New("unable to determine tcp address from listener")
	}

	resp := &conformancev1.ServerCompatResponse{
		Host: fmt.Sprint(tcpAddr.IP),
		Port: uint32(tcpAddr.Port),
	}
	if err := codec.NewEncoder(outWriter).Encode(resp); err != nil {
		return err
	}

	// Finally, start the server
	if req.Protocol == conformancev1.Protocol_PROTOCOL_GRPC_WEB {
		return runGRPCWebServer(ctx, server, listener, trace)
	}
	return runGRPCServer(ctx, server, listener, trace)
}

func createServer(recvLimit uint32) (*grpc.Server, error) { //nolint:unparam
	server := grpc.NewServer(
		grpc.UnaryInterceptor(serverNameUnaryInterceptor),
		grpc.StreamInterceptor(serverNameStreamInterceptor),
		grpc.MaxRecvMsgSize(int(recvLimit)),
	)
	conformancev1.RegisterConformanceServiceServer(server, NewConformanceServiceServer())
	return server, nil
}

func runGRPCServer(ctx context.Context, server *grpc.Server, listener net.Listener, trace *tracer.Tracer) error {
	var serveError error
	serveDone := make(chan struct{})
	if trace != nil {
		listener = &tracingListener{Listener: listener, trace: trace}
	}
	go func() {
		defer close(serveDone)
		serveError = server.Serve(listener)
	}()
	select {
	case <-serveDone:
		return serveError
	case <-ctx.Done():
		server.GracefulStop()
		return nil
	}
}

func runGRPCWebServer(ctx context.Context, server *grpc.Server, listener net.Listener, trace *tracer.Tracer) error {
	grpcWebServer := http.Handler(grpcweb.WrapServer(server,
		// The server needs a lenient cors setup so that it can handle testing
		// browser clients.
		grpcweb.WithOriginFunc(func(string) bool {
			return true
		}),
	))
	if trace != nil {
		grpcWebServer = tracer.TracingHandler(grpcWebServer, trace)
	}

	httpServer := http.Server{
		Handler:           h2c.NewHandler(grpcWebServer, &http2.Server{}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	var serveError error
	serveDone := make(chan struct{})
	go func() {
		defer close(serveDone)
		serveError = httpServer.Serve(listener)
	}()
	select {
	case <-serveDone:
		return serveError
	case <-ctx.Done():
		shutdownContext, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		//nolint:contextcheck // intentionally using context.Background since ctx is already done
		if err := httpServer.Shutdown(shutdownContext); err != nil {
			// Graceful shutdown took too long. Do it more forcefully.
			_ = httpServer.Close()
		}
		return nil
	}
}

type tracingListener struct {
	net.Listener
	trace *tracer.Tracer
}

func (t *tracingListener) Accept() (net.Conn, error) {
	conn, err := t.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return tracer.TracingHTTP2Conn(conn, true, t.trace), nil
}
