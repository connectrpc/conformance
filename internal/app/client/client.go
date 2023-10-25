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

// Wrapper defines a wrapper around a client which can invoke requests based on a ClientCompatRequest
type Wrapper interface {
	// Invoke invokes a request according to a given config
	Invoke(context.Context, *v1alpha1.ClientCompatRequest) (*v1alpha1.ClientCompatResponse, error)
}

// Run runs the server according to server config read from the 'in' reader.
func Run(ctx context.Context, _ []string, inReader io.ReadCloser, outWriter io.WriteCloser) error {
	json := flag.Bool("json", false, "whether to use the JSON format for marshaling / unmarshaling messages")

	flag.Parse()

	// Read the server config from  the in reader
	data, err := io.ReadAll(inReader)
	if err != nil {
		return err
	}

	codec := app.NewCodec(*json)

	req := &v1alpha1.ClientCompatRequest{}
	if err := codec.Unmarshal(data, req); err != nil {
		return err
	}

	resp, err := invoke(ctx, req)
	if err != nil {
		return err
	}
	bytes, err := codec.Marshal(resp)
	if err != nil {
		return err
	}
	if _, err := outWriter.Write(bytes); err != nil {
		return err
	}

	return nil
}

func invoke(ctx context.Context, req *v1alpha1.ClientCompatRequest) (*v1alpha1.ClientCompatResponse, error) {
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

	// TODO - How do we configure each compression algo? i.e.
	// how do we know the string to use and the func for WithCompression?
	switch req.Compression {
	case v1alpha1.Compression_COMPRESSION_GZIP:
		clientOptions = append(clientOptions, connect.WithSendGzip())
	case v1alpha1.Compression_COMPRESSION_IDENTITY:
	case v1alpha1.Compression_COMPRESSION_BR:
	case v1alpha1.Compression_COMPRESSION_ZSTD:
	case v1alpha1.Compression_COMPRESSION_DEFLATE:
	case v1alpha1.Compression_COMPRESSION_SNAPPY:
	case v1alpha1.Compression_COMPRESSION_UNSPECIFIED:
		// Do nothing
	}

	var wrapper Wrapper
	switch req.Service {
	case conformancev1alpha1connect.ConformanceServiceName:
		wrapper = NewConformanceClientWrapper(transport, serverURL, clientOptions)
	default:
		return nil, errors.New("service name " + req.Service + " is not a valid service")
	}
	return wrapper.Invoke(ctx, req)
}
