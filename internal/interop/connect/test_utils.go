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

package interopconnect

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc/benchmark/stats"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	"github.com/bufbuild/connect"
	connectpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/connect/grpc/testing/testingconnect"
	testpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/grpc/testing"
)

var (
	reqSizes            = []int{27182, 8, 1828, 45904}
	respSizes           = []int{31415, 9, 2653, 58979}
	largeReqSize        = 271828
	largeRespSize       = 314159
	initialMetadataKey  = "x-grpc-test-echo-initial"
	trailingMetadataKey = "x-grpc-test-echo-trailing-bin"
)

// ClientNewPayload returns a payload of the given type and size.
func ClientNewPayload(t testpb.PayloadType, size int) *testpb.Payload {
	if size < 0 {
		log.Fatalf("Requested a response with invalid length %d", size)
	}
	body := make([]byte, size)
	switch t {
	case testpb.PayloadType_COMPRESSABLE:
	default:
		log.Fatalf("Unsupported payload type: %d", t)
	}
	return &testpb.Payload{
		Type: t,
		Body: body,
	}
}

// DoEmptyUnaryCall performs a unary RPC with empty request and response messages.
func DoEmptyUnaryCall(tc connectpb.TestServiceClient) {
	reply, err := tc.EmptyCall(
		context.Background(),
		connect.NewRequest(&testpb.Empty{}),
	)
	if err != nil {
		log.Fatal("/TestService/EmptyCall RPC failed: ", err)
	}
	if !proto.Equal(&testpb.Empty{}, reply.Msg) {
		log.Fatalf("/TestService/EmptyCall receives %v, want %v", reply, testpb.Empty{})
	}
	fmt.Println("succcessful unary call")
}

// DoLargeUnaryCall performs a unary RPC with large payload in the request and response.
func DoLargeUnaryCall(tc connectpb.TestServiceClient) {
	pl := ClientNewPayload(testpb.PayloadType_COMPRESSABLE, largeReqSize)
	req := &testpb.SimpleRequest{
		ResponseType: testpb.PayloadType_COMPRESSABLE,
		ResponseSize: int32(largeRespSize),
		Payload:      pl,
	}
	reply, err := tc.UnaryCall(context.Background(), connect.NewRequest(req))
	if err != nil {
		log.Fatal("/TestService/UnaryCall RPC failed: ", err)
	}
	t := reply.Msg.GetPayload().GetType()
	s := len(reply.Msg.GetPayload().GetBody())
	if t != testpb.PayloadType_COMPRESSABLE || s != largeRespSize {
		log.Fatalf("Got the reply with type %d len %d; want %d, %d", t, s, testpb.PayloadType_COMPRESSABLE, largeRespSize)
	}
	fmt.Println("successful large unary call")
}

// DoClientStreaming performs a client streaming RPC.
func DoClientStreaming(tc connectpb.TestServiceClient) {
	stream := tc.StreamingInputCall(context.Background())
	var sum int
	for _, s := range reqSizes {
		pl := ClientNewPayload(testpb.PayloadType_COMPRESSABLE, s)
		req := &testpb.StreamingInputCallRequest{
			Payload: pl,
		}
		if err := stream.Send(req); err != nil {
			log.Fatalf("%v has error %v while sending %v", stream, err, req)
		}
		sum += s
	}
	reply, err := stream.CloseAndReceive()
	if err != nil {
		log.Fatalf("%v.CloseAndReceive() got error %v, want %v", stream, err, nil)
	}
	if reply.Msg.GetAggregatedPayloadSize() != int32(sum) {
		log.Fatalf("%v.CloseAndReceive().GetAggregatePayloadSize() = %v; want %v", stream, reply.Msg.GetAggregatedPayloadSize(), sum)
	}
	fmt.Println("successful client streaming test")
}

// DoServerStreaming performs a server streaming RPC.
func DoServerStreaming(tc connectpb.TestServiceClient) {
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
	stream, err := tc.StreamingOutputCall(context.Background(), connect.NewRequest(req))
	if err != nil {
		log.Fatalf("%v.StreamingOutputCall(_) = _, %v", tc, err)
	}
	var respCnt int
	var index int
	for stream.Receive() {
		t := stream.Msg().GetPayload().GetType()
		if t != testpb.PayloadType_COMPRESSABLE {
			log.Fatalf("Got the reply of type %d, want %d", t, testpb.PayloadType_COMPRESSABLE)
		}
		size := len(stream.Msg().GetPayload().GetBody())
		if size != respSizes[index] {
			log.Fatalf("Got reply body of length %d, want %d", size, respSizes[index])
		}
		index++
		respCnt++
	}
	if stream.Err() != nil && !errors.Is(stream.Err(), io.EOF) {
		log.Fatalf("Failed to finish the server streaming rpc: %v", stream.Err())
	}
	if respCnt != len(respSizes) {
		log.Fatalf("Got %d reply, want %d", len(respSizes), respCnt)
	}
	fmt.Println("successful server streaming test")
}

