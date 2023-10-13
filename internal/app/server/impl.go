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
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"connectrpc.com/conformance/internal/gen/proto/connect/connectrpc/conformance/v1alpha1/conformancev1alpha1connect"
	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	connect "connectrpc.com/connect"
	proto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// ConformanceRequest is a general interface for all conformance requests (UnaryRequest, ServerStreamRequest, etc.)
type ConformanceRequest interface {
	GetResponseHeaders() []*v1alpha1.Header
	GetResponseTrailers() []*v1alpha1.Header
}

// Stream represents a stream with the ability to set headers and trailers and send a message T
type Stream[T any] interface {
	ResponseHeader() http.Header
	ResponseTrailer() http.Header
	Send(*T) error
}

type conformanceServer struct{}

func (s *conformanceServer) Unary(
	ctx context.Context,
	req *connect.Request[v1alpha1.UnaryRequest],
) (*connect.Response[v1alpha1.UnaryResponse], error) {
	msgAsAny, err := asAny(req.Msg)
	if err != nil {
		return nil, err
	}
	payload, err := parseUnaryResponseDefinition(req.Msg.ResponseDefinition, req.Header(), []*anypb.Any{msgAsAny})
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(&v1alpha1.UnaryResponse{
		Payload: payload,
	})

	if req.Msg.ResponseDefinition != nil {
		setResponseHeadersAndTrailers(req.Msg.ResponseDefinition, resp)
	}

	return resp, nil
}

func (s *conformanceServer) ClientStream(
	ctx context.Context,
	stream *connect.ClientStream[v1alpha1.ClientStreamRequest],
) (*connect.Response[v1alpha1.ClientStreamResponse], error) {
	var responseDefinition *v1alpha1.UnaryResponseDefinition
	firstRecv := true
	var reqs []*anypb.Any
	for stream.Receive() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		msg := stream.Msg()
		// If this is the first message received on the stream, save off the response definition we need to send
		if firstRecv {
			responseDefinition = msg.ResponseDefinition
			firstRecv = false
		}
		// Record all the requests received
		msgAsAny, err := asAny(msg)
		if err != nil {
			return nil, err
		}
		reqs = append(reqs, msgAsAny)
	}
	if err := stream.Err(); err != nil {
		return nil, err
	}

	payload, err := parseUnaryResponseDefinition(responseDefinition, stream.RequestHeader(), reqs)

	resp := connect.NewResponse(&v1alpha1.ClientStreamResponse{
		Payload: payload,
	})

	if responseDefinition != nil {
		setResponseHeadersAndTrailers(responseDefinition, resp)
	}

	return resp, err
}

func (s *conformanceServer) ServerStream(
	ctx context.Context,
	req *connect.Request[v1alpha1.ServerStreamRequest],
	stream *connect.ServerStream[v1alpha1.ServerStreamResponse],
) error {
	responseDefinition := req.Msg.ResponseDefinition
	if responseDefinition != nil {
		// Set all requested response headers on the response
		for _, header := range responseDefinition.ResponseHeaders {
			for _, val := range header.Value {
				stream.ResponseHeader().Add(header.Name, val)
			}
		}
		// Set all requested response trailers on the response
		for _, trailer := range responseDefinition.ResponseTrailers {
			for _, val := range trailer.Value {
				stream.ResponseTrailer().Add(trailer.Name, val)
			}
		}
	}

	// Convert the request to an Any so that it can be recorded in the payload
	msgAsAny, err := asAny(req.Msg)
	if err != nil {
		return err
	}
	payload := initPayload(req.Header(), []*anypb.Any{msgAsAny})

	for _, data := range responseDefinition.ResponseData {
		payload.Data = data

		resp := &v1alpha1.ServerStreamResponse{
			Payload: payload,
		}

		if err := sendOnStream[v1alpha1.ServerStreamResponse](ctx, stream, responseDefinition, resp); err != nil {
			return err
		}

	}
	if responseDefinition.Error != nil {
		return createError(responseDefinition.Error)
	}
	return nil
}

