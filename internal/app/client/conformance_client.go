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
	"net/http"
	"net/url"

	"connectrpc.com/conformance/internal/app"
	"connectrpc.com/conformance/internal/gen/proto/connect/connectrpc/conformance/v1alpha1/conformancev1alpha1connect"
	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"connectrpc.com/connect"
)

type Invoker struct {
	client conformancev1alpha1connect.ConformanceServiceClient
}

func (w *Invoker) Invoke(
	ctx context.Context,
	req *v1alpha1.ClientCompatRequest,
) (*v1alpha1.ClientResponseResult, error) {
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
		// TODO - Implement ClientStream
		return nil, errors.New("client streaming is not yet supported")
	case "BidiStream":
		// TODO - Implement BidiStream
		return nil, errors.New("bidi streaming is not yet supported")
	default:
		return nil, errors.New("method name " + req.Method + " does not exist")
	}
}

func (w *Invoker) unary(
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

	return &v1alpha1.ClientResponseResult{
		ResponseHeaders:  headers,
		ResponseTrailers: trailers,
		Payloads:         payloads,
		Error:            protoErr,
		ErrorDetailsRaw:  nil, // TODO
	}, nil
}

func (w *Invoker) serverStream(
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
	app.AddHeaders(req.RequestHeaders, request.Header())

	stream, err := w.client.ServerStream(ctx, request)
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

	err = stream.Close()
	if err != nil {
		return nil, err
	}
	return &v1alpha1.ClientResponseResult{
		ResponseHeaders:  headers,
		ResponseTrailers: trailers,
		Payloads:         payloads,
		Error:            protoErr,
		ErrorDetailsRaw:  nil, // TODO
	}, nil
}

// NewInvoker creates a new wrapper around a ConformanceServiceClient.
func NewInvoker(transport http.RoundTripper, url *url.URL, opts []connect.ClientOption) *Invoker {
	client := conformancev1alpha1connect.NewConformanceServiceClient(
		&http.Client{Transport: transport},
		url.String(),
		opts...,
	)
	return &Invoker{
		client: client,
	}
}