// DoPingPong performs ping-pong style bi-directional streaming RPC.
func DoPingPong(tc connectpb.TestServiceClient) {
	stream := tc.FullDuplexCall(context.Background())
	if stream == nil {
		log.Fatalf("%v.FullDuplexCall(_) = <nil>", tc)
	}
	var index int
	for index < len(reqSizes) {
		respParam := []*testpb.ResponseParameters{
			{
				Size: int32(respSizes[index]),
			},
		}
		pl := ClientNewPayload(testpb.PayloadType_COMPRESSABLE, reqSizes[index])
		req := &testpb.StreamingOutputCallRequest{
			ResponseType:       testpb.PayloadType_COMPRESSABLE,
			ResponseParameters: respParam,
			Payload:            pl,
		}
		if err := stream.Send(req); err != nil {
			log.Fatalf("%v has error %v while sending %v", stream, err, req)
		}
		reply, err := stream.Receive()
		if err != nil {
			log.Fatalf("%v.Receive() = %v", stream, err)
		}
		t := reply.GetPayload().GetType()
		if t != testpb.PayloadType_COMPRESSABLE {
			log.Fatalf("Got the reply of type %d, want %d", t, testpb.PayloadType_COMPRESSABLE)
		}
		size := len(reply.GetPayload().GetBody())
		if size != respSizes[index] {
			log.Fatalf("Got reply body of length %d, want %d", size, respSizes[index])
		}
		index++
	}
	if err := stream.CloseSend(); err != nil {
		log.Fatalf("%v.CloseSend() got %v, want %v", stream, err, nil)
	}
	if _, err := stream.Receive(); !errors.Is(err, io.EOF) {
		log.Fatalf("%v failed to complete the ping pong test: %v", stream, err)
	}
	fmt.Println("successful ping pong")
}

// DoEmptyStream sets up a bi-directional streaming with zero message.
func DoEmptyStream(tc connectpb.TestServiceClient) {
	stream := tc.FullDuplexCall(context.Background())
	if stream == nil {
		log.Fatalf("%v.FullDuplexCall(_) = <nil>", tc)
	}
	if err := stream.CloseSend(); err != nil {
		log.Fatalf("%v.CloseSend() got %v, want %v", stream, err, nil)
	}
	if _, err := stream.Receive(); err != nil && !errors.Is(err, io.EOF) {
		log.Fatalf("%v failed to complete the empty stream test: %v", stream, err)
	}
	fmt.Println("successful empty stream")
}

// DoTimeoutOnSleepingServer performs an RPC on a sleep server which causes RPC timeout.
func DoTimeoutOnSleepingServer(tc connectpb.TestServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	stream := tc.FullDuplexCall(ctx)
	if stream == nil {
		log.Fatalf("%v.FullDuplexCall(_) = <nil>", tc)
	}
	pl := ClientNewPayload(testpb.PayloadType_COMPRESSABLE, 27182)
	req := &testpb.StreamingOutputCallRequest{
		ResponseType: testpb.PayloadType_COMPRESSABLE,
		Payload:      pl,
	}
	if err := stream.Send(req); err != nil && !errors.Is(err, io.EOF) {
		log.Fatalf("%v.Send(_) = %v", stream, err)
	}
	if _, err := stream.Receive(); err != nil {
		if connect.CodeOf(err) != connect.CodeDeadlineExceeded {
			log.Fatalf("%v.Receive() = _, %v, want error code %d", stream, err, connect.CodeDeadlineExceeded)
		}
	}
	fmt.Println("successful timeout on sleep")
}

var testMetadata = metadata.MD{
	"key1": []string{"value1"},
	"key2": []string{"value2"},
}

// DoCancelAfterBegin cancels the RPC after metadata has been sent but before payloads are sent.
func DoCancelAfterBegin(tc connectpb.TestServiceClient) {
	// TODO(doria): don't use grpc metadata library here...?
	ctx, cancel := context.WithCancel(metadata.NewOutgoingContext(context.Background(), testMetadata))
	stream := tc.StreamingInputCall(ctx)
	if stream == nil {
		log.Fatalf("%v.StreamingInputCall(_) = _, <nil>", tc)
	}
	cancel()
	_, err := stream.CloseAndReceive()
	if connect.CodeOf(err) != connect.CodeCanceled {
		log.Fatalf("%v.CloseAndReceive() got error %v, want %d", stream, err, connect.CodeCanceled)
	}
	fmt.Println("successful cancel after begin")
}

