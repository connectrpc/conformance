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

package referenceserver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/conformance/internal"
	v1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1/conformancev1connect"
	connect "connectrpc.com/connect"
	proto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const serverName = "connectconformance-referenceserver"

// ConformanceRequest is a general interface for all conformance requests (UnaryRequest, ServerStreamRequest, etc.)
type ConformanceRequest interface {
	GetResponseHeaders() []*v1.Header
	GetResponseTrailers() []*v1.Header
}

type conformanceServer struct {
	conformancev1connect.UnimplementedConformanceServiceHandler
}

func (s *conformanceServer) Unary(
	ctx context.Context,
	req *connect.Request[v1.UnaryRequest],
) (*connect.Response[v1.UnaryResponse], error) {
	fmt.Printf("Request: %+v\n", req)
	msgAsAny, err := asAny(req.Msg)
	if err != nil {
		return nil, err
	}
	payload, connectErr := parseUnaryResponseDefinition(
		ctx,
		req.Msg.ResponseDefinition,
		req.Header(),
		req.Peer().Query,
		[]*anypb.Any{msgAsAny},
	)
	if connectErr != nil {
		return nil, connectErr
	}

	resp := connect.NewResponse(&v1.UnaryResponse{
		Payload: payload,
	})

	if req.Msg.ResponseDefinition != nil {
		internal.AddHeaders(req.Msg.ResponseDefinition.ResponseHeaders, resp.Header())
		internal.AddHeaders(req.Msg.ResponseDefinition.ResponseTrailers, resp.Trailer())

		// If a response delay was specified, sleep for that amount of ms before responding
		responseDelay := time.Duration(req.Msg.ResponseDefinition.ResponseDelayMs) * time.Millisecond
		time.Sleep(responseDelay)
	}

	fmt.Printf("%+v", resp)

	return resp, nil
}

// TODO - This should be consolidated with the unary implementation since they are
// mostly the same. See https://github.com/connectrpc/conformance/pull/721/files#r1415699842
// for an example.
func (s *conformanceServer) IdempotentUnary(
	ctx context.Context,
	req *connect.Request[v1.IdempotentUnaryRequest],
) (*connect.Response[v1.IdempotentUnaryResponse], error) {
	msgAsAny, err := asAny(req.Msg)
	if err != nil {
		return nil, err
	}
	payload, connectErr := parseUnaryResponseDefinition(
		ctx,
		req.Msg.ResponseDefinition,
		req.Header(),
		req.Peer().Query,
		[]*anypb.Any{msgAsAny},
	)
	if connectErr != nil {
		return nil, connectErr
	}

	resp := connect.NewResponse(&v1.IdempotentUnaryResponse{
		Payload: payload,
	})

	if req.Msg.ResponseDefinition != nil {
		internal.AddHeaders(req.Msg.ResponseDefinition.ResponseHeaders, resp.Header())
		internal.AddHeaders(req.Msg.ResponseDefinition.ResponseTrailers, resp.Trailer())

		time.Sleep((time.Duration(req.Msg.ResponseDefinition.ResponseDelayMs) * time.Millisecond))
	}

	return resp, nil
}

func (s *conformanceServer) ClientStream(
	ctx context.Context,
	stream *connect.ClientStream[v1.ClientStreamRequest],
) (*connect.Response[v1.ClientStreamResponse], error) {
	var responseDefinition *v1.UnaryResponseDefinition
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

	payload, err := parseUnaryResponseDefinition(
		ctx,
		responseDefinition,
		stream.RequestHeader(),
		stream.Peer().Query,
		reqs,
	)
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(&v1.ClientStreamResponse{
		Payload: payload,
	})

	if responseDefinition != nil {
		internal.AddHeaders(responseDefinition.ResponseHeaders, resp.Header())
		internal.AddHeaders(responseDefinition.ResponseTrailers, resp.Trailer())

		// If a response delay was specified, sleep for that amount of ms before responding
		responseDelay := time.Duration(responseDefinition.ResponseDelayMs) * time.Millisecond
		time.Sleep(responseDelay)
	}

	return resp, nil
}

