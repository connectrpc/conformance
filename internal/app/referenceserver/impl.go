// Copyright 2023-2024 The Connect Authors
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
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/conformance/internal"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1/conformancev1connect"
	"connectrpc.com/connect"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const serverName = "connectconformance-referenceserver"

// ConformanceRequest is a general interface for all conformance requests (UnaryRequest, ServerStreamRequest, etc.)
type ConformanceRequest interface {
	GetResponseHeaders() []*conformancev1.Header
	GetResponseTrailers() []*conformancev1.Header
}

type conformanceServer struct {
	conformancev1connect.UnimplementedConformanceServiceHandler
	referenceMode bool
}

func (s *conformanceServer) Unary(
	ctx context.Context,
	req *connect.Request[conformancev1.UnaryRequest],
) (*connect.Response[conformancev1.UnaryResponse], error) {
	return doUnary(ctx, req, s.referenceMode, func(payload *conformancev1.ConformancePayload) *conformancev1.UnaryResponse {
		return &conformancev1.UnaryResponse{
			Payload: payload,
		}
	})
}

func (s *conformanceServer) IdempotentUnary(
	ctx context.Context,
	req *connect.Request[conformancev1.IdempotentUnaryRequest],
) (*connect.Response[conformancev1.IdempotentUnaryResponse], error) {
	return doUnary(ctx, req, s.referenceMode, func(payload *conformancev1.ConformancePayload) *conformancev1.IdempotentUnaryResponse {
		return &conformancev1.IdempotentUnaryResponse{
			Payload: payload,
		}
	})
}

type hasUnaryResponseDefinition[T any] interface {
	*T
	proto.Message
	GetResponseDefinition() *conformancev1.UnaryResponseDefinition
}

func doUnary[ReqT, RespT any, Req hasUnaryResponseDefinition[ReqT]](
	ctx context.Context,
	req *connect.Request[ReqT],
	referenceMode bool,
	makeResp func(payload *conformancev1.ConformancePayload) *RespT,
) (*connect.Response[RespT], error) {
	msg := Req(req.Msg)
	msgAsAny, err := asAny(msg)
	if err != nil {
		return nil, err
	}
	payload, connectErr := parseUnaryResponseDefinition(
		ctx,
		referenceMode,
		msg.GetResponseDefinition(),
		req.Header(),
		req.Peer().Query,
		req.Peer().Protocol,
		req.Header().Get("Content-Type"),
		nil,
		[]*anypb.Any{msgAsAny},
	)
	if connectErr != nil {
		return nil, connectErr
	}

	resp := connect.NewResponse(makeResp(payload))

	if msg.GetResponseDefinition() != nil {
		internal.AddHeaders(msg.GetResponseDefinition().ResponseHeaders, resp.Header())
		internal.AddHeaders(msg.GetResponseDefinition().ResponseTrailers, resp.Trailer())

		// If a response delay was specified, sleep for that amount of ms before responding
		responseDelay := time.Duration(msg.GetResponseDefinition().ResponseDelayMs) * time.Millisecond
		time.Sleep(responseDelay)
	}

	return resp, nil
}