// DoCancelAfterFirstResponse cancels the RPC after receiving the first message from the server.
func DoCancelAfterFirstResponse(tc connectpb.TestServiceClient) {
	ctx, cancel := context.WithCancel(context.Background())
	stream := tc.FullDuplexCall(ctx)
	if stream == nil {
		log.Fatalf("%v.FullDuplexCall(_) = <nil>", tc)
	}
	respParam := []*testpb.ResponseParameters{
		{
			Size: 31415,
		},
	}
	pl := ClientNewPayload(testpb.PayloadType_COMPRESSABLE, 27182)
	req := &testpb.StreamingOutputCallRequest{
		ResponseType:       testpb.PayloadType_COMPRESSABLE,
		ResponseParameters: respParam,
		Payload:            pl,
	}
	if err := stream.Send(req); err != nil {
		log.Fatalf("%v has error %v while sending %v", stream, err, req)
	}
	if _, err := stream.Receive(); err != nil {
		log.Fatalf("%v.Recv() = %v", stream, err)
	}
	cancel()
	_, err := stream.Receive()
	if connect.CodeOf(err) != connect.CodeCanceled {
		log.Fatalf("%v completed with error %v, want %d", stream, err, connect.CodeCanceled)
	}
	fmt.Println("successful cancel after first response")
}

var (
	initialMetadataValue  = "test_initial_metadata_value"
	trailingMetadataValue = []byte("\x0a\x0b\x0a\x0b\x0a\x0b")
)

func validateMetadata(header, trailer http.Header) {
	if len(header.Values(initialMetadataKey)) != 1 {
		log.Fatalf("Expected exactly one header from server. Received %d", len(header.Values(initialMetadataKey)))
	}
	if header.Get(initialMetadataKey) != initialMetadataValue {
		log.Fatalf("Got header %s; want %s", header.Get(initialMetadataKey), initialMetadataValue)
	}
	if len(trailer.Values(trailingMetadataKey)) != 1 {
		log.Fatalf("Expected exactly one trailer from server. Received %d", len(trailer.Values(trailingMetadataKey)))
	}
	decodedTrailer, err := connect.DecodeBinaryHeader(trailer.Get(trailingMetadataKey))
	if err != nil {
		log.Fatalf("Failed to decode response trailer: %v", trailer.Get(trailingMetadataKey))
	}
	if string(decodedTrailer) != string(trailingMetadataValue) {
		log.Fatalf("Got trailer %s; want %s", string(trailer.Get(trailingMetadataKey)), string(trailingMetadataValue))
	}
}

// DoCustomMetadata checks that metadata is echoed back to the client.
func DoCustomMetadata(tc connectpb.TestServiceClient) {
	// Testing with UnaryCall.
	pl := ClientNewPayload(testpb.PayloadType_COMPRESSABLE, 1)
	req := &testpb.SimpleRequest{
		ResponseType: testpb.PayloadType_COMPRESSABLE,
		ResponseSize: int32(1),
		Payload:      pl,
	}
	ctx := context.Background()
	connectReq := connect.NewRequest(req)
	connectReq.Header().Set(initialMetadataKey, initialMetadataValue)
	connectReq.Header().Set(trailingMetadataKey, connect.EncodeBinaryHeader(trailingMetadataValue))
	reply, err := tc.UnaryCall(
		ctx,
		connectReq,
	)
	if err != nil {
		log.Fatal("/TestService/UnaryCall RPC failed: ", err)
	}
	t := reply.Msg.GetPayload().GetType()
	s := len(reply.Msg.GetPayload().GetBody())
	if t != testpb.PayloadType_COMPRESSABLE || s != 1 {
		log.Fatalf("Got the reply with type %d len %d; want %d, %d", t, s, testpb.PayloadType_COMPRESSABLE, 1)
	}
	validateMetadata(reply.Header(), reply.Trailer())

	// Testing with FullDuplex.
	stream := tc.FullDuplexCall(ctx)
	if stream == nil {
		log.Fatalf("%v.FullDuplexCall(_) = <nil>", tc)
	}
	respParam := []*testpb.ResponseParameters{
		{
			Size: 1,
		},
	}
	streamReq := &testpb.StreamingOutputCallRequest{
		ResponseType:       testpb.PayloadType_COMPRESSABLE,
		ResponseParameters: respParam,
		Payload:            pl,
	}
	stream.RequestHeader().Set(initialMetadataKey, initialMetadataValue)
	stream.RequestHeader().Set(trailingMetadataKey, connect.EncodeBinaryHeader(trailingMetadataValue))
	if err := stream.Send(streamReq); err != nil {
		log.Fatalf("%v has error %v while sending %v", stream, err, streamReq)
	}
	if _, err := stream.Receive(); err != nil {
		log.Fatalf("%v.Receive() = %v", stream, err)
	}
	if err := stream.CloseSend(); err != nil {
		log.Fatalf("%v.CloseSend() = %v, want <nil>", stream, err)
	}
	if _, err := stream.Receive(); !errors.Is(err, io.EOF) {
		log.Fatalf("%v failed to complete the custom metadata test: %v", stream, err)
	}
	validateMetadata(stream.ResponseHeader(), stream.ResponseTrailer())
	fmt.Println("successful custom metadata")
}

