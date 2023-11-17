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
	"time"

	"connectrpc.com/conformance/internal"
	"connectrpc.com/conformance/internal/gen/proto/connect/connectrpc/conformance/v2/conformancev2connect"
	v2 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v2"
	connect "connectrpc.com/connect"
	proto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// ConformanceRequest is a general interface for all conformance requests (UnaryRequest, ServerStreamRequest, etc.)
type ConformanceRequest interface {
	GetResponseHeaders() []*v2.Header
	GetResponseTrailers() []*v2.Header
}

type conformanceServer struct {
	conformancev2connect.UnimplementedConformanceServiceHandler
}

func (s *conformanceServer) Unary(
	ctx context.Context,
	req *connect.Request[v2.UnaryRequest],
) (*connect.Response[v2.UnaryResponse], error) {
	msgAsAny, err := asAny(req.Msg)
	if err != nil {
		return nil, err
	}
	payload, connectErr := parseUnaryResponseDefinition(
		ctx,
		req.Msg.ResponseDefinition,
		req.Header(),
		[]*anypb.Any{msgAsAny},
	)
	if connectErr != nil {
		internal.AddHeaders(req.Msg.ResponseDefinition.GetResponseHeaders(), connectErr.Meta())
		internal.AddHeaders(req.Msg.ResponseDefinition.GetResponseTrailers(), connectErr.Meta())
		return nil, connectErr
	}

	resp := connect.NewResponse(&v2.UnaryResponse{
		Payload: payload,
	})

	internal.AddHeaders(req.Msg.ResponseDefinition.ResponseHeaders, resp.Header())
	internal.AddHeaders(req.Msg.ResponseDefinition.ResponseTrailers, resp.Trailer())

	return resp, nil
}

func (s *conformanceServer) ClientStream(
	ctx context.Context,
	stream *connect.ClientStream[v2.ClientStreamRequest],
) (*connect.Response[v2.ClientStreamResponse], error) {
	var responseDefinition *v2.UnaryResponseDefinition
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

	payload, err := parseUnaryResponseDefinition(ctx, responseDefinition, stream.RequestHeader(), reqs)
	if err != nil {
		internal.AddHeaders(responseDefinition.ResponseHeaders, err.Meta())
		internal.AddHeaders(responseDefinition.ResponseTrailers, err.Meta())
		return nil, err
	}

	resp := connect.NewResponse(&v2.ClientStreamResponse{
		Payload: payload,
	})

	internal.AddHeaders(responseDefinition.ResponseHeaders, resp.Header())
	internal.AddHeaders(responseDefinition.ResponseTrailers, resp.Trailer())

	return resp, nil
}

func (s *conformanceServer) ServerStream(
	ctx context.Context,
	req *connect.Request[v2.ServerStreamRequest],
	stream *connect.ServerStream[v2.ServerStreamResponse],
) error {
	responseDefinition := req.Msg.ResponseDefinition
	if responseDefinition != nil {
		internal.AddHeaders(responseDefinition.ResponseHeaders, stream.ResponseHeader())
		internal.AddHeaders(responseDefinition.ResponseTrailers, stream.ResponseTrailer())
	}

	// Convert the request to an Any so that it can be recorded in the payload
	msgAsAny, err := asAny(req.Msg)
	if err != nil {
		return err
	}
	respNum := 0
	for _, data := range responseDefinition.ResponseData {
		resp := &v2.ServerStreamResponse{
			Payload: &v2.ConformancePayload{
				Data: data,
			},
		}

		// Only set the request info if this is the first response being sent back
		// because for server streams, nothing in the request info will change
		// after the first response.
		if respNum == 0 {
			resp.Payload.RequestInfo = createRequestInfo(ctx, req.Header(), []*anypb.Any{msgAsAny})
		}

		time.Sleep((time.Duration(responseDefinition.ResponseDelayMs) * time.Millisecond))

		if err := stream.Send(resp); err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("error sending on stream: %w", err))
		}
		respNum++
	}

	if responseDefinition.Error != nil {
		return internal.ConvertProtoToConnectError(responseDefinition.Error)
	}
	return nil
}

