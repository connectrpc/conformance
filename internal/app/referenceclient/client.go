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
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"connectrpc.com/conformance/internal"
	"connectrpc.com/conformance/internal/compression"
	v1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1/conformancev1connect"
	"connectrpc.com/connect"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/net/http2"
	"golang.org/x/sync/semaphore"
)

// Run runs the client according to a client config read from the 'in' reader. The result of the run
// is written to the 'out' writer, including any errors encountered during the actual run. Any error
// returned from this function is indicative of an issue with the reader or writer and should not be related
// to the actual run.
func Run(ctx context.Context, args []string, inReader io.ReadCloser, outWriter, _ io.WriteCloser) (retErr error) {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	json := flags.Bool("json", false, "whether to use the JSON format for marshaling / unmarshaling messages")
	parallel := flags.Uint("p", uint(runtime.GOMAXPROCS(0))*4, "the number of parallel RPCs to issue")
	showVersion := flags.Bool("version", false, "show version and exit")

	_ = flags.Parse(args[1:])
	if *showVersion {
		_, _ = fmt.Fprintf(outWriter, "%s %s\n", filepath.Base(args[0]), internal.Version)
		return nil
	}
	if flags.NArg() != 0 {
		return errors.New("this command does not accept any positional arguments")
	}
	if *parallel == 0 {
		return errors.New("invalid parallelism; must be greater than zero")
	}

	codec := internal.NewCodec(*json)
	decoder := codec.NewDecoder(inReader)
	encoder := codec.NewEncoder(outWriter)
	var encoderMu sync.Mutex

	var failure atomic.Pointer[error]
	defer func() {
		// if we're about to return nil error, but a goroutine reported
		// a failure, return that failure as the error
		if errPtr := failure.Load(); errPtr != nil && retErr == nil {
			retErr = *errPtr
		}
	}()

	var wg sync.WaitGroup
	defer wg.Wait()
	sema := semaphore.NewWeighted(int64(*parallel))

	for {
		var req v1.ClientCompatRequest
		err := decoder.DecodeNext(&req)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		if err := sema.Acquire(ctx, 1); err != nil {
			return err
		}
		if errPtr := failure.Load(); errPtr != nil {
			// If there's already been a terminal failure, don't spawn
			// anymore goroutines.
			return *errPtr
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer sema.Release(1)

			result, err := invoke(ctx, &req)

			// Build the result for the out writer.
			resp := &v1.ClientCompatResponse{
				TestName: req.TestName,
			}
			// If an error was returned, it was a runtime / unexpected internal error so
			// the written response should contain an error result, not a response with
			// any RPC information
			if err != nil {
				resp.Result = &v1.ClientCompatResponse_Error{
					Error: &v1.ClientErrorResult{
						Message: err.Error(),
					},
				}
			} else {
				resp.Result = &v1.ClientCompatResponse_Response{
					Response: result,
				}
			}

			// Marshal the response and write the output
			func() {
				encoderMu.Lock()
				defer encoderMu.Unlock()
				if err := encoder.Encode(resp); err != nil {
					failure.CompareAndSwap(nil, &err)
				}
			}()
		}()
	}
}

