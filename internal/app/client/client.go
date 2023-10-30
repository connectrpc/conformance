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

package client

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"connectrpc.com/conformance/internal/app"
	"connectrpc.com/conformance/internal/gen/proto/connect/connectrpc/conformance/v1alpha1/conformancev1alpha1connect"
	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"connectrpc.com/connect"
	"golang.org/x/net/http2"
)

// Run runs the client according to a client config read from the 'in' reader. The result of the run
// is written to the 'out' writer, including any errors encountered during the actual run. Any error
// returned from this function is indicative of an issue with the reader or writer and should not be related
// to the actual run.
func Run(ctx context.Context, _ []string, inReader io.ReadCloser, outWriter io.WriteCloser) error {
	json := flag.Bool("json", false, "whether to use the JSON format for marshaling / unmarshaling messages")

	flag.Parse()

	// Read the server config from  the in reader
	// TODO - This should be able to read many compatrequests and make sure that whatever format this
	// reads is the same as the format that the test runner actually writes.
	data, err := io.ReadAll(inReader)
	if err != nil {
		return err
	}

	codec := app.NewCodec(*json)

	req := &v1alpha1.ClientCompatRequest{}
	if err := codec.Unmarshal(data, req); err != nil {
		return err
	}

	result, err := invoke(ctx, req)

	// Build the result for the out writer.
	resp := &v1alpha1.ClientCompatResponse{
		TestName: req.TestName,
	}
	// If an error was returned, it was a runtime / unexpected internal error so
	// the written response should contain an error result, not a response with
	// any RPC information
	if err != nil {
		resp.Result = &v1alpha1.ClientCompatResponse_Error{
			Error: &v1alpha1.ClientErrorResult{
				Message: err.Error(),
			},
		}
	} else {
		resp.Result = &v1alpha1.ClientCompatResponse_Response{
			Response: result,
		}
	}

	// Marshal the response and write the output
	bytes, err := codec.Marshal(resp)
	if err != nil {
		return err
	}
	if _, err := outWriter.Write(bytes); err != nil {
		return err
	}

	return nil
}

// Invokes a ClientCompatRequest, returning either the result of the invocation or an error. The error
// returned from this function indicates a runtime/unexpected internal error and is not indicative of a
// Connect error returned from calling an RPC. Any error (i.e. a Connect error) that _is_ returned from
// the actual RPC invocation will be present in the returned ClientResponseResult.
func invoke(ctx context.Context, req *v1alpha1.ClientCompatRequest) (*v1alpha1.ClientResponseResult, error) {
	var scheme string
	if req.ServerTlsCert != nil {
		scheme = "https://"
	} else {
		scheme = "http://"
	}
	urlString := scheme + net.JoinHostPort(req.Host, strconv.FormatUint(uint64(req.Port), 10))
	serverURL, err := url.ParseRequestURI(urlString)
	if err != nil {
		return nil, errors.New("invalid url: %s" + urlString)
	}

	// TODO - We should cache the transports here so that we're not creating one for each
	// test case
	var transport http.RoundTripper
	switch req.HttpVersion {
	case v1alpha1.HTTPVersion_HTTP_VERSION_1:
		transport = &http.Transport{}
	case v1alpha1.HTTPVersion_HTTP_VERSION_2:
		transport = &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		}
	case v1alpha1.HTTPVersion_HTTP_VERSION_3:
		return nil, errors.New("HTTP/3 is not yet supported")
	case v1alpha1.HTTPVersion_HTTP_VERSION_UNSPECIFIED:
		return nil, errors.New("an HTTP version must be specified")
	}

	// Create client options based on protocol of the implementation
	clientOptions := []connect.ClientOption{connect.WithHTTPGet()}
	switch req.Protocol {
	case v1alpha1.Protocol_PROTOCOL_GRPC:
		clientOptions = append(clientOptions, connect.WithGRPC())
	case v1alpha1.Protocol_PROTOCOL_GRPC_WEB:
		clientOptions = append(clientOptions, connect.WithGRPCWeb())
	case v1alpha1.Protocol_PROTOCOL_CONNECT:
		// Do nothing
	case v1alpha1.Protocol_PROTOCOL_UNSPECIFIED:
		return nil, errors.New("a protocol must be specified")
	}

	if req.Codec == v1alpha1.Codec_CODEC_JSON {
		clientOptions = append(clientOptions, connect.WithProtoJSON())
	}

	// TODO - Add support for other compression algos
	switch req.Compression {
	case v1alpha1.Compression_COMPRESSION_GZIP:
		clientOptions = append(clientOptions, connect.WithSendGzip())
	case v1alpha1.Compression_COMPRESSION_BR, v1alpha1.Compression_COMPRESSION_ZSTD,
		v1alpha1.Compression_COMPRESSION_DEFLATE, v1alpha1.Compression_COMPRESSION_SNAPPY:
		return nil, errors.New(req.Compression.String() + " is not yet supported")
	case v1alpha1.Compression_COMPRESSION_IDENTITY, v1alpha1.Compression_COMPRESSION_UNSPECIFIED:
		// Do nothing
	}

	switch req.Service {
	case conformancev1alpha1connect.ConformanceServiceName:
		return newInvoker(transport, serverURL, clientOptions).Invoke(ctx, req)
	default:
		return nil, errors.New("service name " + req.Service + " is not a valid service")
	}
}