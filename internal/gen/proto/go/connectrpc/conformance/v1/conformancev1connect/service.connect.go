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

// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: connectrpc/conformance/v1/service.proto

package conformancev1connect

import (
	v1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_7_0

const (
	// ConformanceServiceName is the fully-qualified name of the ConformanceService service.
	ConformanceServiceName = "connectrpc.conformance.v1.ConformanceService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ConformanceServiceUnaryProcedure is the fully-qualified name of the ConformanceService's Unary
	// RPC.
	ConformanceServiceUnaryProcedure = "/connectrpc.conformance.v1.ConformanceService/Unary"
	// ConformanceServiceServerStreamProcedure is the fully-qualified name of the ConformanceService's
	// ServerStream RPC.
	ConformanceServiceServerStreamProcedure = "/connectrpc.conformance.v1.ConformanceService/ServerStream"
	// ConformanceServiceClientStreamProcedure is the fully-qualified name of the ConformanceService's
	// ClientStream RPC.
	ConformanceServiceClientStreamProcedure = "/connectrpc.conformance.v1.ConformanceService/ClientStream"
	// ConformanceServiceBidiStreamProcedure is the fully-qualified name of the ConformanceService's
	// BidiStream RPC.
	ConformanceServiceBidiStreamProcedure = "/connectrpc.conformance.v1.ConformanceService/BidiStream"
	// ConformanceServiceUnimplementedProcedure is the fully-qualified name of the ConformanceService's
	// Unimplemented RPC.
	ConformanceServiceUnimplementedProcedure = "/connectrpc.conformance.v1.ConformanceService/Unimplemented"
	// ConformanceServiceIdempotentUnaryProcedure is the fully-qualified name of the
	// ConformanceService's IdempotentUnary RPC.
	ConformanceServiceIdempotentUnaryProcedure = "/connectrpc.conformance.v1.ConformanceService/IdempotentUnary"
)

// ConformanceServiceClient is a client for the connectrpc.conformance.v1.ConformanceService
// service.
type ConformanceServiceClient interface {
	// If the response_delay_ms duration is specified, the server should wait the
	// given duration after reading the request before sending the corresponding
	// response.
	//
	// Servers should allow the response definition to be unset in the request and
	// if it is, set no response headers or trailers and return no response data.
	// The returned payload should only contain the request info.
	Unary(context.Context, *connect.Request[v1.UnaryRequest]) (*connect.Response[v1.UnaryResponse], error)
	// A server-streaming operation. The request indicates the response headers,
	// response messages, trailers, and an optional error to send back. The
	// response data should be sent in the order indicated, and the server should
	// wait between sending response messages as indicated.
	//
	// Response message data is specified as bytes. The service should echo back
	// request properties in the first ConformancePayload, and then include the
	// message data in the data field. Subsequent messages after the first one
	// should contain only the data field.
	//
	// If a response definition is not specified OR is specified, but response data
	// is empty, the server should skip sending anything on the stream. When there
	// are no responses to send, servers should throw an error if one is provided
	// and return without error if one is not. Stream headers and trailers should
	// still be set on the stream if provided regardless of whether a response is
	// sent or an error is thrown.
	ServerStream(context.Context, *connect.Request[v1.ServerStreamRequest]) (*connect.ServerStreamForClient[v1.ServerStreamResponse], error)
	// A client-streaming operation. The first request indicates the response
	// headers and trailers and also indicates either a response message or an
	// error to send back.
	//
	// Response message data is specified as bytes. The service should echo back
	// request properties, including all request messages in the order they were
	// received, in the ConformancePayload and then include the message data in
	// the data field.
	//
	// If the input stream is empty, the server's response will include no data,
	// only the request properties (headers, timeout).
	//
	// Servers should only read the response definition from the first message in
	// the stream and should ignore any definition set in subsequent messages.
	//
	// Servers should allow the response definition to be unset in the request and
	// if it is, set no response headers or trailers and return no response data.
	// The returned payload should only contain the request info.
	ClientStream(context.Context) *connect.ClientStreamForClient[v1.ClientStreamRequest, v1.ClientStreamResponse]
	// A bidirectional-streaming operation. The first request indicates the response
	// headers, response messages, trailers, and an optional error to send back.
	// The response data should be sent in the order indicated, and the server
	// should wait between sending response messages as indicated.
	//
	// Response message data is specified as bytes and should be included in the
	// data field of the ConformancePayload in each response.
	//
	// Servers should send responses indicated according to the rules of half duplex
	// vs. full duplex streams. Once all responses are sent, the server should either
	// return an error if specified or close the stream without error.
	//
	// If a response definition is not specified OR is specified, but response data
	// is empty, the server should skip sending anything on the stream. Stream
	// headers and trailers should always be set on the stream if provided
	// regardless of whether a response is sent or an error is thrown.
	//
	// If the full_duplex field is true:
	//   - the handler should read one request and then send back one response, and
	//     then alternate, reading another request and then sending back another response, etc.
	//
	//   - if the server receives a request and has no responses to send, it
	//     should throw the error specified in the request.
	//
	//   - the service should echo back all request properties in the first response
	//     including the last received request. Subsequent responses should only
	//     echo back the last received request.
	//
	//   - if the response_delay_ms duration is specified, the server should wait the given
	//     duration after reading the request before sending the corresponding
	//     response.
	//
	// If the full_duplex field is false:
	//   - the handler should read all requests until the client is done sending.
	//     Once all requests are read, the server should then send back any responses
	//     specified in the response definition.
	//
	//   - the server should echo back all request properties, including all request
	//     messages in the order they were received, in the first response. Subsequent
	//     responses should only include the message data in the data field.
	//
	//   - if the response_delay_ms duration is specified, the server should wait that
	//     long in between sending each response message.
	BidiStream(context.Context) *connect.BidiStreamForClient[v1.BidiStreamRequest, v1.BidiStreamResponse]
	// A unary endpoint that the server should not implement and should instead
	// return an unimplemented error when invoked.
	Unimplemented(context.Context, *connect.Request[v1.UnimplementedRequest]) (*connect.Response[v1.UnimplementedResponse], error)
	// A unary endpoint denoted as having no side effects (i.e. idempotent).
	// Implementations should use an HTTP GET when invoking this endpoint and
	// leverage query parameters to send data.
	IdempotentUnary(context.Context, *connect.Request[v1.IdempotentUnaryRequest]) (*connect.Response[v1.IdempotentUnaryResponse], error)
}

// NewConformanceServiceClient constructs a client for the
// connectrpc.conformance.v1.ConformanceService service. By default, it uses the Connect protocol
// with the binary Protobuf Codec, asks for gzipped responses, and sends uncompressed requests. To
// use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or connect.WithGRPCWeb()
// options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewConformanceServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) ConformanceServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &conformanceServiceClient{
		unary: connect.NewClient[v1.UnaryRequest, v1.UnaryResponse](
			httpClient,
			baseURL+ConformanceServiceUnaryProcedure,
			opts...,
		),
		serverStream: connect.NewClient[v1.ServerStreamRequest, v1.ServerStreamResponse](
			httpClient,
			baseURL+ConformanceServiceServerStreamProcedure,
			opts...,
		),
		clientStream: connect.NewClient[v1.ClientStreamRequest, v1.ClientStreamResponse](
			httpClient,
			baseURL+ConformanceServiceClientStreamProcedure,
			opts...,
		),
		bidiStream: connect.NewClient[v1.BidiStreamRequest, v1.BidiStreamResponse](
			httpClient,
			baseURL+ConformanceServiceBidiStreamProcedure,
			opts...,
		),
		unimplemented: connect.NewClient[v1.UnimplementedRequest, v1.UnimplementedResponse](
			httpClient,
			baseURL+ConformanceServiceUnimplementedProcedure,
			opts...,
		),
		idempotentUnary: connect.NewClient[v1.IdempotentUnaryRequest, v1.IdempotentUnaryResponse](
			httpClient,
			baseURL+ConformanceServiceIdempotentUnaryProcedure,
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
		),
	}
}