func (s *conformanceServer) ClientStream(
	ctx context.Context,
	stream *connect.ClientStream[conformancev1.ClientStreamRequest],
) (*connect.Response[conformancev1.ClientStreamResponse], error) {
	var responseDefinition *conformancev1.UnaryResponseDefinition
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
		s.referenceMode,
		responseDefinition,
		stream.RequestHeader(),
		stream.Peer().Query,
		stream.Peer().Protocol,
		stream.RequestHeader().Get("Content-Type"),
		stream.Conn(),
		reqs,
	)
	if err != nil {
		return nil, err
	}

	resp := connect.NewResponse(&conformancev1.ClientStreamResponse{
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
	req *connect.Request[conformancev1.ServerStreamRequest],
	stream *connect.ServerStream[conformancev1.ServerStreamResponse],
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

		if len(responseDefinition.ResponseData) > 0 {
			// Immediately send the headers/trailers on the stream so that they can be read by the client
			if err := stream.Send(nil); err != nil {
				return connect.NewError(connect.CodeInternal, fmt.Errorf("error sending on stream: %w", err))
			}
		}

		// Calculate the response delay if specified
		responseDelay := time.Duration(responseDefinition.ResponseDelayMs) * time.Millisecond

		for _, data := range responseDefinition.ResponseData {
			resp := &conformancev1.ServerStreamResponse{
				Payload: &conformancev1.ConformancePayload{
					Data: data,
				},
			}

			// Only set the request info if this is the first response being sent back
			// because for server streams, nothing in the request info will change
			// after the first response.
			if respNum == 0 {
				resp.Payload.RequestInfo = createRequestInfo(ctx, req.Header(), req.Peer().Query, []*anypb.Any{msgAsAny})
			}

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
	stream *connect.BidiStream[conformancev1.BidiStreamRequest, conformancev1.BidiStreamResponse],
) error {
	var responseDefinition *conformancev1.StreamResponseDefinition
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
		if firstRecv { //nolint:nestif
			responseDefinition = req.ResponseDefinition
			fullDuplex = req.FullDuplex
			firstRecv = false

			// If a response definition was provided, add the headers and trailers
			if responseDefinition != nil {
				internal.AddHeaders(responseDefinition.ResponseHeaders, stream.ResponseHeader())
				internal.AddHeaders(responseDefinition.ResponseTrailers, stream.ResponseTrailer())

				if fullDuplex && len(responseDefinition.ResponseData) > 0 {
					// Immediately send the headers on the stream so that they can be read by the client.
					// We can only do this for full-duplex. For half-duplex operation, we must let client
					// complete its upload before trying to send anything.
					if err := stream.Send(nil); err != nil {
						return connect.NewError(connect.CodeInternal, fmt.Errorf("error sending on stream: %w", err))
					}
				}

				// Calculate a response delay if specified
				responseDelay = time.Duration(responseDefinition.ResponseDelayMs) * time.Millisecond
			}
		}

		// If fullDuplex, then send one of the desired responses each time we get a message on the stream
		if fullDuplex {
			if respNum >= len(responseDefinition.GetResponseData()) {
				// If there are no responses to send, then break the receive loop
				// and throw the error specified
				break
			}

			resp := &conformancev1.BidiStreamResponse{
				Payload: &conformancev1.ConformancePayload{
					Data: responseDefinition.ResponseData[respNum],
				},
			}
			var requestInfo *conformancev1.ConformancePayload_RequestInfo
			if respNum == 0 {
				// Only send the full request info (including headers and timeouts)
				// in the first response
				requestInfo = createRequestInfo(ctx, stream.RequestHeader(), stream.Peer().Query, reqs)
			} else {
				// All responses after the first should only include the requests
				// since that is the only thing that will change between responses
				// for a full duplex stream
				requestInfo = &conformancev1.ConformancePayload_RequestInfo{
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

	if !fullDuplex && len(responseDefinition.GetResponseData()) > 0 {
		// Now that upload is complete, we can immediately send headers for half-duplex calls.
		if err := stream.Send(nil); err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("error sending on stream: %w", err))
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
			resp := &conformancev1.BidiStreamResponse{
				Payload: &conformancev1.ConformancePayload{
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
	referenceMode bool,
	def *conformancev1.UnaryResponseDefinition,
	hdrs http.Header,
	queryParams url.Values,
	protocol string,
	contentType string,
	conn connect.StreamingHandlerConn,
	reqs []*anypb.Any,
) (*conformancev1.ConformancePayload, *connect.Error) {
	reqInfo := createRequestInfo(ctx, hdrs, queryParams, reqs)
	if def == nil {
		// If the definition is not set at all, there's nothing to respond with.
		// Just return a payload with the request info
		return &conformancev1.ConformancePayload{
			RequestInfo: reqInfo,
		}, nil
	}

	switch respType := def.Response.(type) {
	case *conformancev1.UnaryResponseDefinition_Error:
		// The server should add the request info to the error details
		// for unary responses that return an error.
		reqInfoAny, err := anypb.New(reqInfo)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		respType.Error.Details = append(respType.Error.Details, reqInfoAny)

		connectErr := internal.ConvertProtoToConnectError(respType.Error)

		if !referenceMode { //nolint:nestif
			// The connect-go APIs don't provide a way to set headers and
			// trailers independently in the face of an error for a unary
			// response.
			//
			// In normal mode, we go with the flow of these APIs and use
			// connectErr.Meta to provide all metadata.
			internal.AddHeaders(def.GetResponseHeaders(), connectErr.Meta())
			internal.AddHeaders(def.GetResponseTrailers(), connectErr.Meta())
		} else {
			// In reference mode, however, it is counter-intuitive that an HTTP
			// trace from the reference server shows headers and trailers all
			// jumbled together. While that is semantically fine (and handled
			// by the conformance test runner's assertion of the header and
			// trailer results), it is confusing to see when someone is
			// debugging a failing test case.
			//
			// So, in reference mode, we will go against the grain of the
			// connect-go APIs and forcibly set headers and trailers.
			if conn != nil {
				// For client stream operations, we can set the headers and
				// trailers independently via the underlying streaming conn.
				internal.AddHeaders(def.GetResponseHeaders(), conn.ResponseHeader())
				internal.AddHeaders(def.GetResponseTrailers(), conn.ResponseTrailer())
			} else {
				// Not a stream? It gets trickier. There's just *no way* in the
				// connect-go API to independently set headers and trailers in the face
				// of an error. So we have to get a bit more clever.
				switch protocol {
				case connect.ProtocolConnect:
					// For the connect unary protocol, everything will end up in HTTP
					// headers. So, to distinguish headers from trailers, we can prefix
					// the trailers with "trailer-".
					internal.AddHeaders(def.GetResponseHeaders(), connectErr.Meta())
					for _, hdr := range def.GetResponseTrailers() {
						hdr.Name = "trailer-" + hdr.Name
					}
					internal.AddHeaders(def.GetResponseTrailers(), connectErr.Meta())
				case connect.ProtocolGRPC:
					// For gRPC and gRPC-web, we resort to hacking in a raw response,
					// which gives us much greater control over the response.
					rawResp := makeRawGRPCResponse(connectErr, contentType, def.GetResponseHeaders(), def.GetResponseTrailers())
					if err := setRawResponse(ctx, rawResp); err != nil {
						return nil, connect.NewError(connect.CodeUnknown, err)
					}
				case connect.ProtocolGRPCWeb:
					if len(def.GetResponseHeaders()) == 0 {
						// For gRPC-web, if there are no custom headers, then we don't have
						// to do anything special: the connect-go framework will send a
						// trailers-only response, so any error metadata will be interpreted
						// as "trailers".
						internal.AddHeaders(def.GetResponseTrailers(), connectErr.Meta())
					} else {
						// But otherwise, we have to employ the same tactics as for gRPC above.
						rawResp := makeRawGRPCWebResponse(connectErr, contentType, def.GetResponseHeaders(), def.GetResponseTrailers())
						if err := setRawResponse(ctx, rawResp); err != nil {
							return nil, connect.NewError(connect.CodeUnknown, err)
						}
					}
				default:
					return nil, connect.NewError(connect.CodeUnknown, fmt.Errorf("unrecognized protocol: %s", protocol))
				}
			}
		}

		return nil, connectErr

	case *conformancev1.UnaryResponseDefinition_ResponseData, nil:
		payload := &conformancev1.ConformancePayload{
			RequestInfo: reqInfo,
		}

		// If response data was provided, set that in the payload response
		if respType, ok := respType.(*conformancev1.UnaryResponseDefinition_ResponseData); ok {
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
) *conformancev1.ConformancePayload_RequestInfo {
	headerInfo := internal.ConvertToProtoHeader(headers)

	var connectGetInfo *conformancev1.ConformancePayload_ConnectGetInfo
	if len(queryParams) > 0 {
		queryParamInfo := internal.ConvertToProtoHeader(queryParams)

		connectGetInfo = &conformancev1.ConformancePayload_ConnectGetInfo{
			QueryParams: queryParamInfo,
		}
	}

	var timeoutMs *int64
	if deadline, ok := ctx.Deadline(); ok {
		timeoutMs = proto.Int64(time.Until(deadline).Milliseconds())
	}

	// Set all observed request headers and requests in the response payload
	return &conformancev1.ConformancePayload_RequestInfo{
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

func makeRawGRPCResponse(err *connect.Error, contentType string, headers, trailers []*conformancev1.Header) *conformancev1.RawHTTPResponse {
	return &conformancev1.RawHTTPResponse{
		Headers:  append(headers, &conformancev1.Header{Name: "Content-Type", Value: []string{contentType}}),
		Trailers: append(trailers, grpcStatusTrailers(err)...),
	}
}

func makeRawGRPCWebResponse(err *connect.Error, contentType string, headers, trailers []*conformancev1.Header) *conformancev1.RawHTTPResponse {
	return &conformancev1.RawHTTPResponse{
		Headers: append(
			headers,
			&conformancev1.Header{Name: "Content-Type", Value: []string{contentType}},
			&conformancev1.Header{Name: "Access-Control-Allow-Origin", Value: []string{"*"}},
			&conformancev1.Header{Name: "Access-Control-Expose-Headers", Value: []string{"*"}},
		),
		Body: &conformancev1.RawHTTPResponse_Stream{
			Stream: &conformancev1.StreamContents{
				Items: []*conformancev1.StreamContents_StreamItem{
					{
						Flags: 128, // indicates end-of-stream message w/ trailers
						Payload: &conformancev1.MessageContents{
							Data: &conformancev1.MessageContents_Text{
								Text: grpcWebStatusEndStream(err, trailers),
							},
						},
					},
				},
			},
		},
	}
}

func grpcStatusTrailers(err *connect.Error) []*conformancev1.Header {
	trailers := []*conformancev1.Header{
		{
			Name:  "grpc-status",
			Value: []string{fmt.Sprintf("%d", err.Code())},
		},
		{
			Name:  "grpc-message",
			Value: []string{err.Message()},
		},
	}
	if len(err.Details()) > 0 {
		statProto := &status.Status{
			Code:    int32(err.Code()),
			Message: err.Message(),
			Details: make([]*anypb.Any, len(err.Details())),
		}
		for i, detail := range err.Details() {
			statProto.Details[i] = &anypb.Any{
				TypeUrl: internal.DefaultAnyResolverPrefix + detail.Type(),
				Value:   detail.Bytes(),
			}
		}
		data, marshalErr := proto.Marshal(statProto)
		if marshalErr == nil {
			trailers = append(trailers, &conformancev1.Header{
				Name:  "grpc-status-details-bin",
				Value: []string{base64.RawStdEncoding.EncodeToString(data)},
			})
		}
	}
	return trailers
}

func grpcWebStatusEndStream(err *connect.Error, trailers []*conformancev1.Header) string {
	trailers = append(grpcStatusTrailers(err), trailers...)
	var buf bytes.Buffer
	for _, trailer := range trailers {
		for _, val := range trailer.Value {
			_, _ = fmt.Fprintf(&buf, "%s: %s\r\n", strings.ToLower(trailer.Name), val)
		}
	}
	return buf.String()
}
