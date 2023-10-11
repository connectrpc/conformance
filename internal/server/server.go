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
	"go.starlark.net/lib/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// ConformanceRequest is a general interface for all conformance requests (UnaryRequest, ServerStreamRequest, etc.)
type ConformanceRequest interface {
	GetResponseHeaders() []*v1alpha1.Header
	GetResponseTrailers() []*v1alpha1.Header
}

// NewConformanceServiceHandler returns a new ConformanceServiceHandler.
func NewConformanceServiceHandler() conformancev1alpha1connect.ConformanceServiceHandler {
	return &conformanceServer{}
}

type conformanceServer struct{}

func (s *conformanceServer) Unary(
	ctx context.Context,
	req *connect.Request[v1alpha1.UnaryRequest],
) (*connect.Response[v1alpha1.ConformancePayload], error) {

	msgAsAny, err := anypb.New(req.Msg)

	return buildUnaryResponse(req.Msg, req.Header())
}

func (s *conformanceServer) ClientStream(
	ctx context.Context,
	stream *connect.ClientStream[v1alpha1.ClientStreamRequest],
) (*connect.Response[v1alpha1.ConformancePayload], error) {
	var responseDefinition *v1alpha1.UnaryRequest
	firstRecv := true
	var requests []*anypb.Any
	for stream.Receive() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		msg := stream.Msg()
		// If this is the first message received on the stream, save off the total responses we need to send
		// plus whether this should be full or half duplex
		if firstRecv {
			if msg.ResponseDefinition == nil {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.New("a response definition must be provided in the first message on a client stream"),
				)
			}
			responseDefinition = msg.ResponseDefinition
			firstRecv = false
		}
		msgAsAny, err := anypb.New(msg)
		if err != nil {
			return nil, connect.NewError(
				connect.CodeInternal,
				fmt.Errorf("unable to convert request message: %w", err),
			)
		}
		requests = append(requests, msgAsAny)
	}
	if err := stream.Err(); err != nil {
		return nil, err
	}

	return buildUnaryResponse(responseDefinition, stream.RequestHeader(), requests)
}

func (s *conformanceServer) ServerStream(
	ctx context.Context,
	req *connect.Request[v1alpha1.ServerStreamRequest],
	stream *connect.ServerStream[v1alpha1.ConformancePayload],
) error {
	var ticker *time.Ticker
	if req.Msg.WaitBeforeEachMessageMillis > 0 {
		ticker = time.NewTicker(time.Duration(req.Msg.WaitBeforeEachMessageMillis) * time.Millisecond)
		defer ticker.Stop()
	}
	res := buildConformancePayload(req.Msg, req.Header())

	for _, data := range req.Msg.ResponseData {
		if ticker != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
			}
		}
		res.Msg.Data = data
		if err := stream.Send(res.Msg); err != nil {
			return err
		}
	}
	if req.Msg.Error != nil {
		return createError(req.Msg.Error)
	}
	return nil
}

func (s *conformanceServer) BidiStream(
	ctx context.Context,
	stream *connect.BidiStream[v1alpha1.BidiStreamRequest, v1alpha1.ConformancePayload],
) error {
	var responseDefinition *v1alpha1.ServerStreamRequest
	fullDuplex := false
	firstRecv := true
	respNum := 0
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

		// If this is the first message in the stream, save off the total responses we need to send
		// plus whether this should be full or half duplex
		if firstRecv {
			if req.ResponseDefinition == nil {
				return connect.NewError(
					connect.CodeInvalidArgument,
					errors.New("a response definition must be provided in the first message on a Bidi stream"),
				)
			}
			responseDefinition = req.ResponseDefinition
			fullDuplex = req.WaitForEachRequest
			firstRecv = false
		}

		// If fullDuplex, then send one of the desired responses each time we get a message on the stream
		if fullDuplex {
			res := buildConformancePayload(responseDefinition, stream.RequestHeader())
			res.Msg.Data = responseDefinition.ResponseData[respNum]
			err := sendBidi(ctx, stream, responseDefinition.WaitBeforeEachMessageMillis, res.Msg)
			if err != nil {
				return err
			}
			respNum++
		}
	}

	// If this is a half duplex call, then send all the responses now.
	// If this is a full deplex call, then flush any remaining responses. It is possible
	// that the initial request specifying the desired response definitions contained more
	// definitions than requests sent on the stream. In that case, if we interleave for
	// full duplex, we should have some responses left over to send.
	if respNum < len(responseDefinition.ResponseData) {
		for i := respNum; i < len(responseDefinition.ResponseData); i++ {
			res := buildConformancePayload(responseDefinition, stream.RequestHeader())
			res.Msg.Data = responseDefinition.ResponseData[i]
			err := sendBidi(ctx, stream, responseDefinition.WaitBeforeEachMessageMillis, res.Msg)
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

func sendBidi(
	ctx context.Context,
	stream *connect.BidiStream[v1alpha1.BidiStreamRequest, v1alpha1.ConformancePayload],
	delay uint32,
	resp *v1alpha1.ConformancePayload,
) error {
	var ticker *time.Ticker
	if delay > 0 {
		ticker = time.NewTicker(time.Duration(delay) * time.Millisecond)
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
		return fmt.Errorf("error sending on bidi stream: %w", err)
	}
	return nil
}

func buildConformancePayload(
	msg ConformanceRequest,
	headers http.Header,
	requests []*anypb.Any,
) *connect.Response[v1alpha1.ConformancePayload] {
	res := connect.NewResponse(&v1alpha1.ConformancePayload{})

	// Set all requested response headers on the response
	for _, header := range msg.GetResponseHeaders() {
		for _, val := range header.Value {
			res.Header().Add(header.Name, val)
		}
	}
	// Set all requested response trailers on the response
	for _, trailer := range msg.GetResponseTrailers() {
		for _, val := range trailer.Value {
			res.Trailer().Add(trailer.Name, val)
		}
	}

	// Set all observed request headers in the response payload
	headerInfo := []*v1alpha1.Header{}
	for key, value := range headers {
		hdr := &v1alpha1.Header{
			Name:  key,
			Value: value,
		}
		headerInfo = append(headerInfo, hdr)
	}

	res.Msg.RequestInfo = &v1alpha1.ConformancePayload_RequestInfo{
		RequestHeaders: headerInfo,
		Requests:       requests,
	}

	return res
}

func buildUnaryResponse(
	def *v1alpha1.UnaryRequest,
	headers http.Header,
	requests []*anypb.Any,
) (*connect.Response[v1alpha1.ConformancePayload], error) {
	res := buildConformancePayload(def, headers, requests)

	switch rt := def.Response.(type) {
	case *v1alpha1.UnaryRequest_ResponseData:
		res.Msg.Data = rt.ResponseData
		return res, nil
	case *v1alpha1.UnaryRequest_Error:
		return nil, createError(rt.Error)
	case nil:
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			errors.New("desired response data or response error is required"),
		)
	default:
		return nil, fmt.Errorf("provided UnaryRequest.Response has an unexpected type %T", rt)
	}
}

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

func asAny(msg *proto.Message) (*anypb.Any, error) {
	msgAsAny, err := anypb.New(msg)
	if err != nil {
		return nil, connect.NewError(
			connect.CodeInternal,
			fmt.Errorf("unable to convert request message: %w", err),
		)
	}
}