// conformanceServiceClient implements ConformanceServiceClient.
type conformanceServiceClient struct {
	unary           *connect.Client[v1.UnaryRequest, v1.UnaryResponse]
	serverStream    *connect.Client[v1.ServerStreamRequest, v1.ServerStreamResponse]
	clientStream    *connect.Client[v1.ClientStreamRequest, v1.ClientStreamResponse]
	bidiStream      *connect.Client[v1.BidiStreamRequest, v1.BidiStreamResponse]
	unimplemented   *connect.Client[v1.UnimplementedRequest, v1.UnimplementedResponse]
	idempotentUnary *connect.Client[v1.IdempotentUnaryRequest, v1.IdempotentUnaryResponse]
}

// Unary calls connectrpc.conformance.v1.ConformanceService.Unary.
func (c *conformanceServiceClient) Unary(ctx context.Context, req *connect.Request[v1.UnaryRequest]) (*connect.Response[v1.UnaryResponse], error) {
	return c.unary.CallUnary(ctx, req)
}

// ServerStream calls connectrpc.conformance.v1.ConformanceService.ServerStream.
func (c *conformanceServiceClient) ServerStream(ctx context.Context, req *connect.Request[v1.ServerStreamRequest]) (*connect.ServerStreamForClient[v1.ServerStreamResponse], error) {
	return c.serverStream.CallServerStream(ctx, req)
}

