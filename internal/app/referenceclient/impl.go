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

package referenceclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"connectrpc.com/conformance/internal"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1/conformancev1connect"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
)

const clientName = "connectconformance-referenceclient"

type invoker struct {
	client        conformancev1connect.ConformanceServiceClient
	referenceMode bool
}

// Creates a new invoker around a ConformanceServiceClient.
func newInvoker(transport http.RoundTripper, referenceMode bool, url *url.URL, opts []connect.ClientOption) *invoker {
	opts = append(opts, connect.WithInterceptors(userAgentClientInterceptor{}))
	client := conformancev1connect.NewConformanceServiceClient(
		&http.Client{Transport: transport},
		url.String(),
		opts...,
	)
	return &invoker{
		client:        client,
		referenceMode: referenceMode,
	}
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
		if len(req.RequestMessages) != 1 {
			return nil, errors.New("unary calls must specify exactly one request message")
		}
		resp, err := i.idempotentUnary(ctx, req)
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
		return nil, fmt.Errorf("method name %s does not exist on service %s", req.GetMethod(), req.GetService())
	}
}

func (i *invoker) unary(
	ctx context.Context,
	req *conformancev1.ClientCompatRequest,
) (*conformancev1.ClientResponseResult, error) {
	msg := req.RequestMessages[0]
	ur := &conformancev1.UnaryRequest{}
	if err := msg.UnmarshalTo(ur); err != nil {
		return nil, err
	}

	request := connect.NewRequest(ur)

	// Add the specified request headers to the request
	internal.AddHeaders(req.RequestHeaders, request.Header())

	var protoErr *conformancev1.Error
	var headers []*conformancev1.Header
	var trailers []*conformancev1.Header
	payloads := make([]*conformancev1.ConformancePayload, 0, 1)

	ctx = i.withWireCapture(ctx)

	// Invoke the Unary call
	resp, err := i.client.Unary(ctx, request)

	if err != nil {
		// If an error was returned, first convert it to a Connect error
		// so that we can get the trailers from the Meta property. Then,
		// convert _that_ to a proto Error so we can set it in the response.
		connectErr := internal.ConvertErrorToConnectError(err)
		trailers = internal.ConvertToProtoHeader(connectErr.Meta())
		protoErr = internal.ConvertConnectToProtoError(connectErr)
	} else {
		// If the call was successful, get the headers and trailers
		headers = internal.ConvertToProtoHeader(resp.Header())
		trailers = internal.ConvertToProtoHeader(resp.Trailer())
		// If there's a payload, add that to the response also
		if resp.Msg.Payload != nil {
			payloads = append(payloads, resp.Msg.Payload)
		}
	}

	statusCode, feedback := i.examineWireDetails(ctx)

	return &conformancev1.ClientResponseResult{
		ResponseHeaders:  headers,
		ResponseTrailers: trailers,
		Payloads:         payloads,
		Error:            protoErr,
		HttpStatusCode:   statusCode,
		Feedback:         feedback,
	}, nil
}

// TODO - This should be consolidated with the unary implementation since they are
// mostly the same. See https://github.com/connectrpc/conformance/pull/721/files#r1415699842
// for an example.
func (i *invoker) idempotentUnary(
	ctx context.Context,
	req *conformancev1.ClientCompatRequest,
) (*conformancev1.ClientResponseResult, error) {
	msg := req.RequestMessages[0]
	ur := &conformancev1.IdempotentUnaryRequest{}
	if err := msg.UnmarshalTo(ur); err != nil {
		return nil, err
	}

	request := connect.NewRequest(ur)

	// Add the specified request headers to the request
	internal.AddHeaders(req.RequestHeaders, request.Header())

	var protoErr *conformancev1.Error
	var headers []*conformancev1.Header
	var trailers []*conformancev1.Header
	payloads := make([]*conformancev1.ConformancePayload, 0, 1)

	ctx = i.withWireCapture(ctx)

	// Invoke the Unary call
	resp, err := i.client.IdempotentUnary(ctx, request)

	if err != nil {
		// If an error was returned, first convert it to a Connect error
		// so that we can get the headers from the Meta property. Then,
		// convert _that_ to a proto Error so we can set it in the response.
		connectErr := internal.ConvertErrorToConnectError(err)
		headers = internal.ConvertToProtoHeader(connectErr.Meta())
		protoErr = internal.ConvertConnectToProtoError(connectErr)
	} else {
		// If the call was successful, get the returned payloads
		// and the headers and trailers
		payloads = append(payloads, resp.Msg.Payload)
		headers = internal.ConvertToProtoHeader(resp.Header())
		trailers = internal.ConvertToProtoHeader(resp.Trailer())
	}

	statusCode, feedback := i.examineWireDetails(ctx)

	return &conformancev1.ClientResponseResult{
		ResponseHeaders:  headers,
		ResponseTrailers: trailers,
		Payloads:         payloads,
		Error:            protoErr,
		HttpStatusCode:   statusCode,
		Feedback:         feedback,
	}, nil
}

