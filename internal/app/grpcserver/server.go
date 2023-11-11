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
	"net"
	"strconv"

	"connectrpc.com/conformance/internal"
	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip" // enables GZIP compression w/ gRPC
)

// Run runs the server according to server config read from the 'in' reader.
func Run(ctx context.Context, args []string, inReader io.ReadCloser, outWriter io.WriteCloser, _ io.WriteCloser) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	json := flags.Bool("json", false, "whether to use the JSON format for marshaling / unmarshaling messages")
	host := flags.String("host", internal.DefaultHost, "the host for the conformance server")
	port := flags.Int("port", internal.DefaultPort, "the port for the conformance server ")

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

	// Create a gRPC server based on the request
	server, err := createServer()
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
		server.GracefulStop()
		return nil
	}
}

func createServer() (*grpc.Server, error) { //nolint:unparam
	server := grpc.NewServer()
	v1alpha1.RegisterConformanceServiceServer(server, NewConformanceServiceServer())
	return server, nil
}