// ClientStream calls connectrpc.conformance.v1.ConformanceService.ClientStream.
func (c *conformanceServiceClient) ClientStream(ctx context.Context) *connect.ClientStreamForClient[v1.ClientStreamRequest, v1.ClientStreamResponse] {
	return c.clientStream.CallClientStream(ctx)
}

// BidiStream calls connectrpc.conformance.v1.ConformanceService.BidiStream.
func (c *conformanceServiceClient) BidiStream(ctx context.Context) *connect.BidiStreamForClient[v1.BidiStreamRequest, v1.BidiStreamResponse] {
	return c.bidiStream.CallBidiStream(ctx)
}

// Unimplemented calls connectrpc.conformance.v1.ConformanceService.Unimplemented.
func (c *conformanceServiceClient) Unimplemented(ctx context.Context, req *connect.Request[v1.UnimplementedRequest]) (*connect.Response[v1.UnimplementedResponse], error) {
	return c.unimplemented.CallUnary(ctx, req)
}

// IdempotentUnary calls connectrpc.conformance.v1.ConformanceService.IdempotentUnary.
func (c *conformanceServiceClient) IdempotentUnary(ctx context.Context, req *connect.Request[v1.IdempotentUnaryRequest]) (*connect.Response[v1.IdempotentUnaryResponse], error) {
	return c.idempotentUnary.CallUnary(ctx, req)
}

// ConformanceServiceHandler is an implementation of the
// connectrpc.conformance.v1.ConformanceService service.
type ConformanceServiceHandler interface {
	// If the response_delay_ms duration is specified, the server should wait the
	// given duration after reading the request before sending the corresponding
	// response.
	//
	// Servers should allow the response definition to be unset in the request and
	// if it is, set no response headers or trailers and return no response data.
	// The returned payload should only contain the request info.
	Unary(context.Context, *connect.Request[v1.UnaryRequest]) (*connect.Response[v1.UnaryResponse], error)
	// A server-streaming operation. The request indicates the response headers,
	// response messages, trailers, and an optional error to send back. The
	// response data should be sent in the order indicated, and the server should
	// wait between sending response messages as indicated.
	//
	// Response message data is specified as bytes. The service should echo back
	// request properties in the first ConformancePayload, and then include the
	// message data in the data field. Subsequent messages after the first one
	// should contain only the data field.
	//
	// If a response definition is not specified OR is specified, but response data
	// is empty, the server should skip sending anything on the stream. When there
	// are no responses to send, servers should throw an error if one is provided
	// and return without error if one is not. Stream headers and trailers should
	// still be set on the stream if provided regardless of whether a response is
	// sent or an error is thrown.
	ServerStream(context.Context, *connect.Request[v1.ServerStreamRequest], *connect.ServerStream[v1.ServerStreamResponse]) error
	// A client-streaming operation. The first request indicates the response
	// headers and trailers and also indicates either a response message or an
	// error to send back.
	//
	// Response message data is specified as bytes. The service should echo back
	// request properties, including all request messages in the order they were
	// received, in the ConformancePayload and then include the message data in
	// the data field.
	//
	// If the input stream is empty, the server's response will include no data,
	// only the request properties (headers, timeout).
	//
	// Servers should only read the response definition from the first message in
	// the stream and should ignore any definition set in subsequent messages.
	//
	// Servers should allow the response definition to be unset in the request and
	// if it is, set no response headers or trailers and return no response data.
	// The returned payload should only contain the request info.
	ClientStream(context.Context, *connect.ClientStream[v1.ClientStreamRequest]) (*connect.Response[v1.ClientStreamResponse], error)
	// A bidirectional-streaming operation. The first request indicates the response
	// headers, response messages, trailers, and an optional error to send back.
	// The response data should be sent in the order indicated, and the server
	// should wait between sending response messages as indicated.
	//
	// Response message data is specified as bytes and should be included in the
	// data field of the ConformancePayload in each response.
	//
	// Servers should send responses indicated according to the rules of half duplex
	// vs. full duplex streams. Once all responses are sent, the server should either
	// return an error if specified or close the stream without error.
	//
	// If a response definition is not specified OR is specified, but response data
	// is empty, the server should skip sending anything on the stream. Stream
	// headers and trailers should always be set on the stream if provided
	// regardless of whether a response is sent or an error is thrown.
	//
	// If the full_duplex field is true:
	//   - the handler should read one request and then send back one response, and
	//     then alternate, reading another request and then sending back another response, etc.
	//
	//   - if the server receives a request and has no responses to send, it
	//     should throw the error specified in the request.
	//
	//   - the service should echo back all request properties in the first response
	//     including the last received request. Subsequent responses should only
	//     echo back the last received request.
	//
	//   - if the response_delay_ms duration is specified, the server should wait the given
	//     duration after reading the request before sending the corresponding
	//     response.
	//
	// If the full_duplex field is false:
	//   - the handler should read all requests until the client is done sending.
	//     Once all requests are read, the server should then send back any responses
	//     specified in the response definition.
	//
	//   - the server should echo back all request properties, including all request
	//     messages in the order they were received, in the first response. Subsequent
	//     responses should only include the message data in the data field.
	//
	//   - if the response_delay_ms duration is specified, the server should wait that
	//     long in between sending each response message.
	BidiStream(context.Context, *connect.BidiStream[v1.BidiStreamRequest, v1.BidiStreamResponse]) error
	// A unary endpoint that the server should not implement and should instead
	// return an unimplemented error when invoked.
	Unimplemented(context.Context, *connect.Request[v1.UnimplementedRequest]) (*connect.Response[v1.UnimplementedResponse], error)
	// A unary endpoint denoted as having no side effects (i.e. idempotent).
	// Implementations should use an HTTP GET when invoking this endpoint and
	// leverage query parameters to send data.
	IdempotentUnary(context.Context, *connect.Request[v1.IdempotentUnaryRequest]) (*connect.Response[v1.IdempotentUnaryResponse], error)
}