func (i *invoker) serverStream(
	ctx context.Context,
	req *conformancev1.ClientCompatRequest,
) (result *conformancev1.ClientResponseResult, _ error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	msg := req.RequestMessages[0]
	ssr := &conformancev1.ServerStreamRequest{}
	if err := msg.UnmarshalTo(ssr); err != nil {
		return nil, err
	}

	request := connect.NewRequest(ssr)

	// Add the specified request headers to the request
	internal.AddHeaders(req.RequestHeaders, request.Header())

	result = &conformancev1.ClientResponseResult{}

	ctx = i.withWireCapture(ctx)

	stream, err := i.client.ServerStream(ctx, request)
	if err != nil {
		// If an error was returned, first convert it to a Connect error
		// so that we can get the headers from the Meta property. Then,
		// convert _that_ to a proto Error so we can set it in the response.
		connectErr := internal.ConvertErrorToConnectError(err)
		headers := internal.ConvertToProtoHeader(connectErr.Meta())
		protoErr := internal.ConvertConnectToProtoError(connectErr)
		return &conformancev1.ClientResponseResult{
			ResponseHeaders: headers,
			Error:           protoErr,
		}, nil
	}
	defer func() {
		// Always make sure stream is closed on exit.
		closeErr := stream.Close()
		if err != nil {
			return
		}
		if result.Error == nil && closeErr != nil {
			result.Error = internal.ConvertErrorToProtoError(closeErr)
		}
		if err == nil {
			// Read headers and trailers from the stream
			result.ResponseHeaders = internal.ConvertToProtoHeader(stream.ResponseHeader())
			result.ResponseTrailers = internal.ConvertToProtoHeader(stream.ResponseTrailer())
			result.HttpStatusCode, result.Feedback = i.examineWireDetails(ctx)
		}
	}()

	if ssr.ResponseDefinition != nil {
		result.Payloads = make([]*conformancev1.ConformancePayload, 0, len(ssr.ResponseDefinition.ResponseData))
	}

	timing, err := internal.GetCancelTiming(req.Cancel)
	if err != nil {
		return nil, err
	}
	// If the cancel timing specifies after 0 responses, then cancel before
	// receiving anything
	if timing.AfterNumResponses == 0 {
		cancel()
	}
	totalRcvd := 0
	for stream.Receive() {
		totalRcvd++
		// If the call was successful, get the returned payloads
		// and the headers and trailers
		result.Payloads = append(result.Payloads, stream.Msg().Payload)

		// If AfterNumResponses is specified, it will be a number > 0 here.
		// If it wasn't specified, it will be -1, which means the totalRcvd
		// will never be equal and we won't cancel.
		if totalRcvd == timing.AfterNumResponses {
			cancel()
		}
	}
	if stream.Err() != nil {
		// If an error was returned, convert it to a proto Error
		result.Error = internal.ConvertErrorToProtoError(stream.Err())
	}

	return result, nil
}

