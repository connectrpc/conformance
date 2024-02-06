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
	v1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const clientName = "connectconformance-grpcclient"

type invoker struct {
	client v1.ConformanceServiceClient
}

func (i *invoker) Invoke(
	ctx context.Context,
	req *v1.ClientCompatRequest,
) (*v1.ClientResponseResult, error) {
	// If a timeout was specified, create a derived context with that deadline
	if req.TimeoutMs != nil {
		deadlineCtx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Duration(*req.TimeoutMs)*time.Millisecond))
		ctx = deadlineCtx
		defer cancel()
	}
	switch req.Method {
	case "Unary":
		if len(req.RequestMessages) != 1 {
			return nil, errors.New("unary calls must specify exactly one request message")
		}
		resp, err := i.unary(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil
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
		return nil, errors.New("method name " + req.Method + " does not exist")
	}
}

func (i *invoker) unary(
	ctx context.Context,
	ccr *v1.ClientCompatRequest,
) (*v1.ClientResponseResult, error) {
	msg := ccr.RequestMessages[0]
	req := &v1.UnaryRequest{}
	if err := msg.UnmarshalTo(req); err != nil {
		return nil, err
	}

	// Add the specified request headers to the request
	ctx = grpcutil.AppendToOutgoingContext(ctx, ccr.RequestHeaders)

	var protoErr *v1.Error
	payloads := make([]*v1.ConformancePayload, 0, 1)

	var headerMD, trailerMD metadata.MD

	// Invoke the Unary call
	resp, err := i.client.Unary(
		ctx,
		req,
		grpc.Header(&headerMD),
		grpc.Trailer(&trailerMD),
	)
	headers := grpcutil.ConvertMetadataToProtoHeader(headerMD)
	trailers := grpcutil.ConvertMetadataToProtoHeader(trailerMD)
	if err != nil {
		// If an error was returned, convert it to a gRPC error
		protoErr = grpcutil.ConvertGrpcToProtoError(err)
	} else if resp.Payload != nil {
		// If the call was successful and there's a payload
		// add that to the response also
		payloads = append(payloads, resp.Payload)
	}

	return &v1.ClientResponseResult{
		ResponseHeaders:  headers,
		ResponseTrailers: trailers,
		Payloads:         payloads,
		Error:            protoErr,
	}, nil
}

func (i *invoker) serverStream(
	ctx context.Context,
	ccr *v1.ClientCompatRequest,
) (result *v1.ClientResponseResult, retErr error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	result = &v1.ClientResponseResult{}

	msg := ccr.RequestMessages[0]
	req := &v1.ServerStreamRequest{}
	if err := msg.UnmarshalTo(req); err != nil {
		return nil, err
	}
	// Add the specified request headers to the request
	ctx = grpcutil.AppendToOutgoingContext(ctx, ccr.RequestHeaders)

	stream, err := i.client.ServerStream(ctx, req)
	if err != nil {
		return nil, err
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

	timing, err := internal.GetCancelTiming(ccr.Cancel)
	if err != nil {
		return nil, err
	}
	// If the cancel timing specifies after 0 responses, then cancel before
	// receiving anything
	if timing.AfterNumResponses == 0 {
		cancel()
	}
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
		// If the call was successful, get the returned payloads
		// and the headers and trailers
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
	ccr *v1.ClientCompatRequest,
) (*v1.ClientResponseResult, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	result := &v1.ClientResponseResult{}

	// Add the specified request headers to the request
	ctx = grpcutil.AppendToOutgoingContext(ctx, ccr.RequestHeaders)

	stream, err := i.client.ClientStream(ctx)
	if err != nil {
		return nil, err
	}

	for _, msg := range ccr.RequestMessages {
		csr := &v1.ClientStreamRequest{}
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
		// If an error was returned, convert it to a gRPC error
		result.Error = grpcutil.ConvertGrpcToProtoError(err)
	} else {
		// If the call was successful, get the returned payloads
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
	ccr *v1.ClientCompatRequest,
) (result *v1.ClientResponseResult, retErr error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	result = &v1.ClientResponseResult{}

	// Add the specified request headers to the request
	ctx = grpcutil.AppendToOutgoingContext(ctx, ccr.RequestHeaders)

	stream, err := i.client.BidiStream(ctx)
	if err != nil {
		return nil, err
	}

	fullDuplex := ccr.StreamType == v1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM

	// Cancellation timing
	timing, err := internal.GetCancelTiming(ccr.Cancel)
	if err != nil {
		return nil, err
	}

	var protoErr *v1.Error
	totalRcvd := 0
	for _, msg := range ccr.RequestMessages {
		bsr := &v1.BidiStreamRequest{}
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
			if totalRcvd == timing.AfterNumResponses {
				cancel()
			}
			// If this is a full duplex stream, receive a response for each request
			msg, err := stream.Recv()
			totalRcvd++
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
			// If the call was successful, get the returned payloads
			result.Payloads = append(result.Payloads, msg.Payload)
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
		if totalRcvd == timing.AfterNumResponses {
			cancel()
		}
		msg, err := stream.Recv()
		totalRcvd++
		if err != nil {
			if !errors.Is(err, io.EOF) {
				// If an error was returned that is not an EOF, convert it
				// to a proto Error. If the error was an EOF, that just means
				// reads are done.
				protoErr = grpcutil.ConvertGrpcToProtoError(err)
			}
			break
		}
		// If the call was successful, save the payloads
		result.Payloads = append(result.Payloads, msg.Payload)
	}

	if protoErr != nil {
		result.Error = protoErr
		return result, nil
	}

	return result, nil
}

func (i *invoker) unimplemented(
	ctx context.Context,
	ccr *v1.ClientCompatRequest,
) (*v1.ClientResponseResult, error) {
	msg := ccr.RequestMessages[0]
	req := &v1.UnimplementedRequest{}
	if err := msg.UnmarshalTo(req); err != nil {
		return nil, err
	}

	// Add the specified request headers to the request
	ctx = grpcutil.AppendToOutgoingContext(ctx, ccr.RequestHeaders)

	// Invoke the Unary call
	_, err := i.client.Unimplemented(ctx, req)
	return &v1.ClientResponseResult{
		Error: grpcutil.ConvertGrpcToProtoError(err),
	}, nil
}

// Creates a new invoker around a ConformanceServiceClient.
func newInvoker(clientConn grpc.ClientConnInterface) *invoker {
	client := v1.NewConformanceServiceClient(clientConn)
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
