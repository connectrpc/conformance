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

package grpcserver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"connectrpc.com/conformance/internal"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	proto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const serverName = "connectconformance-grpcserver"

// NewConformanceServiceServer creates a new Conformance Service server.
func NewConformanceServiceServer() conformancev1.ConformanceServiceServer {
	return &conformanceServiceServer{}
}

type conformanceServiceServer struct {
	conformancev1.UnimplementedConformanceServiceServer
}

func (c *conformanceServiceServer) Unary(
	ctx context.Context,
	req *conformancev1.UnaryRequest,
) (*conformancev1.UnaryResponse, error) {
	msgAsAny, err := asAny(req)
	if err != nil {
		return nil, err
	}

	responseDefinition := req.ResponseDefinition
	if responseDefinition != nil {
		headerMD := grpcutil.ConvertProtoHeaderToMetadata(req.ResponseDefinition.ResponseHeaders)
		if err := grpc.SendHeader(ctx, headerMD); err != nil {
			return nil, err
		}
		trailerMD := grpcutil.ConvertProtoHeaderToMetadata(req.ResponseDefinition.ResponseTrailers)
		if err := grpc.SetTrailer(ctx, trailerMD); err != nil {
			return nil, err
		}
		time.Sleep(time.Duration(req.ResponseDefinition.ResponseDelayMs) * time.Millisecond)
	}

	md, _ := metadata.FromIncomingContext(ctx)
	payload, grpcErr := parseUnaryResponseDefinition(
		ctx,
		responseDefinition,
		md,
		[]*anypb.Any{msgAsAny},
	)
	if grpcErr != nil {
		return nil, grpcErr
	}

	return &conformancev1.UnaryResponse{
		Payload: payload,
	}, nil
}

func (c *conformanceServiceServer) ClientStream(
	stream conformancev1.ConformanceService_ClientStreamServer,
) error {
	ctx := stream.Context()
	var responseDefinition *conformancev1.UnaryResponseDefinition
	firstRecv := true
	var reqs []*anypb.Any

	for {
		if err := stream.Context().Err(); err != nil {
			return err
		}
		msg, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		// If this is the first message received on the stream, save off the response definition we need to send
		if firstRecv {
			responseDefinition = msg.ResponseDefinition
			firstRecv = false
		}
		// Record all the requests received
		msgAsAny, err := asAny(msg)
		if err != nil {
			return err
		}
		reqs = append(reqs, msgAsAny)
	}

	// Set headers and trailers on stream
	if responseDefinition != nil {
		headerMD := grpcutil.ConvertProtoHeaderToMetadata(responseDefinition.ResponseHeaders)
		if err := stream.SendHeader(headerMD); err != nil {
			return err
		}

		trailerMD := grpcutil.ConvertProtoHeaderToMetadata(responseDefinition.ResponseTrailers)
		stream.SetTrailer(trailerMD)

		time.Sleep(time.Duration(responseDefinition.ResponseDelayMs) * time.Millisecond)
	}

	md, _ := metadata.FromIncomingContext(stream.Context())
	payload, err := parseUnaryResponseDefinition(ctx, responseDefinition, md, reqs)
	if err != nil {
		return err
	}

	return stream.SendAndClose(&conformancev1.ClientStreamResponse{
		Payload: payload,
	})
}

func (c *conformanceServiceServer) ServerStream(
	req *conformancev1.ServerStreamRequest,
	stream conformancev1.ConformanceService_ServerStreamServer,
) error {
	ctx := stream.Context()
	// Convert the request to an Any so that it can be recorded in the payload
	msgAsAny, err := asAny(req)
	if err != nil {
		return err
	}

	respNum := 0

	responseDefinition := req.ResponseDefinition
	if responseDefinition != nil { //nolint:nestif
		headerMD := grpcutil.ConvertProtoHeaderToMetadata(responseDefinition.ResponseHeaders)
		// Immediately send the headers on the stream so that metadata can be read by the client
		if err := stream.SendHeader(headerMD); err != nil {
			return err
		}

		trailerMD := grpcutil.ConvertProtoHeaderToMetadata(responseDefinition.ResponseTrailers)
		stream.SetTrailer(trailerMD)

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
				requestMetadata, _ := metadata.FromIncomingContext(stream.Context())
				requestInfo := createRequestInfo(ctx, requestMetadata, []*anypb.Any{msgAsAny})
				resp.Payload.RequestInfo = requestInfo
			}

			time.Sleep(time.Duration(responseDefinition.ResponseDelayMs) * time.Millisecond)

			if err := stream.Send(resp); err != nil {
				return status.Errorf(codes.Internal, "error sending on stream: %s", err.Error())
			}
			respNum++
		}
		if responseDefinition.Error != nil {
			if respNum == 0 {
				// We've sent no responses and are returning an error, so build a
				// RequestInfo message and append to the error details
				requestMetadata, _ := metadata.FromIncomingContext(stream.Context())
				reqInfo := createRequestInfo(ctx, requestMetadata, []*anypb.Any{msgAsAny})
				reqInfoAny, err := anypb.New(reqInfo)
				if err != nil {
					return status.Error(codes.Internal, err.Error())
				}
				responseDefinition.Error.Details = append(responseDefinition.Error.Details, reqInfoAny)
			}
			return grpcutil.ConvertProtoToGrpcError(responseDefinition.Error)
		}
	}

	return nil
}

