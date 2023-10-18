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
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"connectrpc.com/conformance/internal/gen/proto/connect/connectrpc/conformance/v1alpha1/conformancev1alpha1connect"
	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/encoding/protojson"
	proto "google.golang.org/protobuf/proto"
)

const (
	// The default host to use for the server
	defaultHost = "127.0.0.1"
	// The default port to use for the server. We choose 0 so that
	// an ephemeral port is selected by the OS
	defaultPort = "0"
)

func Run(ctx context.Context, args []string, in io.ReadCloser, out, err io.WriteCloser) error {
	rdr := bufio.NewReader(in)

	var data string
	// TODO - How should we read from 'in'? Using ReadString or ReadBytes?
	// Assuming ReadBytes, but I used ReadString so i could test by redirecting a file
	// to stdin with JSON inside
	for {
		bytes, err := rdr.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		data = bytes
	}
	req := &v1alpha1.ServerCompatRequest{}
	if err := protojson.Unmarshal([]byte(data), req); err != nil {
		return err
	}

	resp, serverErr := startServer(req)
	if serverErr != nil {
		return serverErr
	}

	respBytes, marshalErr := proto.Marshal(resp)
	if marshalErr != nil {
		return marshalErr
	}
	out.Write(respBytes)

	return nil
}

func startServer(req *v1alpha1.ServerCompatRequest) (*v1alpha1.ServerCompatResponse, error) {
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
	if req.HttpVersion == v1alpha1.HTTPVersion_HTTP_VERSION_1 {
		server = newH1Server(corsHandler)
	} else if req.HttpVersion == v1alpha1.HTTPVersion_HTTP_VERSION_2 {
		server = newH2Server(mux)
	} else if req.HttpVersion == v1alpha1.HTTPVersion_HTTP_VERSION_3 {
		return nil, errors.New("HTTP/3 is not yet supported")
	} else {
		return nil, errors.New("an HTTP version must be specifed.")
	}
	ln, err := net.Listen("tcp", net.JoinHostPort(defaultHost, defaultPort))
	if err != nil {
		return nil, err
	}
	resp := &v1alpha1.ServerCompatResponse{}
	errs := make(chan error, 1)
	go func() {
		err := server.Serve(ln)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()

	// TODO - Is there a race condition here where we reach the default case before any error is sent
	// on the errs channel?
	select {
	case err := <-errs:
		errResult := &v1alpha1.ServerCompatResponse_Error{
			Error: &v1alpha1.ServerErrorResult{
				Message: err.Error(),
			},
		}
		resp.Result = errResult
	default:
		result := &v1alpha1.ServerCompatResponse_Listening{
			Listening: &v1alpha1.ServerListeningResult{
				Host: fmt.Sprint(ln.Addr().(*net.TCPAddr).IP),
				Port: fmt.Sprint(ln.Addr().(*net.TCPAddr).Port),
			},
		}
		resp.Result = result
	}

	return resp, nil
}

func newH1Server(handler http.Handler) *http.Server {
	h1Server := &http.Server{
		Addr:    net.JoinHostPort(defaultHost, defaultPort),
		Handler: handler,
	}
	return h1Server
}

func newH2Server(handler http.Handler) *http.Server {
	h2Server := &http.Server{
		Addr: net.JoinHostPort(defaultHost, defaultPort),
	}
	h2Server.Handler = h2c.NewHandler(handler, &http2.Server{})
	return h2Server
}
