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
	"bytes"
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

// The key associated with the wire details information stored in context.
type wireCtxKey struct{}

// wireDetails encapsulates the wire details to track for a roundtrip.
type wireDetails struct {
	// The actual HTTP status code observed.
	StatusCode int32
	// The actual trailers observed.
	Trailers []*v1.Header
	// The actual JSON observed on the wire in case of an error from a Connect server.
	// This will only be non-nil if the protocol is Connect and an error occurred.
	ConnectErrorRaw *structpb.Struct
}

type wireWrapper struct {
	val atomic.Pointer[wireDetails]
	buf *bytes.Buffer
}

// withWireCapture returns a new context which will contain wire details during
// a roundtrip.
func withWireCapture(ctx context.Context) context.Context {
	return context.WithValue(ctx, wireCtxKey{}, &wireWrapper{
		buf: &bytes.Buffer{},
	})
}

// setWireDetails sets the given wire details in the given context.
func setWireDetails(ctx context.Context, details *wireDetails) {
	wrapper, ok := ctx.Value(wireCtxKey{}).(*wireWrapper)
	if !ok {
		return
	}
	wrapper.val.Store(details)
}

func getWireDetails(ctx context.Context) *wireDetails {
	wrapper, ok := ctx.Value(wireCtxKey{}).(*wireWrapper)
	if !ok {
		return nil
	}
	ptr := wrapper.val.Load()
	if ptr == nil {
		return nil
	}
	return ptr
}

type wireTracer struct {
	tracer *tracer.Tracer
}

// Complete intercepts the Complete call for a tracer, extracting wire details
// from the passed trace. The wire details will be stored in the context acquired by
// withWireCapture and can be retrieved via getWireDetails.
func (t *wireTracer) Complete(trace tracer.Trace) {
	wrapper, ok := trace.Request.Context().Value(wireCtxKey{}).(*wireWrapper)
	if ok {
		if trace.Response != nil { //nolint:nestif
			statusCode := int32(trace.Response.StatusCode)

			var jsonRaw structpb.Struct
			contentType := trace.Response.Header.Get("content-type")
			if contentType == "application/json" {
				if statusCode != 200 {
					// If this is a unary request, then use the entire response body
					// as the wire error details.
					body, err := io.ReadAll(wrapper.buf)
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

			wire := &wireDetails{
				StatusCode:      statusCode,
				Trailers:        internal.ConvertToProtoHeader(trace.Response.Trailer),
				ConnectErrorRaw: &jsonRaw,
			}
			// fmt.Fprintf(os.Stderr, "wire details %+v\n", wire)

			setWireDetails(trace.Request.Context(), wire)
		}
	}

	if t != nil {
		t.tracer.Complete(trace)
	}
}