func (s *conformanceServer) BidiStream(
	ctx context.Context,
	stream *connect.BidiStream[v1alpha1.BidiStreamRequest, v1alpha1.BidiStreamResponse],
) error {
	var responseDefinition *v1alpha1.StreamResponseDefinition
	fullDuplex := false
	firstRecv := true
	respNum := 0
	var reqs []*anypb.Any
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		req, err := stream.Receive()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// Reads are done, break the receive loop and send any remaining responses
				break
			}
			return fmt.Errorf("receive request: %w", err)
		}

		// Record all requests received
		msgAsAny, err := asAny(req)
		if err != nil {
			return err
		}
		reqs = append(reqs, msgAsAny)

		// If this is the first message in the stream, save off the total responses we need to send
		// plus whether this should be full or half duplex
		if firstRecv {
			responseDefinition = req.ResponseDefinition
			fullDuplex = req.FullDuplex
			firstRecv = false
		}

		// If fullDuplex, then send one of the desired responses each time we get a message on the stream
		if fullDuplex {
			if respNum >= len(responseDefinition.ResponseData) {
				return connect.NewError(
					connect.CodeAborted,
					errors.New("received more requests than desired responses on a full duplex stream"),
				)
			}
			payload := initPayload(stream.RequestHeader(), reqs)
			payload.Data = responseDefinition.ResponseData[respNum]
			resp := &v1alpha1.BidiStreamResponse{
				Payload: payload,
			}
			err := sendOnStream[v1alpha1.BidiStreamResponse](ctx, stream, responseDefinition, resp)
			if err != nil {
				return err
			}
			respNum++
			reqs = nil
		}
	}

	// If this is a half duplex call, then send all the responses now.
	// If this is a full deplex call, then flush any remaining responses. It is possible
	// that the initial request specifying the desired response definitions contained more
	// definitions than requests sent on the stream. In that case, if we interleave for
	// full duplex, we should have some responses left over to send.
	if respNum < len(responseDefinition.ResponseData) {
		for i := respNum; i < len(responseDefinition.ResponseData); i++ {
			payload := initPayload(stream.RequestHeader(), reqs)
			payload.Data = responseDefinition.ResponseData[i]
			resp := &v1alpha1.BidiStreamResponse{
				Payload: payload,
			}
			err := sendOnStream[v1alpha1.BidiStreamResponse](ctx, stream, responseDefinition, resp)
			if err != nil {
				return err
			}

		}
	}

	if responseDefinition.Error != nil {
		return createError(responseDefinition.Error)
	}
	return nil
}

// NewConformanceServiceHandler returns a new ConformanceServiceHandler.
func NewConformanceServiceHandler() conformancev1alpha1connect.ConformanceServiceHandler {
	return &conformanceServer{}
}

// Sends a response T on the given stream, setting response headers and trailers
// according to the provided response definition.
func sendOnStream[T any](
	ctx context.Context,
	stream Stream[T],
	def *v1alpha1.StreamResponseDefinition,
	resp *T,
) error {
	var ticker *time.Ticker
	if def.ResponseDelayMs > 0 {
		ticker = time.NewTicker(time.Duration(def.ResponseDelayMs) * time.Millisecond)
		defer ticker.Stop()
	}
	if ticker != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}

	if err := stream.Send(resp); err != nil {
		return fmt.Errorf("error sending on stream: %w", err)
	}
	return nil
}

// Parses the given unary response definition and returns either
// a built payload or a connect error based on the definition.
func parseUnaryResponseDefinition(
	def *v1alpha1.UnaryResponseDefinition,
	headers http.Header,
	reqs []*anypb.Any,
) (*v1alpha1.ConformancePayload, error) {
	if def != nil {
		switch rt := def.Response.(type) {
		case *v1alpha1.UnaryResponseDefinition_Error:
			return nil, createError(rt.Error)
		case *v1alpha1.UnaryResponseDefinition_ResponseData, nil:
			payload := initPayload(headers, reqs)

			// If response data was provided, set that in the payload response
			if rt, ok := rt.(*v1alpha1.UnaryResponseDefinition_ResponseData); ok {
				payload.Data = rt.ResponseData
			}
			return payload, nil
		default:
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provided UnaryRequest.Response has an unexpected type %T", rt))
		}
	}
	return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("no response definition provided"))
}

// Initializes a conformance payload
func initPayload(headers http.Header, reqs []*anypb.Any) *v1alpha1.ConformancePayload {
	payload := &v1alpha1.ConformancePayload{}

	headerInfo := []*v1alpha1.Header{}
	for key, value := range headers {
		hdr := &v1alpha1.Header{
			Name:  key,
			Value: value,
		}
		headerInfo = append(headerInfo, hdr)
	}

	// Set all observed request headers and requests in the response payload
	payload.RequestInfo = &v1alpha1.ConformancePayload_RequestInfo{
		RequestHeaders: headerInfo,
		Requests:       reqs,
	}
	return payload
}

// Sets all response headers and trailers onto the given response
func setResponseHeadersAndTrailers[T any](def ConformanceRequest, resp *connect.Response[T]) {
	// Set all requested response headers on the response
	for _, header := range def.GetResponseHeaders() {
		for _, val := range header.Value {
			resp.Header().Add(header.Name, val)
		}
	}
	// Set all requested response trailers on the response
	for _, trailer := range def.GetResponseTrailers() {
		for _, val := range trailer.Value {
			resp.Trailer().Add(trailer.Name, val)
		}
	}
}

// Creates a Connect error from the given Error message
func createError(err *v1alpha1.Error) *connect.Error {
	connectErr := connect.NewError(connect.Code(err.Code), errors.New(err.Message))
	for _, detail := range err.Details {
		connectDetail, err := connect.NewErrorDetail(detail)
		if err != nil {
			return connect.NewError(connect.CodeInvalidArgument, err)
		}
		connectErr.AddDetail(connectDetail)
	}
	return connectErr
}

// Converts the given message to an Any
func asAny(msg proto.Message) (*anypb.Any, error) {
	msgAsAny, err := anypb.New(msg)
	if err != nil {
		return nil, connect.NewError(
			connect.CodeInternal,
			fmt.Errorf("unable to convert message: %w", err),
		)
	}
	return msgAsAny, nil
}