// DoStatusCodeAndMessage checks that the status code is propagated back to the client.
func DoStatusCodeAndMessage(tc connectpb.TestServiceClient) {
	code := int32(2)
	msg := "test status message"
	expectedErr := connect.NewError(
		connect.CodeUnknown, // grpc error code 2
		errors.New(msg),
	)
	respStatus := &testpb.EchoStatus{
		Code:    code,
		Message: msg,
	}
	// Test UnaryCall.
	req := &testpb.SimpleRequest{
		ResponseStatus: respStatus,
	}
	if _, err := tc.UnaryCall(context.Background(), connect.NewRequest(req)); err == nil || !errors.Is(err, expectedErr) {
		log.Fatalf("%v.UnaryCall(_, %v) = _, %v, want _, %v", tc, req, err, expectedErr)
	}
	// Test FullDuplexCall.
	stream := tc.FullDuplexCall(context.Background())
	if stream == nil {
		log.Fatalf("%v.FullDuplexCall(_) = <nil>", tc)
	}
	streamReq := &testpb.StreamingOutputCallRequest{
		ResponseStatus: respStatus,
	}
	if err := stream.Send(streamReq); err != nil {
		log.Fatalf("%v has error %v while sending %v, want <nil>", stream, err, streamReq)
	}
	if err := stream.CloseSend(); err != nil {
		log.Fatalf("%v.CloseSend() = %v, want <nil>", stream, err)
	}
	if _, err := stream.Receive(); !errors.Is(err, expectedErr) {
		log.Fatalf("%v.Recv() returned error %v, want %v", stream, err, expectedErr)
	}
	fmt.Println("successful code and message")
}

