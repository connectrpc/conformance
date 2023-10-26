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
	"errors"
	"io"
	"net/http"
	"net/url"

	"connectrpc.com/conformance/internal/app"
	"connectrpc.com/conformance/internal/gen/proto/connect/connectrpc/conformance/v1alpha1/conformancev1alpha1connect"
	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"connectrpc.com/connect"
)

type conformanceClientWrapper struct {
	client conformancev1alpha1connect.ConformanceServiceClient
}

func (w *conformanceClientWrapper) Invoke(
	ctx context.Context,
	req *v1alpha1.ClientCompatRequest,
) (*v1alpha1.ClientCompatResponse, error) {
	switch req.Method {
	case "Unary":
		if len(req.RequestMessages) != 1 {
			return nil, errors.New("unary calls must specify exactly one request message")
		}
		resp, err := w.unary(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	case "ServerStream":
		if len(req.RequestMessages) != 1 {
			return nil, errors.New("server streaming calls must specify exactly one request message")
		}
		resp, err := w.serverStream(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	case "ClientStream":
		if len(req.RequestMessages) < 1 {
			return nil, errors.New("client streaming calls must specify at least one request message")
		}
		resp, err := w.clientStream(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	case "BidiStream":
		if len(req.RequestMessages) < 1 {
			return nil, errors.New("bidi streaming calls must specify at least one request message")
		}
		resp, err := w.bidiStream(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	default:
		// TODO: Should this be a returned 'error' or via the ClientCompatResponse?
		// We should probably treat 'error' here as something independent of invoking a request
		// So any internal error that doesn't involve actually calling an RPC. In that case, then,
		// this should just return an errors.New like the above cases
		return &v1alpha1.ClientCompatResponse{
			TestName: req.TestName,
			Result: &v1alpha1.ClientCompatResponse_Error{
				Error: &v1alpha1.ClientErrorResult{
					Message: "method name " + req.Method + " does not exist",
				},
			},
		}, nil
	}
}

func (w *conformanceClientWrapper) unary(
	ctx context.Context,
	req *v1alpha1.ClientCompatRequest,
) (*v1alpha1.ClientCompatResponse, error) {
	ccResp := &v1alpha1.ClientCompatResponse{
		TestName: req.TestName,
	}
	msg := req.RequestMessages[0]
	ur := &v1alpha1.UnaryRequest{}
	if err := msg.UnmarshalTo(ur); err != nil {
		return nil, err
	}

	request := connect.NewRequest(ur)

	// Add the specified request headers to the request
	app.AddHeaders(req.RequestHeaders, request.Header())

	var protoErr *v1alpha1.Error
	var headers []*v1alpha1.Header
	var trailers []*v1alpha1.Header
	payloads := make([]*v1alpha1.ConformancePayload, 0, 1)

	// Invoke the Unary call
	resp, err := w.client.Unary(ctx, request)
	if err != nil {
		// If an error was returned, first convert it to a Connect error
		// so that we can get the headers from the Meta property. Then,
		// convert _that_ to a proto Error so we can set it in the response.
		connectErr := app.ConvertErrorToConnectError(err)
		headers = app.ConvertToProtoHeader(connectErr.Meta())
		protoErr = app.ConvertConnectToProtoError(connectErr)
	} else {
		// If the call was successful, get the returned payloads
		// and the headers and trailers
		payloads = append(payloads, resp.Msg.Payload)
		headers = app.ConvertToProtoHeader(resp.Header())
		trailers = app.ConvertToProtoHeader(resp.Trailer())
	}

	ccResp.Result = &v1alpha1.ClientCompatResponse_Response{
		Response: &v1alpha1.ClientResponseResult{
			ResponseHeaders:  headers,
			ResponseTrailers: trailers,
			Payloads:         payloads,
			Error:            protoErr,
			ErrorDetailsRaw:  nil, // TODO
		},
	}

	return ccResp, nil
}

func (w *conformanceClientWrapper) serverStream(
	ctx context.Context,
	req *v1alpha1.ClientCompatRequest,
) (*v1alpha1.ClientCompatResponse, error) {
	ccResp := &v1alpha1.ClientCompatResponse{
		TestName: req.TestName,
	}
	msg := req.RequestMessages[0]
	ssr := &v1alpha1.ServerStreamRequest{}
	if err := msg.UnmarshalTo(ssr); err != nil {
		return nil, err
	}

	request := connect.NewRequest(ssr)

	// Add the specified request headers to the request
	app.AddHeaders(req.RequestHeaders, request.Header())

	stream, err := w.client.ServerStream(ctx, request)
	// TODO - should this error be added to the clientcompatresponse or returned here?
	// IMO, the error returned from this function represents an internal error independent
	// of anything invoking the service but it's unclear what this err is vs. a stream.Err()
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
		protoErr = app.ConvertErrorToProtoError(stream.Err())
	}

	// Read headers and trailers from the stream
	headers = app.ConvertToProtoHeader(stream.ResponseHeader())
	trailers = app.ConvertToProtoHeader(stream.ResponseTrailer())

	stream.Close()
	ccResp.Result = &v1alpha1.ClientCompatResponse_Response{
		Response: &v1alpha1.ClientResponseResult{
			ResponseHeaders:  headers,
			ResponseTrailers: trailers,
			Payloads:         payloads,
			Error:            protoErr,
			ErrorDetailsRaw:  nil, // TODO
		},
	}
	return ccResp, nil
}

func (w *conformanceClientWrapper) clientStream(
	ctx context.Context,
	req *v1alpha1.ClientCompatRequest,
) (*v1alpha1.ClientCompatResponse, error) {
	ccResp := &v1alpha1.ClientCompatResponse{
		TestName: req.TestName,
	}
	stream := w.client.ClientStream(ctx)

	// Add the specified request headers to the request
	app.AddHeaders(req.RequestHeaders, stream.RequestHeader())

	for _, msg := range req.RequestMessages {
		csr := &v1alpha1.ClientStreamRequest{}
		if err := msg.UnmarshalTo(csr); err != nil {
			return nil, err
		}
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
		connectErr := app.ConvertErrorToConnectError(err)
		headers = app.ConvertToProtoHeader(connectErr.Meta())
		protoErr = app.ConvertConnectToProtoError(connectErr)
	} else {
		// If the call was successful, get the returned payloads
		// and the headers and trailers
		payloads = append(payloads, resp.Msg.Payload)
		headers = app.ConvertToProtoHeader(resp.Header())
		trailers = app.ConvertToProtoHeader(resp.Trailer())
	}

	ccResp.Result = &v1alpha1.ClientCompatResponse_Response{
		Response: &v1alpha1.ClientResponseResult{
			ResponseHeaders:  headers,
			ResponseTrailers: trailers,
			Payloads:         payloads,
			Error:            protoErr,
			ErrorDetailsRaw:  nil, // TODO
		},
	}
	return ccResp, nil
}

func (w *conformanceClientWrapper) bidiStream(
	ctx context.Context,
	req *v1alpha1.ClientCompatRequest,
) (*v1alpha1.ClientCompatResponse, error) {
	ccResp := &v1alpha1.ClientCompatResponse{
		TestName: req.TestName,
	}
	stream := w.client.BidiStream(context.Background())
	// Add the specified request headers to the request
	app.AddHeaders(req.RequestHeaders, stream.RequestHeader())

	firstSend := true
	fullDuplex := false

	var protoErr *v1alpha1.Error
	var headers []*v1alpha1.Header
	var trailers []*v1alpha1.Header
	payloads := make([]*v1alpha1.ConformancePayload, 0, 1)

	for _, msg := range req.RequestMessages {
		bsr := &v1alpha1.BidiStreamRequest{}
		if err := msg.UnmarshalTo(bsr); err != nil {
			return nil, err
		}
		if firstSend {
			fullDuplex = bsr.FullDuplex
		}
		if err := stream.Send(bsr); err != nil && errors.Is(err, io.EOF) {
			break
		}
		if fullDuplex {
			msg, err := stream.Receive()
			// If this is a full duplex stream, receive a response for each request
			if err != nil {
				if errors.Is(err, io.EOF) {
					// Reads are done, break the receive loop
					break
				}
				// If an error was returned, convert it to a proto Error
				// TODO - If we error here, we should stop immediately right?
				protoErr = app.ConvertErrorToProtoError(err)
				break
			}
			// If the call was successful, get the returned payloads
			// and the headers and trailers
			payloads = append(payloads, msg.Payload)
		}
	}

	// Close the send side of the stream
	if err := stream.CloseRequest(); err != nil {
		// TODO - If we error here, we should stop immediately right?
		protoErr = app.ConvertErrorToProtoError(err)
	}

	for {
		if err := ctx.Err(); err != nil {
			// If an error was returned, convert it to a proto Error
			protoErr = app.ConvertErrorToProtoError(err)
			break
		}
		msg, err := stream.Receive()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// Reads are done, break the receive loop
				break
			}
			// If an error was returned, convert it to a proto Error
			protoErr = app.ConvertErrorToProtoError(err)
			break
		}
		// If the call was successful, get the returned payloads
		// and the headers and trailers
		payloads = append(payloads, msg.Payload)
	}

	if err := stream.CloseResponse(); err != nil {
		protoErr = app.ConvertErrorToProtoError(err)
	}

	// Read headers and trailers from the stream
	headers = app.ConvertToProtoHeader(stream.ResponseHeader())
	trailers = app.ConvertToProtoHeader(stream.ResponseTrailer())

	// TODO - How can we distinguish whether we properly processed full vs. half duplex?
	ccResp.Result = &v1alpha1.ClientCompatResponse_Response{
		Response: &v1alpha1.ClientResponseResult{
			ResponseHeaders:  headers,
			ResponseTrailers: trailers,
			Payloads:         payloads,
			Error:            protoErr,
			ErrorDetailsRaw:  nil, // TODO
		},
	}
	return ccResp, nil
}

// NewConformanceClientWrapper creates a new wrapper around a ConformanceServiceClient.
func NewConformanceClientWrapper(transport http.RoundTripper, url *url.URL, opts []connect.ClientOption) Wrapper {
	client := conformancev1alpha1connect.NewConformanceServiceClient(
		&http.Client{Transport: transport},
		url.String(),
		opts...,
	)
	return &conformanceClientWrapper{
		client: client,
	}
}
