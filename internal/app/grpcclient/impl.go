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

package grpcclient

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
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

const clientName = "connectconformance-grpcclient"

type invoker struct {
	client conformancev1.ConformanceServiceClient
}

func (i *invoker) Invoke(
	ctx context.Context,
	req *conformancev1.ClientCompatRequest,
) (*conformancev1.ClientResponseResult, error) {
	// If a timeout was specified, create a derived context with that deadline
	if req.TimeoutMs != nil {
		deadlineCtx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Duration(*req.TimeoutMs)*time.Millisecond))
		ctx = deadlineCtx
		defer cancel()
	}
	switch req.GetMethod() {
	case "Unary":
		if len(req.RequestMessages) != 1 {
			return nil, errors.New("unary calls must specify exactly one request message")
		}
		resp, err := i.unary(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	case "IdempotentUnary":
		return nil, fmt.Errorf("method name %s of service %s is for use with Connect, not gRPC", req.GetMethod(), req.GetService())
	case "ServerStream":
		if len(req.RequestMessages) != 1 {
			return nil, errors.New("server streaming calls must specify exactly one request message")
		}
		resp, err := i.serverStream(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	case "ClientStream":
		resp, err := i.clientStream(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	case "BidiStream":
		resp, err := i.bidiStream(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	case "Unimplemented":
		resp, err := i.unimplemented(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	default:
		return nil, fmt.Errorf("method name %s does not exist on service %s", req.GetMethod(), req.GetService())
	}
}

func (i *invoker) unary(
	ctx context.Context,
	ccr *conformancev1.ClientCompatRequest,
) (*conformancev1.ClientResponseResult, error) {
	return doUnary(ctx, ccr,
		func(ctx context.Context, req *conformancev1.UnaryRequest, opts ...grpc.CallOption) (*conformancev1.UnaryResponse, error) {
			return i.client.Unary(ctx, req, opts...)
		},
		func(resp *conformancev1.UnaryResponse) *conformancev1.ConformancePayload {
			return resp.Payload
		},
	)
}

func (i *invoker) serverStream(
	ctx context.Context,
	ccr *conformancev1.ClientCompatRequest,
) (result *conformancev1.ClientResponseResult, retErr error) {
	timing, err := internal.GetCancelTiming(ccr.Cancel)
	if err != nil {
		return nil, err
	}

	result = &conformancev1.ClientResponseResult{}

	msg := ccr.RequestMessages[0]
	req := &conformancev1.ServerStreamRequest{}
	if err := msg.UnmarshalTo(req); err != nil {
		return nil, err
	}
	// Add the specified request headers to the request
	ctx = grpcutil.AppendToOutgoingContext(ctx, ccr.RequestHeaders)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	stream, err := i.client.ServerStream(ctx, req)
	if err != nil {
		return nil, err
	}

	if timing.AfterCloseSendMs >= 0 {
		time.Sleep(time.Duration(timing.AfterCloseSendMs) * time.Millisecond)
		cancel()
	}

	// Read headers from the stream
	hdr, err := stream.Header()
	if err != nil {
		return nil, err
	}
	defer func() {
		if result != nil {
			// Set headers and trailers from the stream
			result.ResponseHeaders = grpcutil.ConvertMetadataToProtoHeader(hdr)
			result.ResponseTrailers = grpcutil.ConvertMetadataToProtoHeader(stream.Trailer())
		}
	}()

	totalRcvd := 0
	for {
		msg, err := stream.Recv()
		totalRcvd++
		if err != nil {
			if !errors.Is(err, io.EOF) {
				result.Error = grpcutil.ConvertGrpcToProtoError(err)
			}
			break
		}
		// On successful receive, get the returned payload.
		result.Payloads = append(result.Payloads, msg.Payload)

		// If AfterNumResponses is specified, it will be a number > 0 here.
		// If it wasn't specified, it will be -1, which means the totalRcvd
		// will never be equal and we won't cancel.
		if totalRcvd == timing.AfterNumResponses {
			cancel()
		}
	}

	return result, nil
}

func (i *invoker) clientStream(
	ctx context.Context,
	ccr *conformancev1.ClientCompatRequest,
) (*conformancev1.ClientResponseResult, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	result := &conformancev1.ClientResponseResult{}

	// Add the specified request headers to the request
	ctx = grpcutil.AppendToOutgoingContext(ctx, ccr.RequestHeaders)

	stream, err := i.client.ClientStream(ctx)
	if err != nil {
		return nil, err
	}

	for _, msg := range ccr.RequestMessages {
		csr := &conformancev1.ClientStreamRequest{}
		if err := msg.UnmarshalTo(csr); err != nil {
			return nil, err
		}

		// Sleep for any specified delay
		time.Sleep(time.Duration(ccr.RequestDelayMs) * time.Millisecond)

		if err := stream.Send(csr); err != nil && errors.Is(err, io.EOF) {
			break
		}
	}

	// Cancellation timing
	timing, err := internal.GetCancelTiming(ccr.Cancel)
	if err != nil {
		return nil, err
	}
	if timing.BeforeCloseSend != nil {
		cancel()
	} else if timing.AfterCloseSendMs >= 0 {
		go func() {
			time.Sleep(time.Duration(timing.AfterCloseSendMs) * time.Millisecond)
			cancel()
		}()
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		// If an error was returned, convert it to a gRPC error.
		result.Error = grpcutil.ConvertGrpcToProtoError(err)
	} else {
		// If the call was successful, get the returned payload.
		result.Payloads = append(result.Payloads, resp.Payload)
	}

	hdr, err := stream.Header()
	if err != nil {
		return nil, err
	}
	// Set headers and trailers from the stream
	result.ResponseHeaders = grpcutil.ConvertMetadataToProtoHeader(hdr)
	result.ResponseTrailers = grpcutil.ConvertMetadataToProtoHeader(stream.Trailer())

	return result, nil
}

func (i *invoker) bidiStream(
	ctx context.Context,
	ccr *conformancev1.ClientCompatRequest,
) (result *conformancev1.ClientResponseResult, retErr error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	result = &conformancev1.ClientResponseResult{}

	// Add the specified request headers to the request
	ctx = grpcutil.AppendToOutgoingContext(ctx, ccr.RequestHeaders)

	stream, err := i.client.BidiStream(ctx)
	if err != nil {
		return nil, err
	}

	fullDuplex := ccr.StreamType == conformancev1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM

	// Cancellation timing
	timing, err := internal.GetCancelTiming(ccr.Cancel)
	if err != nil {
		return nil, err
	}

	var protoErr *conformancev1.Error
	totalRcvd := 0
	for _, msg := range ccr.RequestMessages {
		bsr := &conformancev1.BidiStreamRequest{}
		if err := msg.UnmarshalTo(bsr); err != nil {
			// Return the error and nil result because this is an
			// unmarshalling error unrelated to the RPC
			return nil, err
		}
		if err := stream.Send(bsr); err != nil && errors.Is(err, io.EOF) {
			// Call receive to get the error and convert it to a proto error
			if _, recvErr := stream.Recv(); recvErr != nil {
				protoErr = grpcutil.ConvertGrpcToProtoError(recvErr)
			} else {
				// Just in case the receive call doesn't return the error,
				// use the error returned from Send. Note this should never
				// happen, but is here as a safeguard.
				protoErr = grpcutil.ConvertGrpcToProtoError(err)
			}
			// Break the send loop
			break
		}
		if fullDuplex {
			// If this is a full duplex stream, receive a response for each request
			msg, err := stream.Recv()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					// If an error was returned that is not an EOF, convert it
					// to a proto Error. If the error was an EOF, that just means
					// reads are done.
					protoErr = grpcutil.ConvertGrpcToProtoError(err)
				}
				// Reads are done either because we received an error or an EOF
				// In either case, break the outer loop
				break
			}
			// On successful receive, get the returned payload.
			result.Payloads = append(result.Payloads, msg.Payload)
			totalRcvd++
			if totalRcvd == timing.AfterNumResponses {
				cancel()
			}
		}
	}

	var hdr metadata.MD
	defer func() {
		if result != nil {
			// Set headers and trailers from the stream
			result.ResponseHeaders = grpcutil.ConvertMetadataToProtoHeader(hdr)
			result.ResponseTrailers = grpcutil.ConvertMetadataToProtoHeader(stream.Trailer())
		}
	}()

	if timing.BeforeCloseSend != nil {
		cancel()
	}

	// Sends are done, close the send side of the stream
	err = stream.CloseSend()
	if err != nil && protoErr == nil {
		protoErr = grpcutil.ConvertGrpcToProtoError(err)
	}

	if timing.AfterCloseSendMs >= 0 {
		time.Sleep(time.Duration(timing.AfterCloseSendMs) * time.Millisecond)
		cancel()
	}

	// Once the send side is closed, header metadata is ready to be read
	hdr, err = stream.Header()
	if err != nil && protoErr == nil {
		protoErr = grpcutil.ConvertGrpcToProtoError(err)
	}

	// If we received an error in any of the send logic or full-duplex reads, then exit
	if protoErr != nil {
		result.Error = protoErr
		return result, nil
	}

	// Receive any remaining responses
	for {
		msg, err := stream.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				// If an error was returned that is not an EOF, convert it
				// to a proto Error. If the error was an EOF, that just means
				// reads are done.
				protoErr = grpcutil.ConvertGrpcToProtoError(err)
			}
			break
		}
		// On successful receive, get the returned payload.
		result.Payloads = append(result.Payloads, msg.Payload)
		totalRcvd++
		if totalRcvd == timing.AfterNumResponses {
			cancel()
		}
	}

	if protoErr != nil {
		result.Error = protoErr
		return result, nil
	}

	return result, nil
}

