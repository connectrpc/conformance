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

package grpcserver

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"connectrpc.com/conformance/internal"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1/conformancev1connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// A half-duplex bidi call must not write until the upload is complete, or
// net/http closes the HTTP/1.1 request body while requests are still arriving.
func TestGRPCWebHalfDuplexWaitsForUploadOverHTTP1(t *testing.T) {
	t.Parallel()

	grpcServer, err := createServer(1 << 20)
	require.NoError(t, err)
	recorder := &writeOrderRecorder{}
	httpServer := httptest.NewServer(recorder.wrap(newGRPCWebHandler(grpcServer, nil)))
	t.Cleanup(httpServer.Close)

	firstRequest := envelope(t, &conformancev1.BidiStreamRequest{
		ResponseDefinition: &conformancev1.StreamResponseDefinition{
			ResponseHeaders: []*conformancev1.Header{{Name: "x-custom-header", Value: []string{"foo"}}},
			ResponseData:    [][]byte{[]byte("one"), []byte("two")},
		},
		FullDuplex: false,
	})
	secondRequest := envelope(t, &conformancev1.BidiStreamRequest{RequestData: []byte("second")})

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)
	bodyReader, bodyWriter := io.Pipe()
	// Unblocks the transport's body writer so a hung call fails at the deadline.
	stopCloser := context.AfterFunc(ctx, func() { _ = bodyWriter.CloseWithError(ctx.Err()) })
	t.Cleanup(func() { stopCloser() })
	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		httpServer.URL+conformancev1connect.ConformanceServiceBidiStreamProcedure,
		bodyReader,
	)
	require.NoError(t, err)
	httpRequest.Header.Set("Content-Type", "application/grpc-web+proto")
	go func() {
		defer func() { _ = bodyWriter.Close() }()
		if _, err := bodyWriter.Write(firstRequest); err != nil {
			return
		}
		// Leaves a window in which a server that writes early would do so.
		time.Sleep(200 * time.Millisecond)
		_, _ = bodyWriter.Write(secondRequest)
	}()

	httpResponse, err := httpServer.Client().Do(httpRequest)
	require.NoError(t, err)
	t.Cleanup(func() { _ = httpResponse.Body.Close() })
	assert.False(t, recorder.wroteEarly.Load(), "server wrote a response before the upload was complete")
	require.Equal(t, http.StatusOK, httpResponse.StatusCode)
	assert.Equal(t, "foo", httpResponse.Header.Get("X-Custom-Header"))

	responseBody, err := io.ReadAll(httpResponse.Body)
	require.NoError(t, err)
	var dataFrames int
	var trailerFrame []byte
	for len(responseBody) > 0 {
		require.GreaterOrEqual(t, len(responseBody), 5)
		frameLength := binary.BigEndian.Uint32(responseBody[1:5])
		require.GreaterOrEqual(t, len(responseBody)-5, int(frameLength))
		if responseBody[0]&0x80 != 0 {
			trailerFrame = responseBody[5 : 5+frameLength]
		} else {
			dataFrames++
		}
		responseBody = responseBody[5+frameLength:]
	}
	assert.Equal(t, 2, dataFrames)
	assert.Contains(t, string(trailerFrame), "grpc-status: 0")
}

// writeOrderRecorder records whether a handler wrote any of its response
// before it read the request body to EOF.
type writeOrderRecorder struct {
	sawEOF     atomic.Bool
	wroteEarly atomic.Bool
}

func (r *writeOrderRecorder) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
		req.Body = &eofNotingBody{ReadCloser: req.Body, sawEOF: &r.sawEOF}
		next.ServeHTTP(&writeNotingResponseWriter{ResponseWriter: respWriter, recorder: r}, req)
	})
}

func (r *writeOrderRecorder) noteWrite() {
	if !r.sawEOF.Load() {
		r.wroteEarly.Store(true)
	}
}

type eofNotingBody struct {
	io.ReadCloser

	sawEOF *atomic.Bool
}

func (b *eofNotingBody) Read(data []byte) (int, error) {
	n, err := b.ReadCloser.Read(data)
	if errors.Is(err, io.EOF) {
		b.sawEOF.Store(true)
	}
	return n, err
}

type writeNotingResponseWriter struct {
	http.ResponseWriter

	recorder *writeOrderRecorder
}

func (w *writeNotingResponseWriter) WriteHeader(statusCode int) {
	w.recorder.noteWrite()
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *writeNotingResponseWriter) Write(data []byte) (int, error) {
	w.recorder.noteWrite()
	return w.ResponseWriter.Write(data)
}

func (w *writeNotingResponseWriter) Flush() {
	w.recorder.noteWrite()
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func envelope(t *testing.T, message proto.Message) []byte {
	t.Helper()
	messageBytes, err := proto.Marshal(message)
	require.NoError(t, err)
	var buffer bytes.Buffer
	err = internal.WriteRawStreamContents(&conformancev1.StreamContents{
		Items: []*conformancev1.StreamContents_StreamItem{{
			Payload: &conformancev1.MessageContents{
				Data: &conformancev1.MessageContents_Binary{Binary: messageBytes},
			},
		}},
	}, &buffer)
	require.NoError(t, err)
	return buffer.Bytes()
}
