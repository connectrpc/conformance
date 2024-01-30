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

package tracer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"sync/atomic"

	"connectrpc.com/conformance/internal"
	v1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type respKey struct{}

type WireDetails struct {
	StatusCode      int32
	RawErrorDetails *structpb.Struct
	Trailers        []*v1.Header
}

type RespWrapper struct {
	val atomic.Pointer[*WireDetails]
}

func WithResponseCapture(ctx context.Context) (context.Context, *RespWrapper) {
	wrappers := &RespWrapper{}
	ctx = context.WithValue(ctx, respKey{}, wrappers)
	return ctx, wrappers
}

// Get returns the resp captured. Resps are not captured until the response body is
// exhausted.
func (t *RespWrapper) Get() *WireDetails {
	respPtr := t.val.Load()
	if respPtr == nil {
		return nil
	}
	return *respPtr
}

type endStreamError struct {
	Error json.RawMessage `json:"error"`
}

type contextTracer struct {
	Tracer

	tracer *Tracer
}

func NewContextTracer(trace *Tracer) *contextTracer {
	return &contextTracer{
		tracer: trace,
	}
}

func isConnectStreaming(trace Trace) bool {
	contentType := trace.Response.Header.Get("content-type")
	return strings.HasPrefix(contentType, "application/connect+")
}
func isConnectUnary(trace Trace) bool {
	contentType := trace.Response.Header.Get("content-type")
	return contentType == "application/json"
}
func isConnect(trace Trace) bool {
	return isConnectUnary(trace) || isConnectStreaming(trace)
}

func (t *contextTracer) Complete(trace Trace) {
	respWrapper, ok := trace.Request.Context().Value(respKey{}).(*RespWrapper)
	p := internal.NewPrinter(os.Stderr)

	trace.Print(p)

	if ok {
		statusCode := int32(trace.Response.StatusCode)
		wire := &WireDetails{
			StatusCode: statusCode,
			Trailers:   internal.ConvertToProtoHeader(trace.Response.Trailer),
		}

		if statusCode != 200 {
			// Get raw Connect error details only if this is the Connect protocol
			if isConnectUnary(trace) {
				body, err := io.ReadAll(trace.Response.Body)
				if err != nil {
					return
				}
				fmt.Fprintf(os.Stderr, "Das booty", string(body))
				var jsonRaw structpb.Struct
				if err := protojson.Unmarshal(body, &jsonRaw); err != nil {
					fmt.Fprintf(os.Stderr, "OH HELL NO", err)
					return
				}
				wire.RawErrorDetails = &jsonRaw
			} else if isConnectStreaming(trace) {
				for _, ev := range trace.Events {
					switch eventType := ev.(type) {
					case *ResponseBodyEndStream:
						var endStream endStreamError
						json.Unmarshal([]byte(eventType.Content), &endStream)

						var jsonRaw structpb.Struct
						if err := protojson.Unmarshal(endStream.Error, &jsonRaw); err != nil {
							return
						}
						wire.RawErrorDetails = &jsonRaw
					}
				}
			}
		}
		respWrapper.val.Store(&wire)

	}

	if t != nil {
		t.tracer.Complete(trace)
	}
}
