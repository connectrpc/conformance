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

	"connectrpc.com/conformance/internal/gen/proto/connect/connectrpc/conformance/v1alpha1/conformancev1alpha1connect"
	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	connect "connectrpc.com/connect"
)

// NewConformanceServiceHandler returns a new ConformanceServiceHandler.
func NewConformanceServiceHandler() conformancev1alpha1connect.ConformanceServiceHandler {
	return &conformanceServer{}
}

type conformanceServer struct{}

func (s *conformanceServer) Unary(ctx context.Context, req *connect.Request[v1alpha1.UnaryRequest]) (*connect.Response[v1alpha1.ConformancePayload], error) {
	res := connect.NewResponse(&v1alpha1.ConformancePayload{})

	// Set all requested response headers on the response
	for _, header := range req.Msg.ResponseHeaders {
		for _, val := range header.Value {
			res.Header().Add(header.Name, val)
		}
	}

	// Set all requested response trailers on the response
	for _, trailer := range req.Msg.ResponseTrailers {
		for _, val := range trailer.Value {
			res.Trailer().Add(trailer.Name, val)
		}
	}

	// Set all observed request headers in the response payload
	headerInfo := []*v1alpha1.Header{}
	for key, value := range req.Header() {
		hdr := &v1alpha1.Header{
			Name:  key,
			Value: value,
		}
		headerInfo = append(headerInfo, hdr)
	}
	res.Msg.RequestInfo = &v1alpha1.ConformancePayload_RequestInfo{
		RequestHeaders: headerInfo,
	}

	switch rt := req.Msg.Response.(type) {
	case *v1alpha1.UnaryRequest_ResponseData:
		res.Msg.Data = rt.ResponseData
	case *v1alpha1.UnaryRequest_Error:
		connectErr := connect.NewError(connect.Code(rt.Error.Code), errors.New(rt.Error.Message))
		for _, reqDetail := range rt.Error.Details {
			connectDetail, err := connect.NewErrorDetail(reqDetail)
			if err != nil {
				return nil, err
			}
			connectErr.AddDetail(connectDetail)
		}

		return nil, connectErr
	case nil:
		// TODO - Which error should we raise here? Invalid Argument?
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("A desired response data or response error is required"))
	default:
		return nil, fmt.Errorf("UnaryRequest.Response has an unexpected type %T", rt)
	}

	return res, nil
}

func (s *conformanceServer) ClientStream(context.Context, *connect.ClientStream[v1alpha1.ClientStreamRequest]) (*connect.Response[v1alpha1.ConformancePayload], error) {
	return nil, connect.NewError(
		connect.CodeUnimplemented,
		errors.New("ClientStream is not yet implemented"),
	)
}

func (s *conformanceServer) ServerStream(context.Context, *connect.Request[v1alpha1.ServerStreamRequest], *connect.ServerStream[v1alpha1.ConformancePayload]) error {
	return connect.NewError(
		connect.CodeUnimplemented,
		errors.New("ServerStream is not yet implemented"),
	)
}

func (s *conformanceServer) BidiStream(context.Context, *connect.BidiStream[v1alpha1.BidiStreamRequest, v1alpha1.ConformancePayload]) error {
	return connect.NewError(
		connect.CodeUnimplemented,
		errors.New("BidiStream is not yet implemented"),
	)
}
