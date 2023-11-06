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
	_ context.Context,
	_ *v1alpha1.ClientCompatRequest,
) (*v1alpha1.ClientResponseResult, error) {
	return nil, errors.New("server streaming is not yet implemented")
}

func (i *invoker) clientStream(
	_ context.Context,
	_ *v1alpha1.ClientCompatRequest,
) (*v1alpha1.ClientResponseResult, error) {
	return nil, errors.New("client streaming is not yet implemented")
}

func (i *invoker) bidiStream(
	_ context.Context,
	_ *v1alpha1.ClientCompatRequest,
) (result *v1alpha1.ClientResponseResult, retErr error) {
	return nil, errors.New("bidi streaming is not yet implemented")
}

// Creates a new invoker around a ConformanceServiceClient.
func newInvoker(clientConn grpc.ClientConnInterface) *invoker {
	client := v1alpha1.NewConformanceServiceClient(clientConn)
	return &invoker{
		client: client,
	}
}
