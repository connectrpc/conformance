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
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"connectrpc.com/conformance/internal"
	"connectrpc.com/conformance/internal/gen/proto/connect/connectrpc/conformance/v1alpha1/conformancev1alpha1connect"
	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"connectrpc.com/connect"
)

type invoker struct {
	client conformancev1alpha1connect.ConformanceServiceClient
}

func (i *invoker) Invoke(
	ctx context.Context,
	req *v1alpha1.ClientCompatRequest,
) (*v1alpha1.ClientResponseResult, error) {
	switch req.Method {
	case "Unary":
		if len(req.RequestMessages) != 1 {
			return nil, errors.New("unary calls must specify exactly one request message")
		}
		resp, err := i.unary(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	case "ServerStream":
		if len(req.RequestMessages) != 1 {
			return nil, errors.New("server streaming calls must specify exactly one request message")
		}
		resp, err := i.serverStream(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	case "ClientStream":
		resp, err := i.clientStream(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	case "BidiStream":
		resp, err := i.bidiStream(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	default:
		return nil, errors.New("method name " + req.Method + " does not exist")
	}
}

func (i *invoker) unary(
	ctx context.Context,
	req *v1alpha1.ClientCompatRequest,
) (*v1alpha1.ClientResponseResult, error) {
	msg := req.RequestMessages[0]
	ur := &v1alpha1.UnaryRequest{}
	if err := msg.UnmarshalTo(ur); err != nil {
		return nil, err
	}

	request := connect.NewRequest(ur)

	// Add the specified request headers to the request
	internal.AddHeaders(req.RequestHeaders, request.Header())

	var protoErr *v1alpha1.Error
	var headers []*v1alpha1.Header
	var trailers []*v1alpha1.Header
	payloads := make([]*v1alpha1.ConformancePayload, 0, 1)

	// Invoke the Unary call
	resp, err := i.client.Unary(ctx, request)
	if err != nil {
		// If an error was returned, first convert it to a Connect error
		// so that we can get the headers from the Meta property. Then,
		// convert _that_ to a proto Error so we can set it in the response.
		connectErr := internal.ConvertErrorToConnectError(err)
		headers = internal.ConvertToProtoHeader(connectErr.Meta())
		protoErr = internal.ConvertConnectToProtoError(connectErr)
	} else {
		// If the call was successful, get the returned payloads
		// and the headers and trailers
		payloads = append(payloads, resp.Msg.Payload)
		headers = internal.ConvertToProtoHeader(resp.Header())
		trailers = internal.ConvertToProtoHeader(resp.Trailer())
	}

	return &v1alpha1.ClientResponseResult{
		ResponseHeaders:  headers,
		ResponseTrailers: trailers,
		Payloads:         payloads,
		Error:            protoErr,
		ConnectErrorRaw:  nil, // TODO
	}, nil
}

func (i *invoker) serverStream(
	ctx context.Context,
	req *v1alpha1.ClientCompatRequest,
) (*v1alpha1.ClientResponseResult, error) {
	msg := req.RequestMessages[0]
	ssr := &v1alpha1.ServerStreamRequest{}
	if err := msg.UnmarshalTo(ssr); err != nil {
		return nil, err
	}

	request := connect.NewRequest(ssr)

	// Add the specified request headers to the request
	internal.AddHeaders(req.RequestHeaders, request.Header())

	stream, err := i.client.ServerStream(ctx, request)
	if err != nil {
		return nil, err
	}
	var protoErr *v1alpha1.Error
	var headers []*v1alpha1.Header
	var trailers []*v1alpha1.Header
	payloads := make([]*v1alpha1.ConformancePayload, 0, len(ssr.ResponseDefinition.ResponseData))
	for stream.Receive() {
		// If the call was successful, get the returned payloads
		// and the headers and trailers
		payloads = append(payloads, stream.Msg().Payload)
	}
	if stream.Err() != nil {
		// If an error was returned, convert it to a proto Error
		protoErr = internal.ConvertErrorToProtoError(stream.Err())
	}

	// Read headers and trailers from the stream
	headers = internal.ConvertToProtoHeader(stream.ResponseHeader())
	trailers = internal.ConvertToProtoHeader(stream.ResponseTrailer())

	err = stream.Close()
	if err != nil {
		return nil, err
	}
	return &v1alpha1.ClientResponseResult{
		ResponseHeaders:  headers,
		ResponseTrailers: trailers,
		Payloads:         payloads,
		Error:            protoErr,
		ConnectErrorRaw:  nil, // TODO
	}, nil
}

func (i *invoker) clientStream(
	ctx context.Context,
	req *v1alpha1.ClientCompatRequest,
) (*v1alpha1.ClientResponseResult, error) {
	stream := i.client.ClientStream(ctx)

	// Add the specified request headers to the request
	internal.AddHeaders(req.RequestHeaders, stream.RequestHeader())

	for _, msg := range req.RequestMessages {
		csr := &v1alpha1.ClientStreamRequest{}
		if err := msg.UnmarshalTo(csr); err != nil {
			return nil, err
		}

		// Sleep for any specified delay
		time.Sleep(time.Duration(req.RequestDelayMs) * time.Millisecond)

		if err := stream.Send(csr); err != nil && errors.Is(err, io.EOF) {
			break
		}
	}

	var protoErr *v1alpha1.Error
	var headers []*v1alpha1.Header
	var trailers []*v1alpha1.Header
	payloads := make([]*v1alpha1.ConformancePayload, 0, 1)

	resp, err := stream.CloseAndReceive()
	if err != nil {
		// If an error was returned, first convert it to a Connect error
		// so that we can get the headers from the Meta property. Then,
		// convert _that_ to a proto Error so we can set it in the response.
		connectErr := internal.ConvertErrorToConnectError(err)
		headers = internal.ConvertToProtoHeader(connectErr.Meta())
		protoErr = internal.ConvertConnectToProtoError(connectErr)
	} else {
		// If the call was successful, get the returned payloads
		// and the headers and trailers
		payloads = append(payloads, resp.Msg.Payload)
		headers = internal.ConvertToProtoHeader(resp.Header())
		trailers = internal.ConvertToProtoHeader(resp.Trailer())
	}

	return &v1alpha1.ClientResponseResult{
		ResponseHeaders:  headers,
		ResponseTrailers: trailers,
		Payloads:         payloads,
		Error:            protoErr,
		ConnectErrorRaw:  nil, // TODO
	}, nil
}

func (i *invoker) bidiStream(
	ctx context.Context,
	req *v1alpha1.ClientCompatRequest,
) (result *v1alpha1.ClientResponseResult, retErr error) {
	result = &v1alpha1.ClientResponseResult{
		ConnectErrorRaw: nil, // TODO
	}

	stream := i.client.BidiStream(ctx)
	defer func() {
		if result != nil {
			// Read headers and trailers from the stream
			result.ResponseHeaders = internal.ConvertToProtoHeader(stream.ResponseHeader())
			result.ResponseTrailers = internal.ConvertToProtoHeader(stream.ResponseTrailer())
		}
	}()

	// Add the specified request headers to the request
	internal.AddHeaders(req.RequestHeaders, stream.RequestHeader())

	fullDuplex := req.StreamType == v1alpha1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM

	var protoErr *v1alpha1.Error
	for _, msg := range req.RequestMessages {
		if err := ctx.Err(); err != nil {
			// If an error was returned, convert it to a proto Error
			protoErr = internal.ConvertErrorToProtoError(err)
			break
		}
		bsr := &v1alpha1.BidiStreamRequest{}
		if err := msg.UnmarshalTo(bsr); err != nil {
			// Return the error and nil result because this is an
			// unmarshalling error unrelated to the RPC
			return nil, err
		}
		if err := stream.Send(bsr); err != nil && errors.Is(err, io.EOF) {
			// Call receive to get the error and convert it to a proto error
			if _, recvErr := stream.Receive(); recvErr != nil {
				protoErr = internal.ConvertErrorToProtoError(recvErr)
			} else {
				// Just in case the receive call doesn't return the error,
				// use the error returned from Send. Note this should never
				// happen, but is here as a safeguard.
				protoErr = internal.ConvertErrorToProtoError(err)
			}
			// Break the send loop
			break
		}
		if fullDuplex {
			// If this is a full duplex stream, receive a response for each request
			msg, err := stream.Receive()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					// If an error was returned that is not an EOF, convert it
					// to a proto Error. If the error was an EOF, that just means
					// reads are done.
					protoErr = internal.ConvertErrorToProtoError(err)
				}
				// Reads are done either because we received an error or an EOF
				// In either case, break the outer loop
				break
			}
			// If the call was successful, get the returned payloads
			result.Payloads = append(result.Payloads, msg.Payload)
		}
	}

	// If we received an error in any of the send logic or full-duplex reads, then exit
	if protoErr != nil {
		result.Error = protoErr
		return result, nil
	}

	// Sends are done, close the send side of the stream
	if err := stream.CloseRequest(); err != nil {
		result.Error = internal.ConvertErrorToProtoError(err)
		return result, nil
	}

	// Receive any remaining responses
	for {
		if err := ctx.Err(); err != nil {
			// If an error was returned, convert it to a proto Error
			protoErr = internal.ConvertErrorToProtoError(err)
			break
		}
		msg, err := stream.Receive()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				// If an error was returned that is not an EOF, convert it
				// to a proto Error. If the error was an EOF, that just means
				// reads are done.
				protoErr = internal.ConvertErrorToProtoError(err)
			}
			break
		}
		// If the call was successful, save the payloads
		result.Payloads = append(result.Payloads, msg.Payload)
	}

	if protoErr != nil {
		result.Error = protoErr
		return result, nil
	}

	if err := stream.CloseResponse(); err != nil {
		result.Error = internal.ConvertErrorToProtoError(err)
	}

	return result, nil
}

// Creates a new invoker around a ConformanceServiceClient.
func newInvoker(transport http.RoundTripper, url *url.URL, opts []connect.ClientOption) *invoker {
	client := conformancev1alpha1connect.NewConformanceServiceClient(
		&http.Client{Transport: transport},
		url.String(),
		opts...,
	)
	return &invoker{
		client: client,
	}
}
