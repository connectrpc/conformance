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

package grpcclient

import (
	"context"
	"errors"
	"io"
	"time"

	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"connectrpc.com/conformance/internal/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type invoker struct {
	client v1alpha1.ConformanceServiceClient
}

func (i *invoker) Invoke(
	ctx context.Context,
	req *v1alpha1.ClientCompatRequest,
) (*v1alpha1.ClientResponseResult, error) {
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
	default:
		return nil, errors.New("method name " + req.Method + " does not exist")
	}
}

func (i *invoker) unary(
	ctx context.Context,
	ccr *v1alpha1.ClientCompatRequest,
) (*v1alpha1.ClientResponseResult, error) {
	msg := ccr.RequestMessages[0]
	req := &v1alpha1.UnaryRequest{}
	if err := msg.UnmarshalTo(req); err != nil {
		return nil, err
	}

	// Add the specified request headers to the request
	ctx = grpcutil.AppendToOutgoingContext(ctx, ccr.RequestHeaders)

	var protoErr *v1alpha1.Error
	payloads := make([]*v1alpha1.ConformancePayload, 0, 1)

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
	} else {
		// If the call was successful, get the returned payloads
		payloads = append(payloads, resp.Payload)
	}

	return &v1alpha1.ClientResponseResult{
		ResponseHeaders:  headers,
		ResponseTrailers: trailers,
		Payloads:         payloads,
		Error:            protoErr,
		ConnectErrorRaw:  nil, // TODO
	}, nil
}

func (i *invoker) serverStream(
	ctx context.Context,
	ccr *v1alpha1.ClientCompatRequest,
) (*v1alpha1.ClientResponseResult, error) {
	msg := ccr.RequestMessages[0]
	req := &v1alpha1.ServerStreamRequest{}
	if err := msg.UnmarshalTo(req); err != nil {
		return nil, err
	}

	// Add the specified request headers to the request
	ctx = grpcutil.AppendToOutgoingContext(ctx, ccr.RequestHeaders)

	stream, err := i.client.ServerStream(ctx, req)
	if err != nil {
		return nil, err
	}
	// Read headers and trailers from the stream
	hdr, err := stream.Header()
	if err != nil {
		return nil, err
	}
	headers := grpcutil.ConvertMetadataToProtoHeader(hdr)
	trailers := grpcutil.ConvertMetadataToProtoHeader(stream.Trailer())

	var protoErr *v1alpha1.Error
	payloads := make([]*v1alpha1.ConformancePayload, 0, len(req.ResponseDefinition.ResponseData))

	for {
		msg, err := stream.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				protoErr = grpcutil.ConvertGrpcToProtoError(err)
			}
			break
		}
		// If the call was successful, get the returned payloads
		// and the headers and trailers
		payloads = append(payloads, msg.Payload)
	}

	return &v1alpha1.ClientResponseResult{
		ResponseHeaders:  headers,
		ResponseTrailers: trailers,
		Payloads:         payloads,
		Error:            protoErr,
		ConnectErrorRaw:  nil, // TODO
	}, nil
}

func (i *invoker) clientStream(
	ctx context.Context,
	ccr *v1alpha1.ClientCompatRequest,
) (*v1alpha1.ClientResponseResult, error) {
	// Add the specified request headers to the request
	ctx = grpcutil.AppendToOutgoingContext(ctx, ccr.RequestHeaders)

	stream, err := i.client.ClientStream(ctx)
	if err != nil {
		return nil, err
	}

	// Read headers and trailers from the stream
	hdr, err := stream.Header()
	if err != nil {
		return nil, err
	}
	headers := grpcutil.ConvertMetadataToProtoHeader(hdr)
	trailers := grpcutil.ConvertMetadataToProtoHeader(stream.Trailer())

	for _, msg := range ccr.RequestMessages {
		csr := &v1alpha1.ClientStreamRequest{}
		if err := msg.UnmarshalTo(csr); err != nil {
			return nil, err
		}

		// Sleep for any specified delay
		time.Sleep(time.Duration(ccr.RequestDelayMs) * time.Millisecond)

		if err := stream.Send(csr); err != nil && errors.Is(err, io.EOF) {
			break
		}
	}

	var protoErr *v1alpha1.Error
	payloads := make([]*v1alpha1.ConformancePayload, 0, 1)

	resp, err := stream.CloseAndRecv()
	if err != nil {
		// If an error was returned, convert it to a gRPC error
		protoErr = grpcutil.ConvertGrpcToProtoError(err)
	} else {
		// If the call was successful, get the returned payloads
		payloads = append(payloads, resp.Payload)
	}

	return &v1alpha1.ClientResponseResult{
		ResponseHeaders:  headers,
		ResponseTrailers: trailers,
		Payloads:         payloads,
		Error:            protoErr,
		ConnectErrorRaw:  nil, // TODO
	}, nil
}

func (i *invoker) bidiStream(
	ctx context.Context,
	ccr *v1alpha1.ClientCompatRequest,
) (result *v1alpha1.ClientResponseResult, retErr error) {
	result = &v1alpha1.ClientResponseResult{
		ConnectErrorRaw: nil, // TODO
	}

	// Add the specified request headers to the request
	ctx = grpcutil.AppendToOutgoingContext(ctx, ccr.RequestHeaders)

	stream, err := i.client.BidiStream(ctx)
	if err != nil {
		return nil, err
	}
	// Read headers and trailers from the stream
	hdr, err := stream.Header()
	if err != nil {
		return nil, err
	}
	defer func() {
		if result != nil {
			// Read headers and trailers from the stream
			result.ResponseHeaders = grpcutil.ConvertMetadataToProtoHeader(hdr)
			result.ResponseTrailers = grpcutil.ConvertMetadataToProtoHeader(stream.Trailer())
		}
	}()

	fullDuplex := ccr.StreamType == v1alpha1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM

	var protoErr *v1alpha1.Error
	for _, msg := range ccr.RequestMessages {
		if err := ctx.Err(); err != nil {
			// If an error was returned, convert it to a proto Error
			protoErr = grpcutil.ConvertGrpcToProtoError(err)
			break
		}
		bsr := &v1alpha1.BidiStreamRequest{}
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
			// If the call was successful, get the returned payloads
			result.Payloads = append(result.Payloads, msg.Payload)
		}
	}

	// If we received an error in any of the send logic or full-duplex reads, then exit
	if protoErr != nil {
		result.Error = protoErr
		return result, nil
	}

	// Sends are done, close the send side of the stream
	if err := stream.CloseSend(); err != nil {
		result.Error = grpcutil.ConvertGrpcToProtoError(err)
		return result, nil
	}

	// Receive any remaining responses
	for {
		if err := ctx.Err(); err != nil {
			// If an error was returned, convert it to a proto Error
			protoErr = grpcutil.ConvertGrpcToProtoError(err)
			break
		}
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
		// If the call was successful, save the payloads
		result.Payloads = append(result.Payloads, msg.Payload)
	}

	if protoErr != nil {
		result.Error = protoErr
		return result, nil
	}

	return result, nil
}

// Creates a new invoker around a ConformanceServiceClient.
func newInvoker(clientConn grpc.ClientConnInterface) *invoker {
	client := v1alpha1.NewConformanceServiceClient(clientConn)
	return &invoker{
		client: client,
	}
}