// NewConformanceServiceHandler builds an HTTP handler from the service implementation. It returns
// the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewConformanceServiceHandler(svc ConformanceServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	conformanceServiceUnaryHandler := connect.NewUnaryHandler(
		ConformanceServiceUnaryProcedure,
		svc.Unary,
		opts...,
	)
	conformanceServiceServerStreamHandler := connect.NewServerStreamHandler(
		ConformanceServiceServerStreamProcedure,
		svc.ServerStream,
		opts...,
	)
	conformanceServiceClientStreamHandler := connect.NewClientStreamHandler(
		ConformanceServiceClientStreamProcedure,
		svc.ClientStream,
		opts...,
	)
	conformanceServiceBidiStreamHandler := connect.NewBidiStreamHandler(
		ConformanceServiceBidiStreamProcedure,
		svc.BidiStream,
		opts...,
	)
	conformanceServiceUnimplementedHandler := connect.NewUnaryHandler(
		ConformanceServiceUnimplementedProcedure,
		svc.Unimplemented,
		opts...,
	)
	conformanceServiceIdempotentUnaryHandler := connect.NewUnaryHandler(
		ConformanceServiceIdempotentUnaryProcedure,
		svc.IdempotentUnary,
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
	)
	return "/connectrpc.conformance.v1.ConformanceService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ConformanceServiceUnaryProcedure:
			conformanceServiceUnaryHandler.ServeHTTP(w, r)
		case ConformanceServiceServerStreamProcedure:
			conformanceServiceServerStreamHandler.ServeHTTP(w, r)
		case ConformanceServiceClientStreamProcedure:
			conformanceServiceClientStreamHandler.ServeHTTP(w, r)
		case ConformanceServiceBidiStreamProcedure:
			conformanceServiceBidiStreamHandler.ServeHTTP(w, r)
		case ConformanceServiceUnimplementedProcedure:
			conformanceServiceUnimplementedHandler.ServeHTTP(w, r)
		case ConformanceServiceIdempotentUnaryProcedure:
			conformanceServiceIdempotentUnaryHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedConformanceServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedConformanceServiceHandler struct{}

func (UnimplementedConformanceServiceHandler) Unary(context.Context, *connect.Request[v1.UnaryRequest]) (*connect.Response[v1.UnaryResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("connectrpc.conformance.v1.ConformanceService.Unary is not implemented"))
}

func (UnimplementedConformanceServiceHandler) ServerStream(context.Context, *connect.Request[v1.ServerStreamRequest], *connect.ServerStream[v1.ServerStreamResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("connectrpc.conformance.v1.ConformanceService.ServerStream is not implemented"))
}

func (UnimplementedConformanceServiceHandler) ClientStream(context.Context, *connect.ClientStream[v1.ClientStreamRequest]) (*connect.Response[v1.ClientStreamResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("connectrpc.conformance.v1.ConformanceService.ClientStream is not implemented"))
}

func (UnimplementedConformanceServiceHandler) BidiStream(context.Context, *connect.BidiStream[v1.BidiStreamRequest, v1.BidiStreamResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("connectrpc.conformance.v1.ConformanceService.BidiStream is not implemented"))
}

func (UnimplementedConformanceServiceHandler) Unimplemented(context.Context, *connect.Request[v1.UnimplementedRequest]) (*connect.Response[v1.UnimplementedResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("connectrpc.conformance.v1.ConformanceService.Unimplemented is not implemented"))
}

func (UnimplementedConformanceServiceHandler) IdempotentUnary(context.Context, *connect.Request[v1.IdempotentUnaryRequest]) (*connect.Response[v1.IdempotentUnaryResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("connectrpc.conformance.v1.ConformanceService.IdempotentUnary is not implemented"))
}