func (i *invoker) clientStream(
	ctx context.Context,
	req *conformancev1.ClientCompatRequest,
) (*conformancev1.ClientResponseResult, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ctx = i.withWireCapture(ctx)
	stream := i.client.ClientStream(ctx)
	var numUnsent int
	// Add the specified request headers to the request
	internal.AddHeaders(req.RequestHeaders, stream.RequestHeader())

	for i, msg := range req.RequestMessages {
		csr := &conformancev1.ClientStreamRequest{}
		if err := msg.UnmarshalTo(csr); err != nil {
			return nil, err
		}

		// Sleep for any specified delay
		time.Sleep(time.Duration(req.RequestDelayMs) * time.Millisecond)

		if err := stream.Send(csr); err != nil && errors.Is(err, io.EOF) {
			numUnsent = len(req.RequestMessages) - i
			break
		}
	}

	var protoErr *conformancev1.Error
	var headers []*conformancev1.Header
	var trailers []*conformancev1.Header
	payloads := make([]*conformancev1.ConformancePayload, 0, 1)

	// Cancellation timing
	timing, err := internal.GetCancelTiming(req.Cancel)
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
	resp, err := stream.CloseAndReceive()
	if err != nil {
		// If an error was returned, first convert it to a Connect error
		// so that we can get the trailers from the Meta property. Then,
		// convert _that_ to a proto Error so we can set it in the response.
		connectErr := internal.ConvertErrorToConnectError(err)
		trailers = internal.ConvertToProtoHeader(connectErr.Meta())
		protoErr = internal.ConvertConnectToProtoError(connectErr)
	} else {
		// If the call was successful, get the returned payloads
		// and the headers and trailers
		payloads = append(payloads, resp.Msg.Payload)
		headers = internal.ConvertToProtoHeader(resp.Header())
		trailers = internal.ConvertToProtoHeader(resp.Trailer())
	}

	statusCode, feedback := i.examineWireDetails(ctx)

	return &conformancev1.ClientResponseResult{
		ResponseHeaders:   headers,
		ResponseTrailers:  trailers,
		Payloads:          payloads,
		NumUnsentRequests: int32(numUnsent),
		Error:             protoErr,
		HttpStatusCode:    statusCode,
		Feedback:          feedback,
	}, nil
}

func (i *invoker) bidiStream(
	ctx context.Context,
	req *conformancev1.ClientCompatRequest,
) (result *conformancev1.ClientResponseResult, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	result = &conformancev1.ClientResponseResult{}

	ctx = i.withWireCapture(ctx)

	stream := i.client.BidiStream(ctx)
	defer func() {
		// Always make sure stream is closed on exit.
		closeErr := stream.CloseResponse()
		if err != nil {
			return
		}
		if result.Error == nil && closeErr != nil {
			result.Error = internal.ConvertErrorToProtoError(closeErr)
		}
		if err == nil {
			// Read headers and trailers from the stream
			result.ResponseHeaders = internal.ConvertToProtoHeader(stream.ResponseHeader())
			result.ResponseTrailers = internal.ConvertToProtoHeader(stream.ResponseTrailer())
			result.HttpStatusCode, result.Feedback = i.examineWireDetails(ctx)
		}
	}()

	// Add the specified request headers to the request
	internal.AddHeaders(req.RequestHeaders, stream.RequestHeader())

	fullDuplex := req.StreamType == conformancev1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM

	// Cancellation timing
	timing, err := internal.GetCancelTiming(req.Cancel)
	if err != nil {
		return nil, err
	}

	var protoErr *conformancev1.Error
	totalRcvd := 0
	for i, msg := range req.RequestMessages {
		bsr := &conformancev1.BidiStreamRequest{}
		if err := msg.UnmarshalTo(bsr); err != nil {
			// Return the error and nil result because this is an
			// unmarshalling error unrelated to the RPC
			return nil, err
		}

		// Sleep for any specified delay
		time.Sleep(time.Duration(req.RequestDelayMs) * time.Millisecond)

		if err := stream.Send(bsr); err != nil && errors.Is(err, io.EOF) {
			// Call receive to get the error and convert it to a proto error
			if _, recvErr := stream.Receive(); recvErr != nil {
				protoErr = internal.ConvertErrorToProtoError(recvErr)
			} else {
				// Just in case the receive call doesn't return the error,
				// use the error returned from Send. Note this should never
				// happen, but is here as a safeguard.
				protoErr = internal.ConvertErrorToProtoError(err)
			}
			// Break the send loop
			result.NumUnsentRequests = int32(len(req.RequestMessages) - i)
			break
		}
		if fullDuplex {
			if totalRcvd == timing.AfterNumResponses {
				cancel()
			}
			// If this is a full duplex stream, receive a response for each request
			msg, err := stream.Receive()
			totalRcvd++
			if err != nil {
				if !errors.Is(err, io.EOF) {
					// If an error was returned that is not an EOF, convert it
					// to a proto Error. If the error was an EOF, that just means
					// reads are done.
					protoErr = internal.ConvertErrorToProtoError(err)
				}
				// Reads are done either because we received an error or an EOF
				// In either case, break the outer loop
				break
			}
			// If the call was successful, get the returned payloads
			result.Payloads = append(result.Payloads, msg.Payload)
		}
	}

	if timing.BeforeCloseSend != nil {
		cancel()
	}

	// Sends are done, close the send side of the stream
	if err := stream.CloseRequest(); err != nil {
		return nil, err
	}

	if timing.AfterCloseSendMs >= 0 {
		time.Sleep(time.Duration(timing.AfterCloseSendMs) * time.Millisecond)
		cancel()
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
		msg, err := stream.Receive()
		totalRcvd++
		if err != nil {
			if !errors.Is(err, io.EOF) {
				// If an error was returned that is not an EOF, convert it
				// to a proto Error. If the error was an EOF, that just means
				// reads are done.
				protoErr = internal.ConvertErrorToProtoError(err)
			}
			break
		}
		// If the call was successful, save the payloads
		result.Payloads = append(result.Payloads, msg.Payload)
	}

	if protoErr != nil {
		result.Error = protoErr
	}
	return result, nil
}

