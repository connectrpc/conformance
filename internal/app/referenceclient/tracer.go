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
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"

	"connectrpc.com/conformance/internal/tracer"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// The key associated with the wire details information stored in context.
type wireCtxKey struct{}

type wireInterceptor struct {
	Transport http.RoundTripper
}

// newWireInterceptor creates a new wireInterceptor which wraps the given transport
// in a TracingRoundTripper.
func newWireInterceptor(transport http.RoundTripper, trace *tracer.Tracer) http.RoundTripper {
	return &wireInterceptor{
		Transport: tracer.TracingRoundTripper(transport, &wireTracer{
			tracer: trace,
		}),
	}
}

// RoundTrip replaces the response body with a wireReader which captures bytes
// as they are read.
func (w *wireInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := w.Transport.RoundTrip(req)
	wrapper, ok := req.Context().Value(wireCtxKey{}).(*wireWrapper)
	if err != nil || !ok {
		return resp, err
	}
	// If this is a unary error with JSON body, replace the body with a reader
	// that will save off the body bytes as they are read so that we can access
	// the body contents in the tracer
	if isUnaryJSONError(resp.Header.Get("content-type"), int32(resp.StatusCode)) {
		resp.Body = &wireReader{body: resp.Body, wrapper: wrapper}
	}
	return resp, nil
}

type wireReader struct {
	body    io.ReadCloser
	wrapper *wireWrapper
}

func (w *wireReader) Read(p []byte) (int, error) {
	n, err := w.body.Read(p)

	// Capture bytes as they are read
	w.wrapper.buf.Write(p[:n])

	return n, err
}

func (w *wireReader) Close() error {
	return w.body.Close()
}

type wireWrapper struct {
	val atomic.Pointer[tracer.Trace]
	// buf represents the read response body
	buf *bytes.Buffer
}

// withWireCapture returns a new context which will contain wire details during
// a roundtrip.
func withWireCapture(ctx context.Context) context.Context {
	return context.WithValue(ctx, wireCtxKey{}, &wireWrapper{
		buf: &bytes.Buffer{},
	})
}

// setWireTrace sets the given trace in the given context.
func setWireTrace(ctx context.Context, trace *tracer.Trace) {
	wrapper, ok := ctx.Value(wireCtxKey{}).(*wireWrapper)
	if !ok {
		return
	}
	wrapper.val.Store(trace)
}

// examineWireDetails examines certain wire details of the call and returns the
// HTTP status code (if there is one) and feedback that describes issues with
// other parts of the response.
func examineWireDetails(ctx context.Context) (*int32, []string) {
	wrapper, ok := ctx.Value(wireCtxKey{}).(*wireWrapper)
	if !ok {
		return nil, []string{"Unable to examine wire details. Call context not configured (no wire wrapper found)."}
	}
	trace := wrapper.val.Load()
	// A nil response in the trace is valid if the HTTP round trip failed.
	if trace == nil || trace.Response == nil {
		return nil, nil
	}
	statusCode := proto.Int32(int32(trace.Response.StatusCode))

	var jsonRaw structpb.Struct
	contentType := trace.Response.Header.Get("content-type")

	// If this is a unary request that returned an error, then use the entire
	// response body as the wire error details.
	if isUnaryJSONError(contentType, *statusCode) { //nolint:nestif
		body := wrapper.buf.Bytes()
		if err := protojson.Unmarshal(body, &jsonRaw); err != nil {
			return statusCode, []string{fmt.Sprintf("Connect unary error could not be parsed: %v", err)}
		}
	} else if strings.HasPrefix(contentType, "application/connect+") {
		// If this is a streaming request, then look through the trace events
		// for the ResponseBodyEndStream event and parse its content into an
		// endStreamError to see if there are any error details.
		type endStreamError struct {
			Error json.RawMessage `json:"error"`
		}
		for _, ev := range trace.Events {
			switch eventType := ev.(type) {
			case *tracer.ResponseBodyEndStream:
				var endStream endStreamError
				if err := json.Unmarshal([]byte(eventType.Content), &endStream); err != nil {
					return statusCode, []string{fmt.Sprintf("Connect end-stream message could not be parsed: %v", err)}
				}
				// If we unmarshalled any bytes into endStream.Error, then unmarshal _that_
				// into a Struct
				if len(endStream.Error) > 0 {
					if err := protojson.Unmarshal(endStream.Error, &jsonRaw); err != nil {
						return statusCode, []string{fmt.Sprintf("Connect stream error could not be parsed: %v", err)}
					}
				}
			default:
				// Do nothing
			}
		}
	}

	// TODO: more checks on JSON raw; also add checks for gRPC end-stream message, and for HTTP trailers.
	return statusCode, nil
}

type wireTracer struct {
	tracer *tracer.Tracer
}

// Complete intercepts the Complete call for a tracer, extracting wire details
// from the passed trace. The wire details will be stored in the context acquired by
// withWireCapture and can be retrieved via examineWireDetails.
func (t *wireTracer) Complete(trace tracer.Trace) {
	setWireTrace(trace.Request.Context(), &trace)

	if t.tracer != nil {
		t.tracer.Complete(trace)
	}
}

// isUnaryJSONError returns whether the given content type and HTTP status code
// represents a unary JSON error.
func isUnaryJSONError(contentType string, statusCode int32) bool {
	return contentType == "application/json" && statusCode != 200
}
