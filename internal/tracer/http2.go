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
	"bytes"
	"errors"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

const (
	frameHeaderLen = 9
	clientPreface  = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"

	// When the server sends a "refused" RST_STREAM frame, we will wait this long to
	// see if client auto-retries. This should be lenient enough that a client that
	// retries refused streams will can perform the retry before this period lapses.
	// But it also must be less than TraceTimeout, so if the client is *not* retrying
	// then we can still provide a trace within the timeout window.
	retryWait = 3 * time.Second
)

// TracingHTTP2Conn applies tracing to the given net.Conn, which uses the
// HTTP/2 protocol. If TLS is used, it should already be applied, so the
// given conn should be used for reading/writing clear-text (i.e.
// pre-encryption, post-decryption).
//
// If isServer is true, this is a server connection, so requests are read
// and responses are written. Otherwise, this is a client connection, and
// requests are written and responses are read.
func TracingHTTP2Conn(conn net.Conn, isServer bool, collector Collector) net.Conn {
	tracer := &tracingHTTP2Conn{
		Conn:        conn,
		isServer:    isServer,
		collector:   &http2RetryCollector{collector: collector},
		readTracer:  http2FrameTracer{isRequest: isServer},
		writeTracer: http2FrameTracer{isRequest: !isServer},
	}
	tracer.readTracer.c = tracer
	tracer.readTracer.decoder = hpack.NewDecoder(math.MaxUint32, nil)
	tracer.writeTracer.c = tracer
	tracer.writeTracer.decoder = hpack.NewDecoder(math.MaxUint32, nil)
	prefaceBytes := make([]byte, 0, len(clientPreface))
	if isServer {
		tracer.readTracer.prefaceBytes = prefaceBytes
	} else {
		tracer.writeTracer.prefaceBytes = prefaceBytes
	}
	return tracer
}

type tracingHTTP2Conn struct {
	net.Conn
	isServer  bool
	collector *http2RetryCollector

	mu          sync.Mutex
	streams     map[uint32]*http2Stream
	maxStreamID uint32
	readTracer  http2FrameTracer
	writeTracer http2FrameTracer
}

func (c *tracingHTTP2Conn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	c.readTracer.trace(b[:n])
	if err != nil {
		c.cancelAll(err)
	}
	return n, err
}

func (c *tracingHTTP2Conn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	c.writeTracer.trace(b[:n])
	if err != nil {
		c.cancelAll(err)
	}
	return n, err
}

func (c *tracingHTTP2Conn) handleFrame(frame http2.Frame, isRequest bool) {
	switch frame := frame.(type) {
	case *http2.MetaHeadersFrame:
		stream, isNew := c.getStream(frame, isRequest)
		if stream == nil {
			return
		}
		switch {
		case isNew:
		// request headers, which resulted in a new stream; nothing else to do
		case !isRequest && !stream.gotResponse:
			// response headers
			c.receiveResponse(stream, frame)
		case isRequest:
			// request trailers
			stream.builder.trace.Request.Trailer = makeHeaders(frame)
		default:
			// response trailers
			stream.builder.trace.Response.Trailer = makeHeaders(frame)
		}
		if frame.StreamEnded() {
			c.closeStream(frame.StreamID, stream, isRequest, nil)
		}
	case *http2.DataFrame:
		stream := c.getExistingStream(frame.StreamID)
		if stream == nil {
			return
		}
		if isRequest {
			stream.requestTracer.trace(frame.Data())
		} else {
			stream.responseTracer.trace(frame.Data())
		}
		if frame.StreamEnded() {
			c.closeStream(frame.StreamID, stream, isRequest, nil)
		}
	case *http2.RSTStreamFrame:
		stream := c.getExistingStream(frame.StreamID)
		if stream == nil {
			return
		}
		c.closeStream(frame.StreamID, stream, isRequest, http2.StreamError{
			StreamID: frame.StreamID,
			Code:     frame.ErrCode,
		})

	case *http2.GoAwayFrame:
		c.setMaxStreamID(frame.LastStreamID, http2.ConnectionError(frame.ErrCode))
	}
}

func (c *tracingHTTP2Conn) receiveResponse(stream *http2Stream, frame *http2.MetaHeadersFrame) {
	c.mu.Lock()
	defer c.mu.Unlock()
	stream.gotResponse = true
	resp := makeResponse(frame) //nolint:bodyclose // there is no body to close on this response
	stream.builder.add(&ResponseStart{Response: resp})
	stream.responseTracer.isStreamProtocol, stream.responseTracer.decompressor = propertiesFromHeaders(resp.Header)
	stream.responseTracer.builder = stream.builder
}

func (c *tracingHTTP2Conn) getStream(frame *http2.MetaHeadersFrame, createIfNotFound bool) (*http2Stream, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	stream := c.streams[frame.StreamID]
	if stream != nil {
		return stream, false
	}
	if !createIfNotFound {
		return nil, false
	}
	if c.maxStreamID != 0 && frame.StreamID > c.maxStreamID {
		return nil, false // stream ID too high; ignore
	}
	stream = c.newStreamLocked(frame)
	return stream, true
}

