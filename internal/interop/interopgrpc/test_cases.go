// Copyright 2022 Buf Technologies, Inc.
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

// This contains the test cases from grpc-go interop test_utils.go file,
// https://github.com/grpc/grpc-go/blob/master/interop/test_utils.go
// The test cases have been refactored to be compatible with the standard
// library *testing.T and other implementations of the testing interface.

/*
 *
 * Copyright 2014 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package interopgrpc

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/bufbuild/connect-crosstest/internal/crosstesting"
	testpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/grpc/testing"
	"github.com/bufbuild/connect-crosstest/internal/interop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const (
	eightBytes          = 8
	sixteenBytes        = 16
	oneKiB              = 1024
	twoKiB              = 2028
	thirtyTwoKiB        = 32768
	sixtyFourKiB        = 65536
	twoFiftyKiB         = 256000
	fiveHundredKiB      = 512000
	largeReqSize        = twoFiftyKiB
	largeRespSize       = fiveHundredKiB
	leadingMetadataKey  = "x-grpc-test-echo-initial"
	trailingMetadataKey = "x-grpc-test-echo-trailing-bin"
)

var (
	reqSizes  = []int{twoFiftyKiB, eightBytes, oneKiB, thirtyTwoKiB}      //nolint:gochecknoglobals // We do want to make this a global so that we can use it in multiple methods
	respSizes = []int{fiveHundredKiB, sixteenBytes, twoKiB, sixtyFourKiB} //nolint:gochecknoglobals // We do want to make this a global so that we can use it in multiple methods
)

// clientNewPayload returns a payload of the given size.
func clientNewPayload(t crosstesting.TB, size int) (*testpb.Payload, error) {
	t.Helper()
	if size < 0 {
		return nil, fmt.Errorf("requested a response with invalid length %d", size)
	}
	body := make([]byte, size)
	return &testpb.Payload{
		Type: testpb.PayloadType_COMPRESSABLE,
		Body: body,
	}, nil
}

// DoEmptyUnaryCall performs a unary RPC with empty request and response messages.
func DoEmptyUnaryCall(t crosstesting.TB, client testpb.TestServiceClient, args ...grpc.CallOption) {
	reply, err := client.EmptyCall(context.Background(), &testpb.Empty{}, args...)
	require.NoError(t, err)
	assert.True(t, proto.Equal(&testpb.Empty{}, reply))
	t.Successf("successful unary call")
}

// DoLargeUnaryCall performs a unary RPC with large payload in the request and response.
func DoLargeUnaryCall(t crosstesting.TB, client testpb.TestServiceClient, args ...grpc.CallOption) {
	pl, err := clientNewPayload(t, largeReqSize)
	require.NoError(t, err)
	req := &testpb.SimpleRequest{
		ResponseType: testpb.PayloadType_COMPRESSABLE,
		ResponseSize: int32(largeRespSize),
		Payload:      pl,
	}
	reply, err := client.UnaryCall(context.Background(), req, args...)
	require.NoError(t, err)
	assert.Equal(t, reply.GetPayload().GetType(), testpb.PayloadType_COMPRESSABLE)
	assert.Equal(t, len(reply.GetPayload().GetBody()), largeRespSize)
	t.Successf("successful large unary call")
}

// DoClientStreaming performs a client streaming RPC.
func DoClientStreaming(t crosstesting.TB, client testpb.TestServiceClient, args ...grpc.CallOption) {
	stream, err := client.StreamingInputCall(context.Background(), args...)
	require.NoError(t, err)
	var sum int
	for _, size := range reqSizes {
		pl, err := clientNewPayload(t, size)
		require.NoError(t, err)
		req := &testpb.StreamingInputCallRequest{
			Payload: pl,
		}
		require.NoError(t, stream.Send(req))
		sum += size
	}
	reply, err := stream.CloseAndRecv()
	require.NoError(t, err)
	assert.Equal(t, reply.GetAggregatedPayloadSize(), int32(sum))
	t.Successf("successful client streaming test")
}

// DoServerStreaming performs a server streaming RPC.
func DoServerStreaming(t crosstesting.TB, client testpb.TestServiceClient, args ...grpc.CallOption) {
	respParam := make([]*testpb.ResponseParameters, len(respSizes))
	for i, s := range respSizes {
		respParam[i] = &testpb.ResponseParameters{
			Size: int32(s),
		}
	}
	req := &testpb.StreamingOutputCallRequest{
		ResponseType:       testpb.PayloadType_COMPRESSABLE,
		ResponseParameters: respParam,
	}
	stream, err := client.StreamingOutputCall(context.Background(), req, args...)
	require.NoError(t, err)
	var rpcStatus error
	var respCnt int
	var index int
	for {
		reply, err := stream.Recv()
		if err != nil {
			rpcStatus = err
			break
		}
		assert.Equal(t, reply.GetPayload().GetType(), testpb.PayloadType_COMPRESSABLE)
		assert.Equal(t, len(reply.GetPayload().GetBody()), respSizes[index])
		index++
		respCnt++
	}
	assert.Equal(t, rpcStatus, io.EOF)
	assert.Equal(t, respCnt, len(respSizes))
	t.Successf("successful server streaming test")
}

// DoPingPong performs ping-pong style bi-directional streaming RPC.
func DoPingPong(t crosstesting.TB, client testpb.TestServiceClient, args ...grpc.CallOption) {
	stream, err := client.FullDuplexCall(context.Background(), args...)
	require.NoError(t, err)
	var index int
	for index < len(reqSizes) {
		respParam := []*testpb.ResponseParameters{
			{
				Size: int32(respSizes[index]),
			},
		}
		pl, err := clientNewPayload(t, reqSizes[index])
		require.NoError(t, err)
		req := &testpb.StreamingOutputCallRequest{
			ResponseType:       testpb.PayloadType_COMPRESSABLE,
			ResponseParameters: respParam,
			Payload:            pl,
		}
		require.NoError(t, stream.Send(req))
		reply, err := stream.Recv()
		require.NoError(t, err)
		assert.Equal(t, reply.GetPayload().GetType(), testpb.PayloadType_COMPRESSABLE)
		assert.Equal(t, len(reply.GetPayload().GetBody()), respSizes[index])
		index++
	}
	require.NoError(t, stream.CloseSend())
	_, err = stream.Recv()
	assert.Equal(t, err, io.EOF)
	t.Successf("successful ping pong")
}

// DoEmptyStream sets up a bi-directional streaming with zero message.
func DoEmptyStream(t crosstesting.TB, client testpb.TestServiceClient, args ...grpc.CallOption) {
	stream, err := client.FullDuplexCall(context.Background(), args...)
	require.NoError(t, err)
	require.NoError(t, stream.CloseSend())
	_, err = stream.Recv()
	assert.Error(t, err)
	assert.Equal(t, err, io.EOF)
	t.Successf("successful empty stream")
}

// DoTimeoutOnSleepingServer performs an RPC on a sleep server which causes RPC timeout.
func DoTimeoutOnSleepingServer(t crosstesting.TB, client testpb.TestServiceClient, args ...grpc.CallOption) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	stream, err := client.FullDuplexCall(ctx, args...)
	if err != nil {
		// This checks if the stream has already timed out before the `Send`, so we would
		// receive a DeadlineExceeded error from the initial call to the bidi streaming RPC.
		if status.Code(err) == codes.DeadlineExceeded {
			return
		}
	}
	require.NoError(t, err)
	pl, err := clientNewPayload(t, 27182)
	require.NoError(t, err)
	req := &testpb.StreamingOutputCallRequest{
		ResponseType: testpb.PayloadType_COMPRESSABLE,
		Payload:      pl,
	}
	err = stream.Send(req)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	_, err = stream.Recv()
	assert.Equal(t, status.Code(err), codes.DeadlineExceeded)
	t.Successf("successful timeout on sleep")
}

var testMetadata = metadata.MD{ //nolint:gochecknoglobals // We do want to make this a global so that we can use it in multiple methods
	"key1": []string{"value1"},
	"key2": []string{"value2"},
}

// DoCancelAfterBegin cancels the RPC after metadata has been sent but before payloads are sent.
func DoCancelAfterBegin(t crosstesting.TB, client testpb.TestServiceClient, args ...grpc.CallOption) {
	ctx, cancel := context.WithCancel(metadata.NewOutgoingContext(context.Background(), testMetadata))
	stream, err := client.StreamingInputCall(ctx, args...)
	require.NoError(t, err)
	cancel()
	_, err = stream.CloseAndRecv()
	assert.Equal(t, status.Code(err), codes.Canceled)
	t.Successf("successful cancel after begin")
}

// DoCancelAfterFirstResponse cancels the RPC after receiving the first message from the server.
func DoCancelAfterFirstResponse(t crosstesting.TB, client testpb.TestServiceClient, args ...grpc.CallOption) {
	ctx, cancel := context.WithCancel(context.Background())
	stream, err := client.FullDuplexCall(ctx, args...)
	require.NoError(t, err)
	respParam := []*testpb.ResponseParameters{
		{
			Size: 31415,
		},
	}
	pl, err := clientNewPayload(t, 27182)
	require.NoError(t, err)
	req := &testpb.StreamingOutputCallRequest{
		ResponseType:       testpb.PayloadType_COMPRESSABLE,
		ResponseParameters: respParam,
		Payload:            pl,
	}
	require.NoError(t, stream.Send(req))
	_, err = stream.Recv()
	require.NoError(t, err)
	cancel()
	_, err = stream.Recv()
	assert.Equal(t, status.Code(err), codes.Canceled)
	t.Successf("successful cancel after first response")
}

const (
	leadingMetadataValue  = "test_initial_metadata_value"
	trailingMetadataValue = "\x0a\x0b\x0a\x0b\x0a\x0b"
)

var (
	customMetadata = metadata.Pairs( //nolint:gochecknoglobals // We do want to make this a global so that we can use it in multiple methods
		leadingMetadataKey, leadingMetadataValue,
		trailingMetadataKey, trailingMetadataValue,
	)
	duplicatedCustomMetadata = metadata.Pairs( //nolint:gochecknoglobals // We do want to make this a global so that we can use it in multiple methods
		leadingMetadataKey, leadingMetadataValue,
		trailingMetadataKey, trailingMetadataValue,
		leadingMetadataKey, leadingMetadataValue+",more_stuff",
		trailingMetadataKey, trailingMetadataValue+"\x0a",
	)
)

func validateMetadata(t crosstesting.TB, header, trailer, sent metadata.MD) {
	expectedHeaderValues := sent.Get(leadingMetadataKey)
	headerValues := header.Get(leadingMetadataKey)
	assert.Equal(t, len(expectedHeaderValues), len(headerValues))
	expectedValuesMap := map[string]struct{}{}
	for _, expected := range expectedHeaderValues {
		expectedValuesMap[expected] = struct{}{}
	}
	for _, headerValue := range headerValues {
		_, ok := expectedValuesMap[headerValue]
		assert.True(t, ok)
	}
	expectedTrailerValues := sent.Get(trailingMetadataKey)
	trailerValues := trailer.Get(trailingMetadataKey)
	assert.Equal(t, len(expectedTrailerValues), len(trailerValues))
	expectedTrailersMap := map[string]struct{}{}
	for _, expected := range expectedTrailerValues {
		expectedTrailersMap[expected] = struct{}{}
	}
	for _, trailerValue := range trailerValues {
		_, ok := expectedTrailersMap[trailerValue]
		assert.True(t, ok)
	}
}

// DoCustomMetadata checks that metadata is echoed back to the client.
func DoCustomMetadata(t crosstesting.TB, client testpb.TestServiceClient, args ...grpc.CallOption) {
	customMetadataTest(t, client, customMetadata, args...)
	t.Successf("successful custom metadata")
}

// DoDuplicateCustomMetadata adds duplicated metadata keys and checks that the metadata is echoed back
// to the client.
func DoDuplicatedCustomMetadata(t crosstesting.TB, client testpb.TestServiceClient, args ...grpc.CallOption) {
	customMetadataTest(t, client, duplicatedCustomMetadata, args...)
	t.Successf("successful duplicated custom metadata")
}

func customMetadataTest(t crosstesting.TB, client testpb.TestServiceClient, customMetadata metadata.MD, args ...grpc.CallOption) {
	// Testing with UnaryCall.
	customMetadataUnaryTest(t, client, customMetadata, args...)

	// Testing with Server Streaming
	customMetadataServerStreamingTest(t, client, customMetadata, args...)

	// Testing with FullDuplex.
	customMetadataFullDuplexTest(t, client, customMetadata, args...)
}

func customMetadataUnaryTest(t crosstesting.TB, client testpb.TestServiceClient, customMetadata metadata.MD, args ...grpc.CallOption) {
	payload, err := clientNewPayload(t, 1)
	require.NoError(t, err)
	req := &testpb.SimpleRequest{
		ResponseType: testpb.PayloadType_COMPRESSABLE,
		ResponseSize: int32(1),
		Payload:      payload,
	}
	ctx := metadata.NewOutgoingContext(context.Background(), customMetadata)
	var header, trailer metadata.MD
	args = append(args, grpc.Header(&header), grpc.Trailer(&trailer))
	reply, err := client.UnaryCall(
		ctx,
		req,
		args...,
	)
	require.NoError(t, err)
	assert.Equal(t, reply.GetPayload().GetType(), testpb.PayloadType_COMPRESSABLE)
	assert.Equal(t, len(reply.GetPayload().GetBody()), 1)
	validateMetadata(t, header, trailer, customMetadata)
}

func customMetadataServerStreamingTest(t crosstesting.TB, client testpb.TestServiceClient, customMetadata metadata.MD, args ...grpc.CallOption) {
	payload, err := clientNewPayload(t, 1)
	require.NoError(t, err)
	ctx := metadata.NewOutgoingContext(context.Background(), customMetadata)
	respParam := []*testpb.ResponseParameters{
		{
			Size: 1,
		},
	}
	req := &testpb.StreamingOutputCallRequest{
		ResponseType:       testpb.PayloadType_COMPRESSABLE,
		ResponseParameters: respParam,
		Payload:            payload,
	}
	stream, err := client.StreamingOutputCall(ctx, req, args...)
	require.NoError(t, err)
	streamHeader, err := stream.Header()
	require.NoError(t, err)
	_, err = stream.Recv()
	require.NoError(t, err)
	require.NoError(t, stream.CloseSend())
	_, err = stream.Recv()
	assert.Equal(t, err, io.EOF)
	streamTrailer := stream.Trailer()
	validateMetadata(t, streamHeader, streamTrailer, customMetadata)
}

func customMetadataFullDuplexTest(t crosstesting.TB, client testpb.TestServiceClient, customMetadata metadata.MD, args ...grpc.CallOption) {
	payload, err := clientNewPayload(t, 1)
	require.NoError(t, err)
	ctx := metadata.NewOutgoingContext(context.Background(), customMetadata)
	stream, err := client.FullDuplexCall(ctx, args...)
	require.NoError(t, err)
	respParam := []*testpb.ResponseParameters{
		{
			Size: 1,
		},
	}
	streamReq := &testpb.StreamingOutputCallRequest{
		ResponseType:       testpb.PayloadType_COMPRESSABLE,
		ResponseParameters: respParam,
		Payload:            payload,
	}
	require.NoError(t, stream.Send(streamReq))
	streamHeader, err := stream.Header()
	require.NoError(t, err)
	_, err = stream.Recv()
	require.NoError(t, err)
	require.NoError(t, stream.CloseSend())
	_, err = stream.Recv()
	assert.Equal(t, err, io.EOF)
	streamTrailer := stream.Trailer()
	validateMetadata(t, streamHeader, streamTrailer, customMetadata)
}

// DoStatusCodeAndMessage checks that the status code is propagated back to the client.
func DoStatusCodeAndMessage(t crosstesting.TB, client testpb.TestServiceClient, args ...grpc.CallOption) {
	code := int32(codes.Unknown)
	msg := "test status message"
	expectedErr := status.Error(codes.Code(code), msg)
	respStatus := &testpb.EchoStatus{
		Code:    code,
		Message: msg,
	}
	// Test UnaryCall.
	req := &testpb.SimpleRequest{
		ResponseStatus: respStatus,
	}
	_, err := client.UnaryCall(context.Background(), req, args...)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), expectedErr.Error())
	// Test FullDuplexCall.
	stream, err := client.FullDuplexCall(context.Background(), args...)
	require.NoError(t, err)
	streamReq := &testpb.StreamingOutputCallRequest{
		ResponseStatus: respStatus,
	}
	require.NoError(t, stream.Send(streamReq))
	require.NoError(t, stream.CloseSend())
	_, err = stream.Recv()
	assert.Equal(t, err.Error(), expectedErr.Error())
	t.Successf("successful status code and message")
}

// DoSpecialStatusMessage verifies Unicode and whitespace is correctly processed
// in status message.
func DoSpecialStatusMessage(t crosstesting.TB, client testpb.TestServiceClient, args ...grpc.CallOption) {
	code := int32(codes.Unknown)
	msg := "\t\ntest with whitespace\r\nand Unicode BMP ☺ and non-BMP 😈\t\n"
	expectedErr := status.Error(codes.Code(code), msg)
	req := &testpb.SimpleRequest{
		ResponseStatus: &testpb.EchoStatus{
			Code:    code,
			Message: msg,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.UnaryCall(ctx, req, args...)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), expectedErr.Error())
	t.Successf("successful special status message")
}

// DoUnimplementedMethod attempts to call an unimplemented method.
func DoUnimplementedMethod(t crosstesting.TB, cc *grpc.ClientConn, args ...grpc.CallOption) {
	var req, reply proto.Message
	err := cc.Invoke(context.Background(), "/grpc.testing.TestService/UnimplementedCall", req, reply, args...)
	assert.Error(t, err)
	assert.Equal(t, status.Code(err), codes.Unimplemented)
	t.Successf("successful unimplemented method")
}

// DoUnimplementedServerStreamingMethod performs a server streaming RPC that is unimplemented.
func DoUnimplementedServerStreamingMethod(t crosstesting.TB, client testpb.TestServiceClient, args ...grpc.CallOption) {
	stream, err := client.UnimplementedStreamingOutputCall(context.Background(), &testpb.Empty{}, args...)
	require.NoError(t, err)
	_, err = stream.Recv()
	assert.Equal(t, status.Code(err), codes.Unimplemented)
	t.Successf("successful unimplemented server streaming method")
}

// DoUnimplementedService attempts to call a method from an unimplemented service.
func DoUnimplementedService(t crosstesting.TB, client testpb.UnimplementedServiceClient, args ...grpc.CallOption) {
	_, err := client.UnimplementedCall(context.Background(), &testpb.Empty{}, args...)
	assert.Equal(t, status.Code(err), codes.Unimplemented)
	t.Successf("successful unimplemented service")
}

// DoUnimplementedServerStreamingService performs a server streaming RPC from an unimplemented service.
func DoUnimplementedServerStreamingService(t crosstesting.TB, client testpb.UnimplementedServiceClient, args ...grpc.CallOption) {
	stream, err := client.UnimplementedStreamingOutputCall(context.Background(), &testpb.Empty{}, args...)
	require.NoError(t, err)
	_, err = stream.Recv()
	assert.Equal(t, status.Code(err), codes.Unimplemented)
	t.Successf("successful unimplemented server streaming service")
}

// DoFailWithNonASCIIError performs a unary RPC that always return a readable non-ASCII error.
func DoFailWithNonASCIIError(t crosstesting.TB, client testpb.TestServiceClient, args ...grpc.CallOption) {
	reply, err := client.FailUnaryCall(
		context.Background(),
		&testpb.SimpleRequest{
			ResponseType: testpb.PayloadType_COMPRESSABLE,
		},
		args...,
	)
	assert.Nil(t, reply)
	assert.Error(t, err)
	s, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, s.Code(), codes.ResourceExhausted)
	assert.Equal(t, s.Message(), interop.NonASCIIErrMsg)
	errStatus, ok := status.FromError(err)
	require.True(t, ok)
	require.Len(t, errStatus.Details(), 1)
	errorDetail, ok := errStatus.Details()[0].(*testpb.ErrorDetail)
	require.True(t, ok)
	assert.True(t, proto.Equal(errorDetail, interop.ErrorDetail))
	t.Successf("successful fail call with non-ASCII error")
}

// DoFailServerStreamingWithNonASCIIError performs a server streaming RPC that always return a readable non-ASCII error.
func DoFailServerStreamingWithNonASCIIError(t crosstesting.TB, client testpb.TestServiceClient, args ...grpc.CallOption) {
	req := &testpb.StreamingOutputCallRequest{
		ResponseType: testpb.PayloadType_COMPRESSABLE,
	}
	stream, err := client.FailStreamingOutputCall(context.Background(), req, args...)
	require.NoError(t, err)
	reply, err := stream.Recv()
	assert.Nil(t, reply)
	assert.Error(t, err)
	s, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, s.Code(), codes.ResourceExhausted)
	assert.Equal(t, s.Message(), interop.NonASCIIErrMsg)
	errStatus, ok := status.FromError(err)
	require.True(t, ok)
	require.Len(t, errStatus.Details(), 1)
	errorDetail, ok := errStatus.Details()[0].(*testpb.ErrorDetail)
	require.True(t, ok)
	assert.True(t, proto.Equal(errorDetail, interop.ErrorDetail))
	t.Successf("successful fail server streaming with non-ASCII error")
}

// DoFailServerStreamingAfterResponse performs a server streaming RPC that fails after responses have been sent from the server.
func DoFailServerStreamingAfterResponse(t crosstesting.TB, client testpb.TestServiceClient, args ...grpc.CallOption) {
	respParam := make([]*testpb.ResponseParameters, len(respSizes))
	for i, s := range respSizes {
		respParam[i] = &testpb.ResponseParameters{
			Size: int32(s),
		}
	}
	req := &testpb.StreamingOutputCallRequest{
		ResponseType:       testpb.PayloadType_COMPRESSABLE,
		ResponseParameters: respParam,
	}
	stream, err := client.FailStreamingOutputCall(context.Background(), req, args...)
	require.NoError(t, err)
	for i := 0; i < 4; i++ {
		reply, err := stream.Recv()
		require.NoError(t, err)
		require.NotNil(t, reply)
	}
	require.NoError(t, err)
	reply, err := stream.Recv()
	assert.Nil(t, reply)
	assert.Error(t, err)
	s, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, s.Code(), codes.ResourceExhausted)
	assert.Equal(t, s.Message(), interop.NonASCIIErrMsg)
	errStatus, ok := status.FromError(err)
	require.True(t, ok)
	require.Len(t, errStatus.Details(), 1)
	errorDetail, ok := errStatus.Details()[0].(*testpb.ErrorDetail)
	require.True(t, ok)
	assert.True(t, proto.Equal(errorDetail, interop.ErrorDetail))
	t.Successf("successful fail server streaming after response")
}

// DoUnresolvableHost attempts to call a method to an unresolvable host.
func DoUnresolvableHost(t crosstesting.TB, client testpb.TestServiceClient, args ...grpc.CallOption) {
	reply, err := client.EmptyCall(context.Background(), &testpb.Empty{}, args...)
	assert.Nil(t, reply)
	assert.Error(t, err)
	assert.Equal(t, status.Code(err), codes.Unavailable)
	t.Successf("successful fail call with unresolvable call")
}