func (s *conformanceServer) ServerStream(
	ctx context.Context,
	req *connect.Request[v1.ServerStreamRequest],
	stream *connect.ServerStream[v1.ServerStreamResponse],
) error {
	// Convert the request to an Any so that it can be recorded in the payload
	msgAsAny, err := asAny(req.Msg)
	if err != nil {
		return err
	}

	respNum := 0

	responseDefinition := req.Msg.ResponseDefinition
	if responseDefinition != nil { //nolint:nestif
		internal.AddHeaders(responseDefinition.ResponseHeaders, stream.ResponseHeader())
		internal.AddHeaders(responseDefinition.ResponseTrailers, stream.ResponseTrailer())
		// Immediately send the headers/trailers on the stream so that they can be read by the client
		if err := stream.Send(nil); err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("error sending on stream: %w", err))
		}

		// Calculate the response delay if specified
		responseDelay := time.Duration(responseDefinition.ResponseDelayMs) * time.Millisecond

		for _, data := range responseDefinition.ResponseData {
			resp := &v1.ServerStreamResponse{
				Payload: &v1.ConformancePayload{
					Data: data,
				},
			}

			// Only set the request info if this is the first response being sent back
			// because for server streams, nothing in the request info will change
			// after the first response.
			if respNum == 0 {
				resp.Payload.RequestInfo = createRequestInfo(ctx, req.Header(), req.Peer().Query, []*anypb.Any{msgAsAny})
			}

			fmt.Println("snoozin")

			// If a response delay was specified, sleep for that amount of ms before responding
			time.Sleep(responseDelay)

			if err := stream.Send(resp); err != nil {
				return connect.NewError(connect.CodeInternal, fmt.Errorf("error sending on stream: %w", err))
			}
			respNum++
		}

		if responseDefinition.Error != nil {
			if respNum == 0 {
				// We've sent no responses and are returning an error, so build a
				// RequestInfo message and append to the error details
				reqInfo := createRequestInfo(ctx, req.Header(), req.Peer().Query, []*anypb.Any{msgAsAny})
				reqInfoAny, err := anypb.New(reqInfo)
				if err != nil {
					return connect.NewError(connect.CodeInternal, err)
				}
				responseDefinition.Error.Details = append(responseDefinition.Error.Details, reqInfoAny)
			}
			return internal.ConvertProtoToConnectError(responseDefinition.Error)
		}
	}

	return nil
}

func (s *conformanceServer) BidiStream(
	ctx context.Context,
	stream *connect.BidiStream[v1.BidiStreamRequest, v1.BidiStreamResponse],
) error {
	var responseDefinition *v1.StreamResponseDefinition
	var responseDelay time.Duration
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

			// If a response definition was provided, add the headers and trailers
			if responseDefinition != nil {
				internal.AddHeaders(responseDefinition.ResponseHeaders, stream.ResponseHeader())
				internal.AddHeaders(responseDefinition.ResponseTrailers, stream.ResponseTrailer())
				// Immediately send the headers on the stream so that they can be read by the client
				if err := stream.Send(nil); err != nil {
					return connect.NewError(connect.CodeInternal, fmt.Errorf("error sending on stream: %w", err))
				}

				// Calculate a response delay if specified
				responseDelay = time.Duration(responseDefinition.ResponseDelayMs) * time.Millisecond
			}
		}

		// If fullDuplex, then send one of the desired responses each time we get a message on the stream
		if fullDuplex {
			if responseDefinition == nil || respNum >= len(responseDefinition.ResponseData) {
				// If there are no responses to send, then break the receive loop
				// and throw the error specified
				break
			}

			resp := &v1.BidiStreamResponse{
				Payload: &v1.ConformancePayload{
					Data: responseDefinition.ResponseData[respNum],
				},
			}
			var requestInfo *v1.ConformancePayload_RequestInfo
			if respNum == 0 {
				// Only send the full request info (including headers and timeouts)
				// in the first response
				requestInfo = createRequestInfo(ctx, stream.RequestHeader(), stream.Peer().Query, reqs)
			} else {
				// All responses after the first should only include the requests
				// since that is the only thing that will change between responses
				// for a full duplex stream
				requestInfo = &v1.ConformancePayload_RequestInfo{
					Requests: reqs,
				}
			}
			resp.Payload.RequestInfo = requestInfo

			// If a response delay was specified, sleep for that amount of ms before responding
			time.Sleep(responseDelay)

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
	if responseDefinition != nil { //nolint:nestif
		for ; respNum < len(responseDefinition.ResponseData); respNum++ {
			if err := ctx.Err(); err != nil {
				return err
			}
			resp := &v1.BidiStreamResponse{
				Payload: &v1.ConformancePayload{
					Data: responseDefinition.ResponseData[respNum],
				},
			}
			// Only set the request info if this is the first response being sent back
			// because for half duplex streams, nothing in the request info will change
			// after the first response (this includes the requests since they've all
			// been received by this point)
			if respNum == 0 {
				resp.Payload.RequestInfo = createRequestInfo(
					ctx, stream.RequestHeader(),
					stream.Peer().Query,
					reqs,
				)
			}

			// If a response delay was specified, sleep for that amount of ms before responding
			time.Sleep(responseDelay)

			if err := stream.Send(resp); err != nil {
				return connect.NewError(connect.CodeInternal, fmt.Errorf("error sending on stream: %w", err))
			}
		}

		if responseDefinition.Error != nil {
			if respNum == 0 {
				// We've sent no responses and are returning an error, so build a
				// RequestInfo message and append to the error details
				reqInfo := createRequestInfo(ctx, stream.RequestHeader(), stream.Peer().Query, reqs)
				reqInfoAny, err := anypb.New(reqInfo)
				if err != nil {
					return connect.NewError(connect.CodeInternal, err)
				}
				responseDefinition.Error.Details = append(responseDefinition.Error.Details, reqInfoAny)
			}
			return internal.ConvertProtoToConnectError(responseDefinition.Error)
		}
	}

	return nil
}

