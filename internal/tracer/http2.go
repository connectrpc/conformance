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
	"fmt"
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
	// then we can still provide a trace (for the refused attempt) in the time window.
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

// TracingHTTP2Listener applies tracing to the given net.Listener, which
// accepts sockets that use the HTTP/2 protocol. If TLS is used, it should
// already be applied, so the given listener should return connections that
// can be used for reading/writing clear-text (i.e. pre-encryption,
// post-decryption).
//
// This function calls TracingHTTP2Conn for each connection accepted. All
// connections accepted are considered server connections.
func TracingHTTP2Listener(listener net.Listener, collector Collector) net.Listener {
	return &tracingListener{Listener: listener, collector: collector}
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

func (c *tracingHTTP2Conn) Read(data []byte) (n int, err error) {
	n, err = c.Conn.Read(data)
	c.readTracer.trace(data[:n])

	if err != nil {
		// We ignore timeout errors because the HTTP/2 server
		// does some tricks with read timeouts on new connections
		// related to reading the preface and settings frames.
		// So a single read operation may fail, but the connection
		// continues to be used.
		var netErr net.Error
		isTimeout := errors.As(err, &netErr) && netErr.Timeout()
		if !isTimeout {
			c.cancelAll(err)
		}
	}
	return n, err
}

func (c *tracingHTTP2Conn) Write(data []byte) (n int, err error) {
	// Note: we trace the given data as if it were all successfully
	// written, before actually writing (and seeing how much is written)
	// to avoid a race condition where the peer could be processing the
	// data we wrote and then reply and have its reply processed by some
	// other goroutine, all while we are concurrently capturing the trace.
	// That leads to strange non-deterministic issues with event ordering.
	c.writeTracer.trace(data)
	n, err = c.Conn.Write(data)
	if err != nil {
		c.cancelAll(err)
	}
	return n, err
}

func (c *tracingHTTP2Conn) Close() error {
	err := c.Conn.Close()
	if err == nil {
		c.cancelAll(errors.New("socket closed"))
	} else {
		c.cancelAll(fmt.Errorf("socket closed; %w", err))
	}
	return err
}

func (c *tracingHTTP2Conn) handleFrame(frame http2.Frame, isRequest bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch frame := frame.(type) {
	case *http2.MetaHeadersFrame:
		stream, isNew := c.getStreamLocked(frame, isRequest)
		if stream == nil {
			return
		}
		switch {
		case isNew:
			// request headers, which resulted in a new stream; nothing else to do
		case !isRequest && !stream.gotResponse:
			// response headers
			c.receiveResponseLocked(stream, frame)
		case isRequest:
			// request trailers
			stream.builder.trace.Request.Trailer = makeHeaders(frame)
		default:
			// response trailers
			stream.builder.trace.Response.Trailer = makeHeaders(frame)
		}
		if frame.StreamEnded() {
			c.closeStreamLocked(frame.StreamID, stream, isRequest, nil)
		}
	case *http2.DataFrame:
		stream := c.getExistingStreamLocked(frame.StreamID)
		if stream == nil {
			return
		}
		if isRequest {
			stream.requestTracer.trace(frame.Data())
		} else {
			stream.responseTracer.trace(frame.Data())
		}
		if frame.StreamEnded() {
			c.closeStreamLocked(frame.StreamID, stream, isRequest, nil)
		}
	case *http2.RSTStreamFrame:
		stream := c.getExistingStreamLocked(frame.StreamID)
		if stream == nil {
			return
		}
		c.closeStreamLocked(frame.StreamID, stream, isRequest, http2.StreamError{
			StreamID: frame.StreamID,
			Code:     frame.ErrCode,
		})

	case *http2.GoAwayFrame:
		c.setMaxStreamIDLocked(frame.LastStreamID, http2.ConnectionError(frame.ErrCode))
	}
}

func (c *tracingHTTP2Conn) receiveResponseLocked(stream *http2Stream, frame *http2.MetaHeadersFrame) {
	stream.gotResponse = true
	resp := makeResponse(frame) //nolint:bodyclose // there is no body to close on this response
	stream.builder.add(&ResponseStart{Response: resp})
	stream.responseTracer.isStreamProtocol, stream.responseTracer.decompressor = propertiesFromHeaders(resp.Header)
	stream.responseTracer.builder = stream.builder
}

func (c *tracingHTTP2Conn) getStreamLocked(frame *http2.MetaHeadersFrame, createIfNotFound bool) (*http2Stream, bool) {
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

func (c *tracingHTTP2Conn) getExistingStreamLocked(streamID uint32) *http2Stream {
	return c.streams[streamID]
}

func (c *tracingHTTP2Conn) closeStreamLocked(streamID uint32, stream *http2Stream, isRequest bool, err error) {
	if !isRequest || err != nil {
		// This is either end of response or an error, which means the
		// whole operation done.
		delete(c.streams, streamID)
	}
	if isRequest {
		stream.requestTracer.emitUnfinished()
		stream.builder.add(&RequestBodyEnd{Err: err})
	} else if stream.responseTracer.builder != nil {
		stream.requestTracer.emitUnfinished()
		stream.responseTracer.emitUnfinished()
		stream.builder.add(&ResponseBodyEnd{Err: err})
	}
}

func (c *tracingHTTP2Conn) setMaxStreamIDLocked(maxStreamID uint32, err error) {
	c.maxStreamID = maxStreamID
	for streamID, stream := range c.streams {
		if streamID > maxStreamID {
			delete(c.streams, streamID)
			stream.requestTracer.emitUnfinished()
			stream.responseTracer.emitUnfinished()
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
				stream.requestTracer.emitUnfinished()
				stream.responseTracer.emitUnfinished()
				stream.builder.add(&ResponseBodyEnd{Err: err})
			} else {
				// TODO: We shouldn't add RequestBodyEnd event if the trace
				//       already has an event of that type.
				stream.requestTracer.emitUnfinished()
				stream.builder.add(&RequestBodyEnd{Err: err})
				stream.builder.add(&RequestCanceled{})
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

type tracingListener struct {
	net.Listener
	collector Collector
}

func (t *tracingListener) Accept() (net.Conn, error) {
	conn, err := t.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return TracingHTTP2Conn(conn, true, t.collector), nil
}

type http2RetryWaitState struct {
	trace Trace
	stop  func()
}

type http2RetryCollector struct {
	collector Collector
	mu        sync.Mutex
	waiting   map[string]*http2RetryWaitState
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
	path := getPseudoHeader(frame, ":path")
	var query string
	var forceQuery bool
	if strings.Contains(path, "?") {
		// There's a query string.
		parts := strings.SplitN(path, "?", 2)
		path, query = parts[0], parts[1]
		forceQuery = query == ""
	}
	req := &http.Request{
		Proto:      "HTTP/2.0",
		ProtoMajor: 2,
		ProtoMinor: 0,
		URL: &url.URL{
			Scheme:     getPseudoHeader(frame, ":scheme"),
			Host:       getPseudoHeader(frame, ":authority"),
			Path:       path,
			RawQuery:   query,
			ForceQuery: forceQuery,
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
		Status:     fmt.Sprintf("%d %s", statusInt, http.StatusText(statusInt)),
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
