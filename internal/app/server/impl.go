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

package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"connectrpc.com/conformance/internal/app"
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

type conformanceServer struct{}

func (s *conformanceServer) Unary(
	ctx context.Context,
	req *connect.Request[v1alpha1.UnaryRequest],
) (*connect.Response[v1alpha1.UnaryResponse], error) {
	msgAsAny, err := asAny(req.Msg)
	if err != nil {
		return nil, err
	}
	payload, connectErr := parseUnaryResponseDefinition(
		req.Msg.ResponseDefinition,
		req.Header(),
		[]*anypb.Any{msgAsAny},
	)
	if connectErr != nil {
		app.AddHeaders(req.Msg.ResponseDefinition.ResponseHeaders, connectErr.Meta())
		app.AddHeaders(req.Msg.ResponseDefinition.ResponseTrailers, connectErr.Meta())
		return nil, connectErr
	}

	resp := connect.NewResponse(&v1alpha1.UnaryResponse{
		Payload: payload,
	})

	app.AddHeaders(req.Msg.ResponseDefinition.ResponseHeaders, resp.Header())
	app.AddHeaders(req.Msg.ResponseDefinition.ResponseTrailers, resp.Trailer())

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
	if err != nil {
		app.AddHeaders(responseDefinition.ResponseHeaders, err.Meta())
		app.AddHeaders(responseDefinition.ResponseTrailers, err.Meta())
		return nil, err
	}

	resp := connect.NewResponse(&v1alpha1.ClientStreamResponse{
		Payload: payload,
	})

	app.AddHeaders(responseDefinition.ResponseHeaders, resp.Header())
	app.AddHeaders(responseDefinition.ResponseTrailers, resp.Trailer())

	return resp, err
}

func (s *conformanceServer) ServerStream(
	ctx context.Context,
	req *connect.Request[v1alpha1.ServerStreamRequest],
	stream *connect.ServerStream[v1alpha1.ServerStreamResponse],
) error {
	responseDefinition := req.Msg.ResponseDefinition
	if responseDefinition != nil {
		app.AddHeaders(responseDefinition.ResponseHeaders, stream.ResponseHeader())
		app.AddHeaders(responseDefinition.ResponseTrailers, stream.ResponseTrailer())
	}

	// Convert the request to an Any so that it can be recorded in the payload
	msgAsAny, err := asAny(req.Msg)
	if err != nil {
		return err
	}
	requestInfo := createRequestInfo(req.Header(), []*anypb.Any{msgAsAny})
	payload := &v1alpha1.ConformancePayload{
		RequestInfo: requestInfo,
	}

	for _, data := range responseDefinition.ResponseData {
		payload.Data = data

		resp := &v1alpha1.ServerStreamResponse{
			Payload: payload,
		}

		time.Sleep((time.Duration(responseDefinition.ResponseDelayMs) * time.Millisecond))

		if err := stream.Send(resp); err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("error sending on stream: %w", err))
		}
		// Only echo back the request info in the first response
		payload.RequestInfo = nil
	}
	if responseDefinition.Error != nil {
		return app.ConvertToConnectError(responseDefinition.Error)
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
			requestInfo := createRequestInfo(stream.RequestHeader(), reqs)
			resp := &v1alpha1.BidiStreamResponse{
				Payload: &v1alpha1.ConformancePayload{
					RequestInfo: requestInfo,
					Data:        responseDefinition.ResponseData[respNum],
				},
			}
			time.Sleep((time.Duration(responseDefinition.ResponseDelayMs) * time.Millisecond))

			if err := stream.Send(resp); err != nil {
				return connect.NewError(connect.CodeInternal, fmt.Errorf("error sending on stream: %w", err))
			}
			respNum++
			reqs = nil
		}
	}

	// If we still have responses left to send, flush them now. This accommodates
	// both scenarios of half duplex (we haven't sent any responses yet) or full duplex
	// where the requested responses are greater than the total requests.
	for ; respNum < len(responseDefinition.ResponseData); respNum++ {
		requestInfo := createRequestInfo(stream.RequestHeader(), reqs)
		resp := &v1alpha1.BidiStreamResponse{
			Payload: &v1alpha1.ConformancePayload{
				RequestInfo: requestInfo,
				Data:        responseDefinition.ResponseData[respNum],
			},
		}
		time.Sleep((time.Duration(responseDefinition.ResponseDelayMs) * time.Millisecond))

		if err := stream.Send(resp); err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("error sending on stream: %w", err))
		}
	}

	if responseDefinition.Error != nil {
		return app.ConvertToConnectError(responseDefinition.Error)
	}
	return nil
}

// Parses the given unary response definition and returns either
// a built payload or a connect error based on the definition.
func parseUnaryResponseDefinition(
	def *v1alpha1.UnaryResponseDefinition,
	headers http.Header,
	reqs []*anypb.Any,
) (*v1alpha1.ConformancePayload, *connect.Error) {
	if def != nil {
		switch rt := def.Response.(type) {
		case *v1alpha1.UnaryResponseDefinition_Error:
			return nil, app.ConvertToConnectError(rt.Error)
		case *v1alpha1.UnaryResponseDefinition_ResponseData, nil:
			requestInfo := createRequestInfo(headers, reqs)
			payload := &v1alpha1.ConformancePayload{
				RequestInfo: requestInfo,
			}

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

// Creates request info for a conformance payload
func createRequestInfo(headers http.Header, reqs []*anypb.Any) *v1alpha1.ConformancePayload_RequestInfo {
	headerInfo := app.ConvertToProtoHeader(headers)

	// Set all observed request headers and requests in the response payload
	return &v1alpha1.ConformancePayload_RequestInfo{
		RequestHeaders: headerInfo,
		Requests:       reqs,
	}
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