func (i *invoker) unimplemented(
	ctx context.Context,
	ccr *conformancev1.ClientCompatRequest,
) (*conformancev1.ClientResponseResult, error) {
	return doUnary(ctx, ccr,
		func(ctx context.Context, req *conformancev1.UnimplementedRequest, opts ...grpc.CallOption) (*conformancev1.UnimplementedResponse, error) {
			return i.client.Unimplemented(ctx, req, opts...)
		},
		func(_ *conformancev1.UnimplementedResponse) *conformancev1.ConformancePayload {
			return nil
		},
	)
}

type pointerMessage[T any] interface {
	*T
	proto.Message
}

func doUnary[ReqT, RespT any, Req pointerMessage[ReqT], Resp pointerMessage[RespT]](
	ctx context.Context,
	ccr *conformancev1.ClientCompatRequest,
	stub func(context.Context, Req, ...grpc.CallOption) (Resp, error),
	getPayload func(Resp) *conformancev1.ConformancePayload,
) (*conformancev1.ClientResponseResult, error) {
	timing, err := internal.GetCancelTiming(ccr.Cancel)
	if err != nil {
		return nil, err
	}

	msg := ccr.RequestMessages[0]
	req := Req(new(ReqT))
	if err := msg.UnmarshalTo(req); err != nil {
		return nil, err
	}

	// Add the specified request headers to the request
	ctx = grpcutil.AppendToOutgoingContext(ctx, ccr.RequestHeaders)

	var protoErr *conformancev1.Error
	payloads := make([]*conformancev1.ConformancePayload, 0, 1)

	var headerMD, trailerMD metadata.MD

	if timing.AfterCloseSendMs >= 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		time.AfterFunc(time.Duration(timing.AfterCloseSendMs)*time.Millisecond, cancel)
	}
	// Invoke the Unary call
	resp, err := stub(
		ctx,
		req,
		grpc.Header(&headerMD),
		grpc.Trailer(&trailerMD),
	)
	headers := grpcutil.ConvertMetadataToProtoHeader(headerMD)
	trailers := grpcutil.ConvertMetadataToProtoHeader(trailerMD)
	if err != nil {
		// If an error was returned, convert it to a gRPC error.
		protoErr = grpcutil.ConvertGrpcToProtoError(err)
	} else {
		// If the call was successful, get the returned payload.
		payloads = append(payloads, getPayload(resp))
	}

	return &conformancev1.ClientResponseResult{
		ResponseHeaders:  headers,
		ResponseTrailers: trailers,
		Payloads:         payloads,
		Error:            protoErr,
	}, nil
}

// Creates a new invoker around a ConformanceServiceClient.
func newInvoker(clientConn grpc.ClientConnInterface) *invoker {
	client := conformancev1.NewConformanceServiceClient(clientConn)
	return &invoker{
		client: client,
	}
}

func userAgentUnaryClientInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return invoker(addUserAgent(ctx), method, req, reply, cc, opts...)
}

func userAgentStreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return streamer(addUserAgent(ctx), desc, cc, method, opts...)
}

func addUserAgent(ctx context.Context) context.Context {
	reqMD, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		reqMD = metadata.MD{}
	}
	// decorate user-agent with the program name and version
	userAgent := fmt.Sprintf("%s %s/%s", reqMD.Get("User-Agent"), clientName, internal.Version)
	reqMD.Set("User-Agent", userAgent)
	return metadata.NewOutgoingContext(ctx, reqMD)
}