func (s *conformanceServer) BidiStream(
	ctx context.Context,
	stream *connect.BidiStream[v2.BidiStreamRequest, v2.BidiStreamResponse],
) error {
	var responseDefinition *v2.StreamResponseDefinition
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
			}
		}

		// If fullDuplex, then send one of the desired responses each time we get a message on the stream
		if fullDuplex {
			if respNum >= len(responseDefinition.ResponseData) {
				return connect.NewError(
					connect.CodeAborted,
					errors.New("received more requests than desired responses on a full duplex stream"),
				)
			}
			resp := &v2.BidiStreamResponse{
				Payload: &v2.ConformancePayload{
					Data: responseDefinition.ResponseData[respNum],
				},
			}
			var requestInfo *v2.ConformancePayload_RequestInfo
			if respNum == 0 {
				// Only send the full request info (including headers and timeouts)
				// in the first response
				requestInfo = createRequestInfo(ctx, stream.RequestHeader(), reqs)
			} else {
				// All responses after the first should only include the requests
				// since that is the only thing that will change between responses
				// for a full duplex stream
				requestInfo = &v2.ConformancePayload_RequestInfo{
					Requests: reqs,
				}
			}
			resp.Payload.RequestInfo = requestInfo
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
		resp := &v2.BidiStreamResponse{
			Payload: &v2.ConformancePayload{
				Data: responseDefinition.ResponseData[respNum],
			},
		}
		// Only set the request info if this is the first response being sent back
		// because for half duplex streams, nothing in the request info will change
		// after the first response (this includes the requests since they've all
		// been received by this point)
		if respNum == 0 {
			resp.Payload.RequestInfo = createRequestInfo(ctx, stream.RequestHeader(), reqs)
		}
		time.Sleep((time.Duration(responseDefinition.ResponseDelayMs) * time.Millisecond))

		if err := stream.Send(resp); err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("error sending on stream: %w", err))
		}
	}

	if responseDefinition.Error != nil {
		return internal.ConvertProtoToConnectError(responseDefinition.Error)
	}
	return nil
}

// Parses the given unary response definition and returns either
// a built payload or a connect error based on the definition.
func parseUnaryResponseDefinition(
	ctx context.Context,
	def *v2.UnaryResponseDefinition,
	hdrs http.Header,
	reqs []*anypb.Any,
) (*v2.ConformancePayload, *connect.Error) {
	if def != nil {
		switch respType := def.Response.(type) {
		case *v2.UnaryResponseDefinition_Error:
			requestInfo := createRequestInfo(ctx, hdrs, reqs)
			// details:
			//   - "@type": "connectrpc.conformance.v2.Header"
			//     name: "test error detail name"
			//     value:
			//       - "test error detail value"
			reqInfoAny, err := anypb.New(requestInfo)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, err)
			}
			respType.Error.Details = []*anypb.Any{reqInfoAny}

			return nil, internal.ConvertProtoToConnectError(respType.Error)

		case *v2.UnaryResponseDefinition_ResponseData, nil:
			requestInfo := createRequestInfo(ctx, hdrs, reqs)
			payload := &v2.ConformancePayload{
				RequestInfo: requestInfo,
			}

			// If response data was provided, set that in the payload response
			if respType, ok := respType.(*v2.UnaryResponseDefinition_ResponseData); ok {
				payload.Data = respType.ResponseData
			}
			return payload, nil
		default:
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("provided UnaryRequest.Response has an unexpected type %T", respType))
		}
	}
	return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("no response definition provided"))
}

// Creates request info for a conformance payload.
func createRequestInfo(ctx context.Context, headers http.Header, reqs []*anypb.Any) *v2.ConformancePayload_RequestInfo {
	headerInfo := internal.ConvertToProtoHeader(headers)

	var timeoutMs *int64
	if deadline, ok := ctx.Deadline(); ok {
		timeoutMs = proto.Int64(time.Until(deadline).Milliseconds())
	}

	// Set all observed request headers and requests in the response payload
	return &v2.ConformancePayload_RequestInfo{
		RequestHeaders: headerInfo,
		Requests:       reqs,
		TimeoutMs:      timeoutMs,
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