func (c *conformanceServiceServer) BidiStream(
	stream conformancev1.ConformanceService_BidiStreamServer,
) error {
	ctx := stream.Context()
	var responseDefinition *conformancev1.StreamResponseDefinition
	fullDuplex := false
	firstRecv := true
	respNum := 0
	var reqs []*anypb.Any
	for {
		if err := stream.Context().Err(); err != nil {
			return err
		}
		req, err := stream.Recv()
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

			if responseDefinition != nil {
				headerMD := grpcutil.ConvertProtoHeaderToMetadata(responseDefinition.ResponseHeaders)
				// Immediately send the headers on the stream so that metadata can be read by the client
				if err := stream.SendHeader(headerMD); err != nil {
					return err
				}

				trailerMD := grpcutil.ConvertProtoHeaderToMetadata(responseDefinition.ResponseTrailers)
				stream.SetTrailer(trailerMD)
			}
		}

		// If fullDuplex, then send one of the desired responses each time we get a message on the stream
		if fullDuplex {
			if responseDefinition == nil || respNum >= len(responseDefinition.ResponseData) {
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
				requestMetadata, _ := metadata.FromIncomingContext(stream.Context())
				requestInfo = createRequestInfo(ctx, requestMetadata, reqs)
			} else {
				// All responses after the first should only include the requests
				// since that is the only thing that will change between responses
				// for a full duplex stream
				requestInfo = &conformancev1.ConformancePayload_RequestInfo{
					Requests: reqs,
				}
			}
			resp.Payload.RequestInfo = requestInfo
			time.Sleep(time.Duration(responseDefinition.ResponseDelayMs) * time.Millisecond)

			if err := stream.Send(resp); err != nil {
				return status.Errorf(codes.Internal, "error sending on stream: %s", err.Error())
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
			if err := stream.Context().Err(); err != nil {
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
				requestMetadata, _ := metadata.FromIncomingContext(stream.Context())
				resp.Payload.RequestInfo = createRequestInfo(ctx, requestMetadata, reqs)
			}
			time.Sleep(time.Duration(responseDefinition.ResponseDelayMs) * time.Millisecond)

			if err := stream.Send(resp); err != nil {
				return status.Errorf(codes.Internal, "error sending on stream: %s", err.Error())
			}
		}

		if responseDefinition.Error != nil {
			if respNum == 0 {
				// We've sent no responses and are returning an error, so build a
				// RequestInfo message and append to the error details
				requestMetadata, _ := metadata.FromIncomingContext(stream.Context())
				reqInfo := createRequestInfo(ctx, requestMetadata, reqs)
				reqInfoAny, err := anypb.New(reqInfo)
				if err != nil {
					return status.Error(codes.Internal, err.Error())
				}
				responseDefinition.Error.Details = append(responseDefinition.Error.Details, reqInfoAny)
			}
			return grpcutil.ConvertProtoToGrpcError(responseDefinition.Error)
		}
	}
	return nil
}

// Parses the given unary response definition and returns either
// a built payload or a gRPC error based on the definition.
func parseUnaryResponseDefinition(
	ctx context.Context,
	def *conformancev1.UnaryResponseDefinition,
	metadata metadata.MD,
	reqs []*anypb.Any,
) (*conformancev1.ConformancePayload, error) {
	reqInfo := createRequestInfo(ctx, metadata, reqs)
	if def == nil {
		// If the definition is not set at all, there's nothing to respond with.
		// Just return a payload with the request info
		return &conformancev1.ConformancePayload{
			RequestInfo: reqInfo,
		}, nil
	}
	switch respType := def.Response.(type) {
	case *conformancev1.UnaryResponseDefinition_Error:
		reqInfoAny, err := anypb.New(reqInfo)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		respType.Error.Details = append(respType.Error.Details, reqInfoAny)
		return nil, grpcutil.ConvertProtoToGrpcError(respType.Error)

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
		return nil, status.Errorf(
			codes.InvalidArgument,
			"provided UnaryRequest.Response has an unexpected type %T",
			respType,
		)
	}
}

// Creates request info for a conformance payload.
func createRequestInfo(ctx context.Context, metadata metadata.MD, reqs []*anypb.Any) *conformancev1.ConformancePayload_RequestInfo {
	headerInfo := grpcutil.ConvertMetadataToProtoHeader(metadata)

	var timeoutMs *int64
	if deadline, ok := ctx.Deadline(); ok {
		timeout := time.Until(deadline)
		if timeout < 0 {
			timeout = 0
		}
		timeoutMs = proto.Int64(timeout.Milliseconds())
	}

	// Set all observed request headers and requests in the response payload
	return &conformancev1.ConformancePayload_RequestInfo{
		RequestHeaders: headerInfo,
		Requests:       reqs,
		TimeoutMs:      timeoutMs,
	}
}

// Converts the given message to an Any.
func asAny(msg proto.Message) (*anypb.Any, error) {
	msgAsAny, err := anypb.New(msg)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"unable to convert message: %s",
			err.Error(),
		)
	}
	return msgAsAny, nil
}

func serverNameUnaryInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	_ = grpc.SetHeader(ctx, serverNameMetadata())
	return handler(ctx, req)
}

func serverNameStreamInterceptor(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	_ = ss.SetHeader(serverNameMetadata())
	return handler(srv, ss)
}

func serverNameMetadata() metadata.MD {
	server := fmt.Sprintf("%s/%s", serverName, internal.Version)
	return metadata.Pairs("Server", server)
}
