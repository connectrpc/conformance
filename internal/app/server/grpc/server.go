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

package grpcserver

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"connectrpc.com/conformance/internal/app"
	conformancev1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"google.golang.org/grpc"
)

const (
	// The default host to use for the server.
	defaultHost = "127.0.0.1"
	// The default port to use for the server. We choose 0 so that
	// an ephemeral port is selected by the OS if no port is specified.
	defaultPort = "0"
)

// Run runs the server according to server config read from the 'in' reader.
func Run(_ context.Context, _ []string, inReader io.ReadCloser, outWriter io.WriteCloser) error {
	json := flag.Bool("json", false, "whether to use the JSON format for marshaling / unmarshaling messages")
	host := flag.String("host", defaultHost, "the host for the conformance server")
	port := flag.String("port", defaultPort, "the port for the conformance server ")

	flag.Parse()

	// Read the server config from  the in reader
	data, err := io.ReadAll(inReader)
	if err != nil {
		return err
	}

	codec := app.NewCodec(*json)

	// Unmarshal into a ServerCompatRequest
	req := &v1alpha1.ServerCompatRequest{}
	if err := codec.Unmarshal(data, req); err != nil {
		return err
	}

	// Create an HTTP server based on the request
	server, err := run()
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

func run() (*grpc.Server, error) {
	lis, err := net.Listen("tcp", ":"+defaultPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()
	conformancev1alpha1.RegisterConformanceServiceServer(server, NewConformanceServiceServer())
	return server, nil
}