// Parses the given unary response definition and returns either
// a built payload or a connect error based on the definition.
func parseUnaryResponseDefinition(
	ctx context.Context,
	def *v1.UnaryResponseDefinition,
	hdrs http.Header,
	queryParams url.Values,
	reqs []*anypb.Any,
) (*v1.ConformancePayload, *connect.Error) {
	reqInfo := createRequestInfo(ctx, hdrs, queryParams, reqs)
	if def == nil {
		// If the definition is not set at all, there's nothing to respond with.
		// Just return a payload with the request info
		return &v1.ConformancePayload{
			RequestInfo: reqInfo,
		}, nil
	}

	switch respType := def.Response.(type) {
	case *v1.UnaryResponseDefinition_Error:
		// The server should add the request info to the error details
		// for unary responses that return an error.
		reqInfoAny, err := anypb.New(reqInfo)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		respType.Error.Details = append(respType.Error.Details, reqInfoAny)

		connectErr := internal.ConvertProtoToConnectError(respType.Error)

		internal.AddHeaders(def.GetResponseHeaders(), connectErr.Meta())
		internal.AddHeaders(def.GetResponseTrailers(), connectErr.Meta())

		return nil, connectErr

	case *v1.UnaryResponseDefinition_ResponseData, nil:
		payload := &v1.ConformancePayload{
			RequestInfo: reqInfo,
		}

		// If response data was provided, set that in the payload response
		if respType, ok := respType.(*v1.UnaryResponseDefinition_ResponseData); ok {
			payload.Data = respType.ResponseData
		}
		return payload, nil
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provided UnaryRequest.Response has an unexpected type %T", respType))
	}
}

// Creates request info for a conformance payload.
func createRequestInfo(
	ctx context.Context,
	headers http.Header,
	queryParams url.Values,
	reqs []*anypb.Any,
) *v1.ConformancePayload_RequestInfo {
	headerInfo := internal.ConvertToProtoHeader(headers)

	var connectGetInfo *v1.ConformancePayload_ConnectGetInfo
	if len(queryParams) > 0 {
		queryParamInfo := internal.ConvertToProtoHeader(queryParams)

		connectGetInfo = &v1.ConformancePayload_ConnectGetInfo{
			QueryParams: queryParamInfo,
		}
	}

	var timeoutMs *int64
	if deadline, ok := ctx.Deadline(); ok {
		timeoutMs = proto.Int64(time.Until(deadline).Milliseconds())
	}

	// Set all observed request headers and requests in the response payload
	return &v1.ConformancePayload_RequestInfo{
		RequestHeaders: headerInfo,
		Requests:       reqs,
		TimeoutMs:      timeoutMs,
		ConnectGetInfo: connectGetInfo,
	}
}

// Converts the given message to an Any.
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

// serverNameHandlerInterceptor adds a "server" header on outgoing responses.
type serverNameHandlerInterceptor struct{}

func (i serverNameHandlerInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		resp, err := next(ctx, req)
		if req.Spec().IsClient {
			return resp, err
		}

		var headers http.Header
		if err != nil {
			var connErr *connect.Error
			if !errors.As(err, &connErr) {
				connErr = connect.NewError(connect.CodeUnknown, err)
				err = connErr
			}
			headers = connErr.Meta()
		} else {
			headers = resp.Header()
		}
		// decorate server with the program name and version
		server := strings.TrimSpace(fmt.Sprintf("%s %s/%s", headers.Get("Server"), serverName, internal.Version))
		headers.Set("Server", server)
		return resp, err
	}
}

func (i serverNameHandlerInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i serverNameHandlerInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, stream connect.StreamingHandlerConn) error {
		// decorate server with the program name and version
		server := strings.TrimSpace(fmt.Sprintf("%s %s/%s", stream.ResponseHeader().Get("Server"), serverName, internal.Version))
		stream.ResponseHeader().Set("Server", server)
		return next(ctx, stream)
	}
}