func (i *invoker) unimplemented(
	ctx context.Context,
	req *conformancev1.ClientCompatRequest,
) (*conformancev1.ClientResponseResult, error) {
	msg := req.RequestMessages[0]
	ur := &conformancev1.UnimplementedRequest{}
	if err := msg.UnmarshalTo(ur); err != nil {
		return nil, err
	}

	request := connect.NewRequest(ur)
	internal.AddHeaders(req.RequestHeaders, request.Header())

	ctx = i.withWireCapture(ctx)

	// Invoke the Unary call
	_, err := i.client.Unimplemented(ctx, request)

	statusCode, feedback := i.examineWireDetails(ctx)

	return &conformancev1.ClientResponseResult{
		Error:          internal.ConvertErrorToProtoError(err),
		HttpStatusCode: statusCode,
		Feedback:       feedback,
	}, nil
}

func (i *invoker) withWireCapture(ctx context.Context) context.Context {
	if !i.referenceMode {
		return ctx
	}
	return withWireCapture(ctx)
}

func (i *invoker) examineWireDetails(ctx context.Context) (*int32, []string) {
	if !i.referenceMode {
		return nil, nil
	}
	printer := &internal.SimplePrinter{}
	statusCode, ok := examineWireDetails(ctx, printer)
	var statusCodePtr *int32
	if ok {
		statusCodePtr = proto.Int32(int32(statusCode))
	}
	return statusCodePtr, printer.Messages
}

// userAgentClientInterceptor adds to the user-agent header on outgoing requests.
type userAgentClientInterceptor struct{}

func (userAgentClientInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if req.Spec().IsClient {
			// decorate user-agent with the program name and version
			userAgent := fmt.Sprintf("%s %s/%s", req.Header().Get("User-Agent"), clientName, internal.Version)
			req.Header().Set("User-Agent", userAgent)
		}
		return next(ctx, req)
	}
}

func (userAgentClientInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)
		// decorate user-agent with the program name and version
		userAgent := fmt.Sprintf("%s %s/%s", conn.RequestHeader().Get("User-Agent"), clientName, internal.Version)
		conn.RequestHeader().Set("User-Agent", userAgent)
		return conn
	}
}

func (userAgentClientInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}
