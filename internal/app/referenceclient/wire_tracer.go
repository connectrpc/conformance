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
	"sync"

	"connectrpc.com/conformance/internal"
	v1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/tracer"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type wireKey struct{}

// WireDetails encapsulates the wire details to track for a roundtrip.
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
	val *WireDetails
	mtx sync.Mutex
}

// Get returns the wire details.
func (w *WireWrapper) Get() *WireDetails {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	if w.val == nil {
		return nil
	}
	return w.val
}

// Lock acquires the internal lock for setting the wire details. This should
// always be called before calling Set.
func (w *WireWrapper) Lock() {
	w.mtx.Lock()
}

// Lock releases the internal lock for setting the wire details.
func (w *WireWrapper) Unlock() {
	w.mtx.Unlock()
}

// Set sets the wire details. Note that calls to Set should always be
// preceded by calls to Lock.
func (w *WireWrapper) Set(details *WireDetails) {
	w.val = details
}

// WithWireCapture returns a new context which will contain wire details during
// a roundtrip.
func WithWireCapture(ctx context.Context) (context.Context, *WireWrapper) {
	wrapper := &WireWrapper{}
	ctx = context.WithValue(ctx, wireKey{}, wrapper)
	return ctx, wrapper
}

type WireTracer struct {
	tracer *tracer.Tracer
}

// Complete intercepts the Complete call for a tracer, extracting wire details
// from the passed trace. The wire details will be stored in the context acquired byte
// WithWireCapture and can be retrieved via WireWrapper.Get().
func (t *WireTracer) Complete(trace tracer.Trace) {
	wrapper, ok := trace.Request.Context().Value(wireKey{}).(*WireWrapper)
	// Lock the mutex on the wrapper so that the client implementation isn't
	// reading the wire details before we get a chance to populate them here.
	wrapper.Lock()
	defer wrapper.Unlock()
	if ok { //nolint:nestif
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
				type endStreamError struct {
					Error json.RawMessage `json:"error"`
				}
				// If this is a streaming request, then look through the trace events
				// for the ResponseBodyEndStream event and parse its content into an
				// endStreamError to see if there are any error details.
				for _, ev := range trace.Events {
					switch eventType := ev.(type) {
					case *tracer.ResponseBodyEndStream:
						var endStream endStreamError
						if err := json.Unmarshal([]byte(eventType.Content), &endStream); err != nil {
							return
						}
						if err := protojson.Unmarshal(endStream.Error, &jsonRaw); err != nil {
							return
						}
					default:
						// Do nothing
					}
				}
			}

			wire := &WireDetails{
				StatusCode:      statusCode,
				Trailers:        internal.ConvertToProtoHeader(trace.Response.Trailer),
				ConnectErrorRaw: &jsonRaw,
			}

			wrapper.Set(wire)
		}
	}

	if t != nil {
		t.tracer.Complete(trace)
	}
}

// NewWireTracer returns a new wire tracer with the given tracer.
// If trace is nil, all tracer operations will be bypassed.
func NewWireTracer(trace *tracer.Tracer) *WireTracer {
	return &WireTracer{
		tracer: trace,
	}
}