func (c *tracingHTTP2Conn) newStreamLocked(frame *http2.MetaHeadersFrame) *http2Stream {
	req := makeRequest(frame)
	builder, _ := newBuilder(req, !c.isServer, c.collector)
	isStream, decompressor := propertiesFromHeaders(req.Header)
	stream := &http2Stream{
		builder:       builder,
		requestTracer: dataTracer{isRequest: true, isStreamProtocol: isStream, decompressor: decompressor, builder: builder},
	}
	c.collector.newAttempt(builder.trace.TestName)
	if c.streams == nil {
		c.streams = map[uint32]*http2Stream{}
	}
	c.streams[frame.StreamID] = stream
	return stream
}

func (c *tracingHTTP2Conn) getExistingStream(streamID uint32) *http2Stream {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.streams[streamID]
}

func (c *tracingHTTP2Conn) closeStream(streamID uint32, stream *http2Stream, isRequest bool, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !isRequest || err != nil {
		// This is either end of response or an error, which means the
		// whole operation done.
		delete(c.streams, streamID)
	}
	if isRequest {
		stream.requestTracer.emitUnfinished()
		stream.builder.add(&RequestBodyEnd{Err: err})
	} else if stream.responseTracer.builder != nil {
		stream.responseTracer.emitUnfinished()
		stream.builder.add(&ResponseBodyEnd{Err: err})
	}
}

func (c *tracingHTTP2Conn) setMaxStreamID(maxStreamID uint32, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.maxStreamID = maxStreamID
	for streamID, stream := range c.streams {
		if streamID > maxStreamID {
			delete(c.streams, streamID)
			stream.builder.add(&ResponseBodyEnd{Err: err})
		}
	}
}

func (c *tracingHTTP2Conn) cancelAll(err error) {
	func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		for streamID, stream := range c.streams {
			delete(c.streams, streamID)
			if c.isServer {
				stream.builder.add(&ResponseBodyEnd{Err: err})
			} else {
				stream.builder.add(&RequestBodyEnd{Err: err})
			}
		}
	}()
	c.collector.cancel()
}

type http2Stream struct {
	builder        *builder
	requestTracer  dataTracer
	gotResponse    bool
	responseTracer dataTracer
}

type http2FrameTracer struct {
	c         *tracingHTTP2Conn
	isRequest bool
	decoder   *hpack.Decoder

	prefaceBytes []byte
	broken       bool
	prefix       []byte
	header       http2.FrameHeader
	frame        bytes.Buffer
	expecting    uint32
	actual       uint64
}

func (h *http2FrameTracer) trace(data []byte) {
	if h.broken {
		return
	}
	for {
		if len(data) == 0 {
			return
		}
		if h.isRequest && len(h.prefaceBytes) < len(clientPreface) {
			need := len(clientPreface) - len(h.prefaceBytes)
			if len(data) < need {
				h.prefaceBytes = append(h.prefaceBytes, data...)
				return
			}
			h.prefaceBytes = append(h.prefaceBytes, data[:need]...)
			data = data[need:]
			if !prefaceIsValid(h.prefaceBytes) {
				h.broken = true
				return
			}
			if len(data) == 0 {
				return
			}
		}

		if h.expecting == 0 {
			// still reading envelope prefix
			n, done := h.traceHeaderLocked(data)
			if !done {
				// need to read more data to finish prefix
				return
			}
			data = data[n:]
			continue
		}

		n, done := h.traceFrameLocked(data)
		if !done {
			// need to read more data to finish message
			return
		}
		data = data[n:]
	}
}

func (h *http2FrameTracer) traceHeaderLocked(data []byte) (int, bool) {
	need := frameHeaderLen - len(h.prefix)
	if len(data) < need {
		// envelope still not complete...
		h.prefix = append(h.prefix, data...)
		return need, false
	}

	h.prefix = append(h.prefix, data[:need]...)
	var err error
	h.header, err = http2.ReadFrameHeader(bytes.NewReader(h.prefix))
	if err != nil {
		h.broken = true
		return need, false
	}
	h.expecting = h.header.Length
	h.frame.Write(h.prefix)
	h.prefix = h.prefix[:0]
	if h.expecting == 0 {
		return need, h.emitFrame()
	}
	return need, true
}

func (h *http2FrameTracer) traceFrameLocked(data []byte) (int, bool) {
	need := int(h.expecting - uint32(h.actual))
	if len(data) < need {
		// message still not complete...
		h.actual += uint64(len(data))
		h.frame.Write(data)
		return need, false
	}

	h.frame.Write(data[:need])
	h.expecting = 0
	h.actual = 0
	return need, h.emitFrame()
}

