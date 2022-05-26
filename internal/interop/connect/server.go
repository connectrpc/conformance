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
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	testrpc "github.com/bufbuild/connect-crosstest/internal/gen/proto/connect/grpc/testing/testingconnect"
	testpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/grpc/testing"
	"github.com/bufbuild/connect-go"
)

const NonASCIIErrMsg = "soirée 🎉" // readable non-ASCII

type testServer struct {
	testrpc.UnimplementedTestServiceHandler
}

func NewTestConnectServer() testrpc.TestServiceHandler {
	return &testServer{}
}

func serverNewPayload(payloadType testpb.PayloadType, size int32) (*testpb.Payload, error) {
	if size < 0 {
		return nil, fmt.Errorf("requested a response with invalid length %d", size)
	}
	body := make([]byte, size)
	switch payloadType {
	case testpb.PayloadType_COMPRESSABLE:
	default:
		return nil, fmt.Errorf("unsupported payload type: %d", payloadType)
	}
	return &testpb.Payload{
		Type: payloadType,
		Body: body,
	}, nil
}

func (s *testServer) EmptyCall(ctx context.Context, req *connect.Request[testpb.Empty]) (*connect.Response[testpb.Empty], error) {
	return connect.NewResponse(new(testpb.Empty)), nil
}

func (s *testServer) UnaryCall(ctx context.Context, req *connect.Request[testpb.SimpleRequest]) (*connect.Response[testpb.SimpleResponse], error) {
	if st := req.Msg.GetResponseStatus(); st != nil && st.Code != 0 {
		return nil, connect.NewError(connect.Code(st.Code), errors.New(st.Message))
	}
	pl, err := serverNewPayload(req.Msg.GetResponseType(), req.Msg.GetResponseSize())
	if err != nil {
		return nil, err
	}
	res := connect.NewResponse(&testpb.SimpleResponse{
		Payload: pl,
	})
	if initialMetadata := req.Header().Values(initialMetadataKey); len(initialMetadata) != 0 {
		for _, value := range initialMetadata {
			res.Header().Add(initialMetadataKey, value)
		}
	}
	if trailingMetadata := req.Header().Values(trailingMetadataKey); len(trailingMetadata) != 0 {
		for _, value := range trailingMetadata {
			decodedTrailingMetadata, err := connect.DecodeBinaryHeader(value)
			if err != nil {
				return nil, err
			}
			res.Trailer().Add(trailingMetadataKey, connect.EncodeBinaryHeader(decodedTrailingMetadata))
		}
	}
	return res, nil
}

func (s *testServer) FailUnaryCall(ctx context.Context, in *connect.Request[testpb.SimpleRequest]) (*connect.Response[testpb.SimpleResponse], error) {
	return nil, connect.NewError(connect.CodeResourceExhausted, errors.New(NonASCIIErrMsg))
}

func (s *testServer) StreamingOutputCall(ctx context.Context, args *connect.Request[testpb.StreamingOutputCallRequest], stream *connect.ServerStream[testpb.StreamingOutputCallResponse]) error {
	for _, param := range args.Msg.GetResponseParameters() {
		if us := param.GetIntervalUs(); us > 0 {
			time.Sleep(time.Duration(us) * time.Microsecond)
		}
		// Checking if the context is canceled or deadline exceeded, in a real world usage it will
		// make more sense to put this checking before the expensive works (i.e. the time.Sleep above),
		// but in order to simulate a network latency issue, we put the context checking here.
		if err := ctx.Err(); err != nil {
			return err
		}
		pl, err := serverNewPayload(args.Msg.GetResponseType(), param.GetSize())
		if err != nil {
			return err
		}
		if err := stream.Send(&testpb.StreamingOutputCallResponse{
			Payload: pl,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *testServer) StreamingInputCall(ctx context.Context, stream *connect.ClientStream[testpb.StreamingInputCallRequest, testpb.StreamingInputCallResponse]) error {
	var sum int
	for stream.Receive() {
		if err := ctx.Err(); err != nil {
			return err
		}
		p := stream.Msg().GetPayload().GetBody()
		sum += len(p)
	}
	if err := stream.Err(); err != nil {
		return err
	}
	return stream.SendAndClose(connect.NewResponse(
		&testpb.StreamingInputCallResponse{
			AggregatedPayloadSize: int32(sum),
		},
	))
}

func (s *testServer) FullDuplexCall(ctx context.Context, stream *connect.BidiStream[testpb.StreamingOutputCallRequest, testpb.StreamingOutputCallResponse]) error {
	if initialMetadata := stream.RequestHeader().Values(initialMetadataKey); len(initialMetadata) != 0 {
		for _, value := range initialMetadata {
			stream.ResponseHeader().Add(initialMetadataKey, value)
		}
	}
	if trailingMetadata := stream.RequestHeader().Values(trailingMetadataKey); len(trailingMetadata) != 0 {
		for _, value := range trailingMetadata {
			decodedTrailingMetadata, err := connect.DecodeBinaryHeader(value)
			if err != nil {
				return err
			}
			stream.ResponseTrailer().Add(trailingMetadataKey, connect.EncodeBinaryHeader(decodedTrailingMetadata))
		}
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		req, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			// read done.
			return nil
		} else if err != nil {
			return err
		}
		st := req.GetResponseStatus()
		if st != nil && st.Code != 0 {
			return connect.NewError(connect.Code(st.Code), errors.New(st.Message))
		}
		cs := req.GetResponseParameters()
		for _, c := range cs {
			if us := c.GetIntervalUs(); us > 0 {
				time.Sleep(time.Duration(us) * time.Microsecond)
			}
			pl, err := serverNewPayload(req.GetResponseType(), c.GetSize())
			if err != nil {
				return err
			}
			if err := stream.Send(&testpb.StreamingOutputCallResponse{
				Payload: pl,
			}); err != nil {
				return err
			}
		}
	}
}

func (s *testServer) HalfDuplexCall(ctx context.Context, stream *connect.BidiStream[testpb.StreamingOutputCallRequest, testpb.StreamingOutputCallResponse]) error {
	var msgBuf []*testpb.StreamingOutputCallRequest
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		req, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			// read done.
			break
		}
		if err != nil {
			return err
		}
		msgBuf = append(msgBuf, req)
	}
	for _, msg := range msgBuf {
		cs := msg.GetResponseParameters()
		for _, c := range cs {
			if us := c.GetIntervalUs(); us > 0 {
				time.Sleep(time.Duration(us) * time.Microsecond)
			}
			pl, err := serverNewPayload(msg.GetResponseType(), c.GetSize())
			if err != nil {
				return err
			}
			if err := stream.Send(&testpb.StreamingOutputCallResponse{
				Payload: pl,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}
