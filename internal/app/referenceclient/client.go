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

package referenceclient

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

	"connectrpc.com/conformance/internal"
	"connectrpc.com/conformance/internal/compression"
	"connectrpc.com/conformance/internal/gen/proto/connect/connectrpc/conformance/v1alpha1/conformancev1alpha1connect"
	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"connectrpc.com/connect"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/net/http2"
)

// Run runs the client according to a client config read from the 'in' reader. The result of the run
// is written to the 'out' writer, including any errors encountered during the actual run. Any error
// returned from this function is indicative of an issue with the reader or writer and should not be related
// to the actual run.
func Run(ctx context.Context, args []string, inReader io.ReadCloser, outWriter, _ io.WriteCloser) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	json := flags.Bool("json", false, "whether to use the JSON format for marshaling / unmarshaling messages")

	_ = flags.Parse(args[1:])
	if flags.NArg() != 0 {
		return errors.New("this command does not accept any positional arguments")
	}

	codec := internal.NewCodec(*json)
	decoder := codec.NewDecoder(inReader)
	encoder := codec.NewEncoder(outWriter)

	for {
		var req v1alpha1.ClientCompatRequest
		err := decoder.DecodeNext(&req)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		result, err := invoke(ctx, &req)

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
		if err := encoder.Encode(resp); err != nil {
			return err
		}
	}
}

// Invokes a ClientCompatRequest, returning either the result of the invocation or an error. The error
// returned from this function indicates a runtime/unexpected internal error and is not indicative of a
// Connect error returned from calling an RPC. Any error (i.e. a Connect error) that _is_ returned from
// the actual RPC invocation will be present in the returned ClientResponseResult.
func invoke(ctx context.Context, req *v1alpha1.ClientCompatRequest) (*v1alpha1.ClientResponseResult, error) {
	tlsConf, err := createTLSConfig(req)
	if err != nil {
		return nil, err
	}
	var scheme string
	if tlsConf != nil {
		scheme = "https://"
	} else {
		scheme = "http://"
	}
	urlString := scheme + net.JoinHostPort(req.Host, strconv.Itoa(int(req.Port)))
	serverURL, err := url.ParseRequestURI(urlString)
	if err != nil {
		return nil, errors.New("invalid url: %s" + urlString)
	}

	// TODO - We should cache the transports here so that we're not creating one for each
	// test case
	var transport http.RoundTripper
	switch req.HttpVersion {
	case v1alpha1.HTTPVersion_HTTP_VERSION_1:
		if tlsConf != nil {
			tlsConf.NextProtos = []string{"http/1.1"}
		}
		transport = &http.Transport{
			DisableCompression: true,
			TLSClientConfig:    tlsConf,
		}
	case v1alpha1.HTTPVersion_HTTP_VERSION_2:
		if tlsConf != nil {
			tlsConf.NextProtos = []string{"h2"}
			transport = &http.Transport{
				DisableCompression: true,
				TLSClientConfig:    tlsConf,
				ForceAttemptHTTP2:  true,
			}
		} else {
			transport = &http2.Transport{
				DisableCompression: true,
				AllowHTTP:          true,
				DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
					return (&net.Dialer{}).DialContext(ctx, network, addr)
				},
			}
		}
	case v1alpha1.HTTPVersion_HTTP_VERSION_3:
		if tlsConf == nil {
			return nil, errors.New("HTTP/3 indicated in request but no TLS info provided")
		}
		transport = &http3.RoundTripper{
			DisableCompression: true,
			EnableDatagrams:    true,
			TLSClientConfig:    tlsConf,
		}
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

	switch req.Compression {
	case v1alpha1.Compression_COMPRESSION_BR:
		clientOptions = append(
			clientOptions,
			connect.WithAcceptCompression(
				compression.Brotli,
				compression.NewBrotliDecompressor,
				compression.NewBrotliCompressor,
			),
			connect.WithSendCompression(compression.Brotli),
		)
	case v1alpha1.Compression_COMPRESSION_DEFLATE:
		clientOptions = append(
			clientOptions,
			connect.WithAcceptCompression(
				compression.Deflate,
				compression.NewDeflateDecompressor,
				compression.NewDeflateCompressor,
			),
			connect.WithSendCompression(compression.Deflate),
		)
	case v1alpha1.Compression_COMPRESSION_GZIP:
		// Connect clients send uncompressed requests and ask for gzipped responses by default
		// As a result, specifying a compression of gzip for a client indicates it should also
		// send gzipped requests
		clientOptions = append(clientOptions, connect.WithSendGzip())
	case v1alpha1.Compression_COMPRESSION_SNAPPY:
		clientOptions = append(
			clientOptions,
			connect.WithAcceptCompression(
				compression.Snappy,
				compression.NewSnappyDecompressor,
				compression.NewSnappyCompressor,
			),
			connect.WithSendCompression(compression.Snappy),
		)
	case v1alpha1.Compression_COMPRESSION_ZSTD:
		clientOptions = append(
			clientOptions,
			connect.WithAcceptCompression(
				compression.Zstd,
				compression.NewZstdDecompressor,
				compression.NewZstdCompressor,
			),
			connect.WithSendCompression(compression.Zstd),
		)
	case v1alpha1.Compression_COMPRESSION_IDENTITY, v1alpha1.Compression_COMPRESSION_UNSPECIFIED:
		// Do nothing
	}

	if req.MessageReceiveLimit > 0 {
		clientOptions = append(clientOptions, connect.WithReadMaxBytes(int(req.MessageReceiveLimit)))
	}

	switch req.Service {
	case conformancev1alpha1connect.ConformanceServiceName:
		return newInvoker(transport, serverURL, clientOptions).Invoke(ctx, req)
	default:
		return nil, errors.New("service name " + req.Service + " is not a valid service")
	}
}

func createTLSConfig(req *v1alpha1.ClientCompatRequest) (*tls.Config, error) {
	if req.ServerTlsCert == nil {
		if req.ClientTlsCreds != nil {
			return nil, errors.New("request indicated TLS client credentials but not server TLS cert provided")
		}
		return nil, nil //nolint:nilnil
	}
	return internal.NewClientTLSConfig(req.ServerTlsCert, req.ClientTlsCreds.GetCert(), req.ClientTlsCreds.GetKey())
}
