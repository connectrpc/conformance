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

package referenceserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"connectrpc.com/conformance/internal"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1/conformancev1connect"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
)

var errNoRawResponseHolder = errors.New("reference server needs to use raw response but no RawHTTPResponse holder in context")
var errNonRawResponseStarted = errors.New("reference server needs to use raw response but non-raw response already started")

// rawResponseKey is used to store the raw response that the server
// should send in the context. The value type will be **conformancev1.RawHTTPResponse.
type rawResponseKey struct{}

// rawResponder is HTTP middleware that can send back a raw HTTP response
// if so directed. The handler directs it to do so by storing a raw response
// in a placeholder context value before otherwise interacting with the
// response writer. (If response writer interactions and no raw response has
// been stored, those response writer interactions take precedence and no
// raw response can be sent.)
func rawResponder(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
		snapshotHeaders := respWriter.Header().Clone()
		rawResponder := &rawResponseWriter{respWriter: respWriter}
		ctx := context.WithValue(req.Context(), rawResponseKey{}, rawResponder)
		req = req.WithContext(ctx)
		handler.ServeHTTP(rawResponder, req)
		rawResponder.finish(snapshotHeaders)
	})
}

type rawResponseWriter struct {
	respWriter      http.ResponseWriter
	mu              sync.Mutex
	rawResp         *conformancev1.RawHTTPResponse
	startedResponse bool
}

// canSendResponse returns true if the server handler can use the
// http.ResponseWriter methods to send a response. This returns false
// if we will instead be finishing the call with a raw response.
func (r *rawResponseWriter) canSendResponse() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.startedResponse {
		return true
	}
	if r.rawResp == nil {
		r.startedResponse = true
		return true
	}
	return false
}

// rawResponse returns non-nil if the call will be finished with a
// raw response. If it returns nil, nothing need be done to finish
// the call; the server handler was already allowed to send it.
func (r *rawResponseWriter) rawResponse() *conformancev1.RawHTTPResponse {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rawResp
}

func (r *rawResponseWriter) setRawResponse(resp *conformancev1.RawHTTPResponse) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.startedResponse {
		return false
	}
	r.rawResp = resp
	return true
}

func (r *rawResponseWriter) Header() http.Header {
	return r.respWriter.Header()
}

func (r *rawResponseWriter) Write(bytes []byte) (int, error) {
	if r.canSendResponse() {
		return r.respWriter.Write(bytes)
	}
	return len(bytes), nil
}

func (r *rawResponseWriter) WriteHeader(statusCode int) {
	if r.canSendResponse() {
		r.respWriter.WriteHeader(statusCode)
	}
}

func (r *rawResponseWriter) Flush() {
	if r.canSendResponse() {
		if flusher, ok := r.respWriter.(http.Flusher); ok {
			flusher.Flush()
		}
	}
}

func (r *rawResponseWriter) Unwrap() http.ResponseWriter {
	return r.respWriter
}

func (r *rawResponseWriter) finish(snapshotHeaders http.Header) {
	resp := r.rawResponse()
	if resp == nil {
		return
	}

	// clean any headers that may have been set by the handler
	// and restore to the snapshot we initially took (which
	// may have headers set by earlier middleware, like CORS)
	for k := range r.respWriter.Header() {
		delete(r.respWriter.Header(), k)
	}
	for k, v := range snapshotHeaders {
		r.respWriter.Header()[k] = v
	}

	internal.AddHeaders(resp.Headers, r.respWriter.Header())
	r.respWriter.Header()["Date"] = nil // suppress automatic date header
	// We must pre-declare trailers to make sure that chunked encoding is used and
	// trailers can actually be sent.
	for _, hdr := range resp.Trailers {
		r.respWriter.Header().Add("Trailer", hdr.Name)
	}
	statusCode := int(resp.StatusCode)
	// If no status code was specified in the raw response, default to 200
	if statusCode == 0 {
		statusCode = 200
	}
	r.respWriter.WriteHeader(statusCode)
	switch contents := resp.Body.(type) {
	case *conformancev1.RawHTTPResponse_Unary:
		_ = internal.WriteRawMessageContents(contents.Unary, r.respWriter)
	case *conformancev1.RawHTTPResponse_Stream:
		_ = internal.WriteRawStreamContents(contents.Stream, r.respWriter)
	}
	internal.AddTrailers(resp.Trailers, r.respWriter.Header())
}