// Invokes a ClientCompatRequest, returning either the result of the invocation or an error. The error
// returned from this function indicates a runtime/unexpected internal error and is not indicative of a
// Connect error returned from calling an RPC. Any error (i.e. a Connect error) that _is_ returned from
// the actual RPC invocation will be present in the returned ClientResponseResult.
func invoke(ctx context.Context, req *v1.ClientCompatRequest) (*v1.ClientResponseResult, error) {
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
	case v1.HTTPVersion_HTTP_VERSION_1:
		if tlsConf != nil {
			tlsConf.NextProtos = []string{"http/1.1"}
		}
		transport = &http.Transport{
			DisableCompression: true,
			TLSClientConfig:    tlsConf,
		}
	case v1.HTTPVersion_HTTP_VERSION_2:
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
	case v1.HTTPVersion_HTTP_VERSION_3:
		if tlsConf == nil {
			return nil, errors.New("HTTP/3 indicated in request but no TLS info provided")
		}
		transport = &contextFixTransport{http3.RoundTripper{
			DisableCompression: true,
			TLSClientConfig:    tlsConf,
			QuicConfig:         &quic.Config{MaxIdleTimeout: 20 * time.Second, KeepAlivePeriod: 5 * time.Second},
		}}
	case v1.HTTPVersion_HTTP_VERSION_UNSPECIFIED:
		return nil, errors.New("an HTTP version must be specified")
	}

	// Create client options based on protocol of the implementation
	clientOptions := []connect.ClientOption{connect.WithHTTPGet()}
	switch req.Protocol {
	case v1.Protocol_PROTOCOL_GRPC:
		clientOptions = append(clientOptions, connect.WithGRPC())
	case v1.Protocol_PROTOCOL_GRPC_WEB:
		clientOptions = append(clientOptions, connect.WithGRPCWeb())
	case v1.Protocol_PROTOCOL_CONNECT:
		// Do nothing
	case v1.Protocol_PROTOCOL_UNSPECIFIED:
		return nil, errors.New("a protocol must be specified")
	}

	switch req.Codec {
	case v1.Codec_CODEC_PROTO:
		// this is the default, no option needed
	case v1.Codec_CODEC_JSON:
		clientOptions = append(clientOptions, connect.WithProtoJSON())
	case v1.Codec_CODEC_TEXT:
		clientOptions = append(clientOptions, connect.WithCodec(&internal.TextConnectCodec{}))
	default:
		return nil, errors.New("a codec must be specified")
	}

	switch req.Compression {
	case v1.Compression_COMPRESSION_BR:
		clientOptions = append(
			clientOptions,
			connect.WithAcceptCompression(
				compression.Brotli,
				compression.NewBrotliDecompressor,
				compression.NewBrotliCompressor,
			),
			connect.WithSendCompression(compression.Brotli),
		)
	case v1.Compression_COMPRESSION_DEFLATE:
		clientOptions = append(
			clientOptions,
			connect.WithAcceptCompression(
				compression.Deflate,
				compression.NewDeflateDecompressor,
				compression.NewDeflateCompressor,
			),
			connect.WithSendCompression(compression.Deflate),
		)
	case v1.Compression_COMPRESSION_GZIP:
		// Connect clients send uncompressed requests and ask for gzipped responses by default
		// As a result, specifying a compression of gzip for a client indicates it should also
		// send gzipped requests
		clientOptions = append(clientOptions, connect.WithSendGzip())
	case v1.Compression_COMPRESSION_SNAPPY:
		clientOptions = append(
			clientOptions,
			connect.WithAcceptCompression(
				compression.Snappy,
				compression.NewSnappyDecompressor,
				compression.NewSnappyCompressor,
			),
			connect.WithSendCompression(compression.Snappy),
		)
	case v1.Compression_COMPRESSION_ZSTD:
		clientOptions = append(
			clientOptions,
			connect.WithAcceptCompression(
				compression.Zstd,
				compression.NewZstdDecompressor,
				compression.NewZstdCompressor,
			),
			connect.WithSendCompression(compression.Zstd),
		)
	case v1.Compression_COMPRESSION_IDENTITY, v1.Compression_COMPRESSION_UNSPECIFIED:
		// Do nothing
	}

	if req.MessageReceiveLimit > 0 {
		clientOptions = append(clientOptions, connect.WithReadMaxBytes(int(req.MessageReceiveLimit)))
	}

	switch req.Service {
	case conformancev1connect.ConformanceServiceName:
		return newInvoker(transport, serverURL, clientOptions).Invoke(ctx, req)
	default:
		return nil, errors.New("service name " + req.Service + " is not a valid service")
	}
}

func createTLSConfig(req *v1.ClientCompatRequest) (*tls.Config, error) {
	if req.ServerTlsCert == nil {
		if req.ClientTlsCreds != nil {
			return nil, errors.New("request indicated TLS client credentials but not server TLS cert provided")
		}
		return nil, nil //nolint:nilnil
	}
	return internal.NewClientTLSConfig(req.ServerTlsCert, req.ClientTlsCreds.GetCert(), req.ClientTlsCreds.GetKey())
}

// contextFixTransport wraps an HTTP/3 transport so that context errors can be correctly
// classified by the connect-go framework. This is a work-around until a fix
// can be implemented in connect-go and/or quic-go.
// See: https://github.com/quic-go/quic-go/issues/4196
type contextFixTransport struct {
	http3.RoundTripper
}

func (t *contextFixTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	resp, err := t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, maybeWrapContextError(ctx, err)
	}
	resp.Body = &contextFixReader{ctx: ctx, r: resp.Body}
	return resp, nil
}

type contextFixReader struct {
	ctx context.Context
	r   io.ReadCloser
}

func (r *contextFixReader) Read(data []byte) (int, error) {
	n, err := r.r.Read(data)
	return n, maybeWrapContextError(r.ctx, err)
}

func (r *contextFixReader) Close() error {
	return maybeWrapContextError(r.ctx, r.r.Close())
}

func maybeWrapContextError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	ctxErr := ctx.Err()
	if ctxErr == nil {
		return err
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return &contextFixError{timeout: true, error: err}
	}
	var httpErr *http3.Error
	if errors.As(err, &httpErr) && httpErr.ErrorCode == http3.ErrCodeRequestCanceled {
		return &contextFixError{timeout: errors.Is(ctxErr, context.DeadlineExceeded), error: err}
	}
	return err
}

type contextFixError struct {
	timeout bool
	error
}

func (e *contextFixError) Is(err error) bool {
	return (e.timeout && err == context.DeadlineExceeded) ||
		(!e.timeout && err == context.Canceled)
}
