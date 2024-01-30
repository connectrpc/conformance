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

	"sync/atomic"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type respKey struct{}

type WireDetails struct {
	StatusCode      int32
	RawErrorDetails *structpb.Struct
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

func (t *contextTracer) Complete(trace Trace) {
	respWrapper, ok := trace.Request.Context().Value(respKey{}).(*RespWrapper)
	if ok {
		wire := &WireDetails{
			StatusCode: int32(trace.Response.StatusCode),
		}

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
		respWrapper.val.Store(&wire)

	}

	// p := internal.NewPrinter(os.Stderr)

	// trace.Print(p)

	if t != nil {
		t.tracer.Complete(trace)
	}
}