type rawResponseRecorder struct{}

func (r rawResponseRecorder) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if msg, ok := req.Any().(*conformancev1.UnaryRequest); ok {
			rawResponse := msg.GetResponseDefinition().GetRawResponse()
			if rawResponse != nil {
				if err := setRawResponse(ctx, rawResponse); err != nil {
					return nil, err
				}
				return nil, connect.NewError(connect.CodeAborted, errors.New("use raw response instead"))
			}
		}
		return next(ctx, req)
	}
}

func (r rawResponseRecorder) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (r rawResponseRecorder) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, stream connect.StreamingHandlerConn) error {
		var req proto.Message
		var rawResponseFunc func() *conformancev1.RawHTTPResponse
		switch stream.Spec().Procedure {
		case conformancev1connect.ConformanceServiceClientStreamProcedure:
			streamReq := &conformancev1.ClientStreamRequest{}
			rawResponseFunc = func() *conformancev1.RawHTTPResponse {
				return streamReq.GetResponseDefinition().GetRawResponse()
			}
			req = streamReq
		case conformancev1connect.ConformanceServiceServerStreamProcedure:
			streamReq := &conformancev1.ServerStreamRequest{}
			rawResponseFunc = func() *conformancev1.RawHTTPResponse {
				return streamReq.GetResponseDefinition().GetRawResponse()
			}
			req = streamReq
		case conformancev1connect.ConformanceServiceBidiStreamProcedure:
			streamReq := &conformancev1.BidiStreamRequest{}
			rawResponseFunc = func() *conformancev1.RawHTTPResponse {
				return streamReq.GetResponseDefinition().GetRawResponse()
			}
			req = streamReq
		}
		var reqErr error
		if req == nil {
			return next(ctx, stream)
		}
		reqErr = stream.Receive(req)
		if reqErr == nil { //nolint:nestif
			rawResponse := rawResponseFunc()
			if rawResponse != nil {
				if err := setRawResponse(ctx, rawResponse); err != nil {
					return err
				}
				// If we have a raw response, go ahead and drain the request stream
				// before sending back the raw response.
				// NOTE: This means that raw responses cannot be used with full-duplex
				//       request definitions.
				for {
					if err := stream.Receive(req); err != nil {
						break
					}
				}
				return connect.NewError(connect.CodeAborted, errors.New("use raw response instead"))
			}
		}
		return next(ctx, &firstReqCachingStream{
			StreamingHandlerConn: stream,
			request:              req,
			recvErr:              reqErr,
		})
	}
}

type firstReqCachingStream struct {
	connect.StreamingHandlerConn
	request any
	recvErr error
}

func (str *firstReqCachingStream) Receive(dest any) error {
	if str.recvErr != nil {
		err := str.recvErr
		str.request, str.recvErr = nil, nil
		return err
	}
	if str.request != nil {
		destMsg, ok := dest.(proto.Message)
		if !ok {
			return fmt.Errorf("%T does not implement proto.Message", dest)
		}
		srcMsg, ok := str.request.(proto.Message)
		if !ok {
			return fmt.Errorf("%T does not implement proto.Message", dest)
		}
		proto.Reset(destMsg)
		proto.Merge(destMsg, srcMsg)
		str.request, str.recvErr = nil, nil
		return nil
	}
	// Otherwise, we've already provided the cached first request.
	// So all subsequent receives use the underlying stream.
	return str.StreamingHandlerConn.Receive(dest)
}

func setRawResponse(ctx context.Context, resp *conformancev1.RawHTTPResponse) error {
	respWriter, ok := ctx.Value(rawResponseKey{}).(*rawResponseWriter)
	if !ok {
		return errNoRawResponseHolder
	}
	if !respWriter.setRawResponse(resp) {
		return errNonRawResponseStarted
	}
	return nil
}
