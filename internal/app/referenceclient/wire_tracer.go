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
	"encoding/json"
	"io"
	"strings"

	"sync/atomic"

	"connectrpc.com/conformance/internal"
	v1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/tracer"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type wireKey struct{}

type WireDetails struct {
	// The actual HTTP status code observed.
	StatusCode int32
	// The actual trailers observed.
	Trailers []*v1.Header
	// The actual JSON observed on the wire in case of an error from a Connect server.
	// This will only be non-nil if the protocol is Connect and an error occurred.
	ConnectErrorRaw *structpb.Struct
}

type WireWrapper struct {
	val atomic.Pointer[*WireDetails]
}

func WithWireCapture(ctx context.Context) (context.Context, *WireWrapper) {
	wrappers := &WireWrapper{}
	ctx = context.WithValue(ctx, wireKey{}, wrappers)
	return ctx, wrappers
}

func (t *WireWrapper) Get() *WireDetails {
	respPtr := t.val.Load()
	if respPtr == nil {
		return nil
	}
	return *respPtr
}

type endStreamError struct {
	Error json.RawMessage `json:"error"`
}

type wireTracer struct {
	tracer *tracer.Tracer
}

func NewWireTracer(trace *tracer.Tracer) *wireTracer {
	return &wireTracer{
		tracer: trace,
	}
}

func (t *wireTracer) Complete(trace tracer.Trace) {
	// p := internal.NewPrinter(os.Stderr)
	// trace.Print(p)
	respWrapper, ok := trace.Request.Context().Value(wireKey{}).(*WireWrapper)
	if ok {
		if trace.Response != nil {
			statusCode := int32(trace.Response.StatusCode)

			var jsonRaw structpb.Struct
			contentType := trace.Response.Header.Get("content-type")
			if contentType == "application/json" {
				if statusCode != 200 {
					// If this is a unary request, then use the entire response body
					// as the wire error details.
					body, err := io.ReadAll(trace.Response.Body)
					if err != nil {
						return
					}
					if err := protojson.Unmarshal(body, &jsonRaw); err != nil {
						return
					}
				}
			} else if strings.HasPrefix(contentType, "application/connect+") {
				// If this is a streaming request, then look through the trace events
				// for the ResponseBodyEndStream event and parse its content into an
				// endStreamError to see if there are any error details.
				for _, ev := range trace.Events {
					switch eventType := ev.(type) {
					case *tracer.ResponseBodyEndStream:
						var endStream endStreamError
						json.Unmarshal([]byte(eventType.Content), &endStream)
						if err := protojson.Unmarshal(endStream.Error, &jsonRaw); err != nil {
							return
						}
						break
					}
				}
			}

			wire := &WireDetails{
				StatusCode:      statusCode,
				Trailers:        internal.ConvertToProtoHeader(trace.Response.Trailer),
				ConnectErrorRaw: &jsonRaw,
			}

			respWrapper.val.Store(&wire)
			// } else {
			// 	p := internal.NewPrinter(os.Stderr)
			// 	trace.Print(p)
		}
	}

	if t != nil {
		t.tracer.Complete(trace)
	}
}