// DoSpecialStatusMessage verifies Unicode and whitespace is correctly processed
// in status message.
func DoSpecialStatusMessage(tc connectpb.TestServiceClient) {
	const (
		code int32  = 2
		msg  string = "\t\ntest with whitespace\r\nand Unicode BMP ☺ and non-BMP 😈\t\n"
	)
	expectedErr := connect.NewError(connect.CodeUnknown, errors.New(msg))
	req := &testpb.SimpleRequest{
		ResponseStatus: &testpb.EchoStatus{
			Code:    code,
			Message: msg,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := tc.UnaryCall(ctx, connect.NewRequest(req)); err == nil || !errors.Is(err, expectedErr) {
		log.Fatalf("%v.UnaryCall(_, %v) = _, %v, want _, %v", tc, req, err, expectedErr)
	}
	fmt.Println("successful code and message")
}

// DoUnimplementedService attempts to call a method from an unimplemented service.
func DoUnimplementedService(tc connectpb.UnimplementedServiceClient) {
	_, err := tc.UnimplementedCall(context.Background(), connect.NewRequest(&testpb.Empty{}))
	if connect.CodeOf(err) != connect.CodeUnimplemented {
		log.Fatalf("%v.UnimplementedCall() = _, %v, want _, %v", tc, err, connect.CodeUnimplemented)
	}
	fmt.Println("successful unimplemented service")
}

// DoPickFirstUnary runs multiple RPCs (rpcCount) and checks that all requests
// are sent to the same backend.
func DoPickFirstUnary(tc connectpb.TestServiceClient) {
	const rpcCount = 100

	pl := ClientNewPayload(testpb.PayloadType_COMPRESSABLE, 1)
	req := &testpb.SimpleRequest{
		ResponseType: testpb.PayloadType_COMPRESSABLE,
		ResponseSize: int32(1),
		Payload:      pl,
		FillServerId: true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var serverID string
	for i := 0; i < rpcCount; i++ {
		resp, err := tc.UnaryCall(ctx, connect.NewRequest(req))
		if err != nil {
			log.Fatalf("iteration %d, failed to do UnaryCall: %v", i, err)
		}
		id := resp.Msg.ServerId
		if id == "" {
			log.Fatalf("iteration %d, got empty server ID", i)
		}
		if i == 0 {
			serverID = id
			continue
		}
		if serverID != id {
			log.Fatalf("iteration %d, got different server ids: %q vs %q", i, serverID, id)
		}
	}
}

func doOneSoakIteration(ctx context.Context, tc connectpb.TestServiceClient, resetChannel bool, serverAddr string) (latency time.Duration, err error) {
	start := time.Now()
	client := tc
	if resetChannel {
		newClient, err := connectpb.NewTestServiceClient(&http.Client{}, serverAddr)
		if err != nil {
			return time.Nanosecond, err
		}
		client = newClient
	}
	// per test spec, don't include channel shutdown in latency measurement
	defer func() { latency = time.Since(start) }()
	// do a large-unary RPC
	pl := ClientNewPayload(testpb.PayloadType_COMPRESSABLE, largeReqSize)
	req := &testpb.SimpleRequest{
		ResponseType: testpb.PayloadType_COMPRESSABLE,
		ResponseSize: int32(largeRespSize),
		Payload:      pl,
	}
	reply, err := client.UnaryCall(ctx, connect.NewRequest(req))
	if err != nil {
		err = fmt.Errorf("/TestService/UnaryCall RPC failed: %s", err)
		return
	}
	t := reply.Msg.GetPayload().GetType()
	s := len(reply.Msg.GetPayload().GetBody())
	if t != testpb.PayloadType_COMPRESSABLE || s != largeRespSize {
		err = fmt.Errorf("got the reply with type %d len %d; want %d, %d", t, s, testpb.PayloadType_COMPRESSABLE, largeRespSize)
		return
	}
	return
}

// DoSoakTest runs large unary RPCs in a loop for a configurable number of times, with configurable failure thresholds.
// If resetChannel is false, then each RPC will be performed on tc. Otherwise, each RPC will be performed on a new
// stub that is created with the provided server address and dial options.
func DoSoakTest(tc connectpb.TestServiceClient, serverAddr string, resetChannel bool, soakIterations int, maxFailures int, perIterationMaxAcceptableLatency time.Duration, overallDeadline time.Time) {
	start := time.Now()
	ctx, cancel := context.WithDeadline(context.Background(), overallDeadline)
	defer cancel()
	iterationsDone := 0
	totalFailures := 0
	hopts := stats.HistogramOptions{
		NumBuckets:     20,
		GrowthFactor:   1,
		BaseBucketSize: 1,
		MinValue:       0,
	}
	h := stats.NewHistogram(hopts)
	for i := 0; i < soakIterations; i++ {
		if time.Now().After(overallDeadline) {
			break
		}
		iterationsDone++
		latency, err := doOneSoakIteration(ctx, tc, resetChannel, serverAddr)
		latencyMs := int64(latency / time.Millisecond)
		h.Add(latencyMs)
		if err != nil {
			totalFailures++
			fmt.Fprintf(os.Stderr, "soak iteration: %d elapsed_ms: %d failed: %s\n", i, latencyMs, err)
			continue
		}
		if latency > perIterationMaxAcceptableLatency {
			totalFailures++
			fmt.Fprintf(os.Stderr, "soak iteration: %d elapsed_ms: %d exceeds max acceptable latency: %d\n", i, latencyMs, perIterationMaxAcceptableLatency.Milliseconds())
			continue
		}
		fmt.Fprintf(os.Stderr, "soak iteration: %d elapsed_ms: %d succeeded\n", i, latencyMs)
	}
	var b bytes.Buffer
	h.Print(&b)
	fmt.Fprintln(os.Stderr, "Histogram of per-iteration latencies in milliseconds:")
	fmt.Fprintln(os.Stderr, b.String())
	fmt.Fprintf(os.Stderr, "soak test ran: %d / %d iterations. total failures: %d. max failures threshold: %d. See breakdown above for which iterations succeeded, failed, and why for more info.\n", iterationsDone, soakIterations, totalFailures, maxFailures)
	if iterationsDone < soakIterations {
		log.Fatalf("soak test consumed all %f seconds of time and quit early, only having ran %d out of desired %d iterations.", overallDeadline.Sub(start).Seconds(), iterationsDone, soakIterations)
	}
	if totalFailures > maxFailures {
		log.Fatalf("soak test total failures: %d exceeds max failures threshold: %d.", totalFailures, maxFailures)
	}
}