func (h *http2FrameTracer) emitFrame() bool {
	defer func() {
		h.frame.Reset()
	}()
	framer := http2.NewFramer(io.Discard, &h.frame)
	framer.ReadMetaHeaders = h.decoder
	frame, err := framer.ReadFrame()
	if err != nil {
		h.broken = true
		return false
	}
	h.c.handleFrame(frame, h.isRequest)
	return true
}

type http2RetryCollector struct {
	collector Collector
	mu        sync.Mutex
	waiting   map[string]*http2RetryWaitState
}

type http2RetryWaitState struct {
	trace Trace
	stop  func()
}

func (h *http2RetryCollector) Complete(trace Trace) {
	if isRetryable(trace.Err) {
		// When the server refuses the stream, the client may auto-retry. So instead of marking
		// this trace as complete with this error, we will wait for a retry. If no retry occurs
		// after a timeout (3 seconds), then we will complete the trace. If the retry does occur,
		// we will ignore this erroneous trace.
		h.mu.Lock()
		defer h.mu.Unlock()
		if h.waiting == nil {
			h.waiting = map[string]*http2RetryWaitState{}
		}
		timer := time.AfterFunc(retryWait, func() {
			h.timesUp(trace.TestName)
		})
		h.waiting[trace.TestName] = &http2RetryWaitState{
			trace: trace,
			stop:  func() { timer.Stop() },
		}
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	_, alreadyInvoked := h.waiting[trace.TestName]
	if !alreadyInvoked {
		h.collector.Complete(trace)
	}
}

func (h *http2RetryCollector) newAttempt(testName string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if state, ok := h.waiting[testName]; ok {
		// This is a retry; cancel the pending wait task.
		delete(h.waiting, testName)
		state.stop()
	}
}

func (h *http2RetryCollector) timesUp(testName string) {
	h.mu.Lock()
	state, ok := h.waiting[testName]
	if ok {
		delete(h.waiting, testName)
	}
	h.mu.Unlock()

	if state == nil {
		// This entry has already been replaced by a retry.
		return
	}
	h.collector.Complete(state.trace)
}

func (h *http2RetryCollector) cancel() {
	// No more retries coming, so go ahead and finish
	// all operations that were waiting on a retry.
	var allWaiting []*http2RetryWaitState
	func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		allWaiting = make([]*http2RetryWaitState, 0, len(h.waiting))
		for _, state := range h.waiting {
			allWaiting = append(allWaiting, state)
		}
		h.waiting = nil
	}()

	for _, state := range allWaiting {
		state.stop() // release timer resources
		h.collector.Complete(state.trace)
	}
}

func makeRequest(frame *http2.MetaHeadersFrame) *http.Request {
	req := &http.Request{
		Proto:      "HTTP/2.0",
		ProtoMajor: 2,
		ProtoMinor: 0,
		URL: &url.URL{
			Scheme: getPseudoHeader(frame, ":scheme"),
			Host:   getPseudoHeader(frame, ":authority"),
			Path:   getPseudoHeader(frame, ":path"),
		},
		Method: getPseudoHeader(frame, ":method"),
		Header: makeHeaders(frame),
	}
	return req
}

func makeResponse(frame *http2.MetaHeadersFrame) *http.Response {
	status := getPseudoHeader(frame, ":status")
	var statusInt int
	if status == "" {
		statusInt = 500
	} else {
		var err error
		statusInt, err = strconv.Atoi(status)
		if err != nil {
			statusInt = 500
		}
	}
	return &http.Response{
		Proto:      "HTTP/2.0",
		ProtoMajor: 2,
		ProtoMinor: 0,
		StatusCode: statusInt,
		Status:     http.StatusText(statusInt),
		Header:     makeHeaders(frame),
	}
}

func makeHeaders(frame *http2.MetaHeadersFrame) http.Header {
	headers := make(http.Header, len(frame.Fields))
	for _, hdr := range frame.Fields {
		if strings.HasPrefix(hdr.Name, ":") {
			continue // http/2 pseudo-header
		}
		headers.Add(hdr.Name, hdr.Value)
	}
	return headers
}

func getPseudoHeader(frame *http2.MetaHeadersFrame, headerName string) string {
	for _, hdr := range frame.Fields {
		if hdr.Name == headerName {
			return hdr.Value
		}
	}
	return ""
}

func prefaceIsValid(actual []byte) bool {
	// We don't use bytes.Equal since clientPreface is a string, not []byte.
	// And we don't convert both to string or both to []byte to avoid allocation.
	if len(actual) != len(clientPreface) {
		return false
	}
	for i := range actual {
		if actual[i] != clientPreface[i] {
			return false
		}
	}
	return true
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	var streamErr http2.StreamError
	if errors.As(err, &streamErr) {
		// Retryable if server refused this individual stream
		return streamErr.Code == http2.ErrCodeRefusedStream
	}
	var connErr http2.ConnectionError
	if errors.As(err, &connErr) {
		// Retryable if server is performing graceful shutdown.
		return http2.ErrCode(connErr) == http2.ErrCodeNo
	}
	return false
}
