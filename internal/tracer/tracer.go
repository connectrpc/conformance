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
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"connectrpc.com/conformance/internal"
)

const (
	requestPrefix  = " request>"
	responsePrefix = "response<"
)

// Tracer stores traces as they are produced and makes them available to a consumer.
// Each operation, identified by a test name, must first be initialized by the consumer
// via Init. The producer then populates the information for that operation via Complete.
// The consumer can then use Await to retrieve the trace (which may be produced
// asynchronously) and should finally use Clear, to free up resources associated with
// the operation. (If Clear is never called, the Tracer will use more and more memory,
// but limited by the amount to store all traces for every operation traced.)
type Tracer struct {
	mu     sync.Mutex
	traces map[string]*traceResult
}

// Init initializes the tracer to accept data for a trace for the given test name.
// This must be called before Clear, Complete, or Await for the same name.
func (t *Tracer) Init(testName string) {
	if t == nil {
		return
	}
	var result traceResult
	result.done = make(chan struct{})
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.traces == nil {
		t.traces = map[string]*traceResult{}
	}
	t.traces[testName] = &result
}

// Clear clears the data for the given test name. This frees up resources so
// that the tracer doesn't use more memory than necessary.
func (t *Tracer) Clear(testName string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.traces, testName)
}

// Complete marks a test as complete with the given trace data. If Clear
// has already been called or Init was never called, this does nothing.
func (t *Tracer) Complete(trace Trace) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	result := t.traces[trace.TestName]
	if result == nil || result.done == nil {
		return
	}
	done := result.done
	result.trace = trace
	result.done = nil
	close(done)
}

// Await waits for the given test to complete and for its trace data to
// become available. It returns a context error if the given context is
// cancelled or its deadline is reached before completion. It also returns
// an error if Clear has alreadu been called for the test or if Init was
// never called.
func (t *Tracer) Await(ctx context.Context, testName string) (*Trace, error) {
	if t == nil {
		return nil, fmt.Errorf("%s: tracing not enabled", testName)
	}
	t.mu.Lock()
	result := t.traces[testName]
	var done chan struct{}
	if result != nil {
		done = result.done
	}
	t.mu.Unlock()
	if result == nil {
		return nil, fmt.Errorf("%s: trace already cleared", testName)
	}
	if done == nil {
		return &result.trace, nil
	}
	select {
	case <-done:
		return &result.trace, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Trace represents the sequence of activity for a single HTTP operation.
type Trace struct {
	TestName string
	Request  *http.Request
	Response *http.Response
	Err      error
	Events   []Event
}

func (t *Trace) Print(printer internal.Printer) {
	for _, event := range t.Events {
		event.print(printer)
	}
	if t.Response != nil && len(t.Response.Trailer) > 0 {
		printer.Printf(responsePrefix)
		printHeaders(responsePrefix, t.Response.Trailer, printer)
	}
}

// Collector is a consumer of traces. This is usually an
// instance of *Tracer, but is an interface so that the implementation
// can vary, even allowing decorating or intercepting the method on
// *Tracer.
type Collector interface {
	// Complete accepts a trace once it is completed.
	Complete(Trace)
}

var _ Collector = (*Tracer)(nil)

// Event is a single item in a sequence of activity for an HTTP operation.
type Event interface {
	setEventOffset(time.Duration)
	print(internal.Printer)
}

// Envelope represents the metadata about an enveloped message in an
// RPC stream. Streaming protocols prefix each message with this
// metadata.
type Envelope struct {
	Flags byte
	Len   uint32
}

// RequestStart is an event that represents when the request starts. This
// is recorded when the client sends the request or when the server
// receives it. This is always the first event for an HTTP operation.
type RequestStart struct {
	Request *http.Request

	eventOffset
}

func (r *RequestStart) print(printer internal.Printer) {
	urlClone := *r.Request.URL
	if urlClone.Host == "" {
		urlClone.Host = "..."
	}
	if r.Request.TLS != nil {
		urlClone.Scheme = "https"
	} else {
		urlClone.Scheme = "http"
	}
	printer.Printf("%s %9.3fms %s %s %s", requestPrefix, r.offsetMillis(), r.Request.Method, urlClone.String(), r.Request.Proto)
	printHeaders(requestPrefix, r.Request.Header, printer)
	if r.Request.ContentLength != -1 && len(r.Request.Header.Values("Content-Length")) == 0 {
		printer.Printf("%s %11s Content-Length: %d", requestPrefix, "", r.Request.ContentLength)
	}
	printer.Printf(requestPrefix)
}

// RequestBodyData represents some data written to or read from the
// request body. These operations are "chunked" so that a single event
// represents a full message (or incomplete, partial message if a full
// message is not written or read).
type RequestBodyData struct {
	// For streaming protocols, each message is
	// enveloped and this should be non-nil. It may
	// be nil in a streaming protocol if an envelope
	// prefix was expected, but only a partial prefix
	// could be written/read. In such a case, a
	// RequestBodyData event is emitted that has no
	// envelope and whose Len field indicates the
	// number of bytes written/read of the incomplete
	// prefix.
	Envelope *Envelope
	// Actual length of the data, which could differ
	// from the length indicated in the envelope if
	// the full message could not be written/read.
	Len uint64

	// Sequentially numbered index. The first message
	// in the stream should have an index of zero, and
	// then one, etc.
	MessageIndex int

	eventOffset
}

func (r *RequestBodyData) print(printer internal.Printer) {
	printData(requestPrefix, r.offsetMillis(), r.MessageIndex, r.Envelope, r.Len, printer)
}

// RequestBodyEnd represents the end of the request body being reached.
// The Err value is the error returned from the final read (on the server)
// or call to close the body (on the client). If the final read returned
// io.EOF, Err will be nil. So a non-nil Err means an abnormal conclusion
// to the operation. No more request events will appear after this.
type RequestBodyEnd struct {
	Err error

	eventOffset
}

func (r *RequestBodyEnd) print(printer internal.Printer) {
	if r.Err != nil {
		printer.Printf("%s %9.3fms body end (err=%v)", requestPrefix, r.offsetMillis(), r.Err)
	} else {
		printer.Printf("%s %9.3fms body end", requestPrefix, r.offsetMillis())
	}
}

// ResponseStart is an event that represents when the response starts. This
// is recorded when the client receives the response headers or when the
// server sends them. This will precede all other response events.
type ResponseStart struct {
	Response *http.Response

	eventOffset
}

func (r *ResponseStart) print(printer internal.Printer) {
	printer.Printf("%s %9.3fms %s", responsePrefix, r.offsetMillis(), r.Response.Status)
	printHeaders(responsePrefix, r.Response.Header, printer)
	if r.Response.ContentLength != -1 && len(r.Response.Header.Values("Content-Length")) == 0 {
		printer.Printf("%s %11s Content-Length: %d", responsePrefix, "", r.Response.ContentLength)
	}
	printer.Printf(responsePrefix)
}

// ResponseError is an event that represents when the response fails. This
// is recorded when the client receives an error instead of a response, like
// due to a network error. No more events will appear after this.
type ResponseError struct {
	Err error

	eventOffset
}

func (r *ResponseError) print(printer internal.Printer) {
	printer.Printf("%s %9.3fms failed: %v", responsePrefix, r.offsetMillis(), r.Err)
}

// ResponseBodyData represents some data written to or read from the
// response body. These operations are "chunked" so that a single event
// represents a full message (or incomplete, partial message if a full
// message is not written or read).
type ResponseBodyData struct {
	// For streaming protocols, each message is
	// enveloped and this should be non-nil. It may
	// be nil in a streaming protocol if an envelope
	// prefix was expected, but only a partial prefix
	// could be written/read. In such a case, a
	// ResponseBodyData event is emitted that has no
	// envelope and whose Len field indicates the
	// number of bytes written/read of the incomplete
	// prefix.
	Envelope *Envelope
	// Actual length of the data, which could differ
	// from the length indicated in the envelope if
	// the full message could not be written/read.
	Len uint64

	// Sequentially numbered index. The first message
	// in the stream should have an index of zero, and
	// then one, etc.
	MessageIndex int

	eventOffset
}

func (r *ResponseBodyData) print(printer internal.Printer) {
	printData(responsePrefix, r.offsetMillis(), r.MessageIndex, r.Envelope, r.Len, printer)
}

// ResponseBodyEndStream represents the an "end-stream" message in the
// Connect streaming and gRPC-Web protocols. It is a special representation
// of the operation's status and trailers that is part of the response
// body.
type ResponseBodyEndStream struct {
	Content string

	eventOffset
}

func (r *ResponseBodyEndStream) print(printer internal.Printer) {
	lines := strings.Split(r.Content, "\n")
	for _, line := range lines {
		line = strings.Trim(line, "\r")
		printer.Printf("%s %11s   eos: %s", responsePrefix, "", line)
	}
}

// ResponseBodyEnd represents the end of the response body being reached.
// The Err value is the error returned from the final read (on the client)
// or final write (on the server). If the final read returned io.EOF, Err
// will be nil. So a non-nil Err means an abnormal conclusion to the
// operation. No more events will appear after this.
type ResponseBodyEnd struct {
	Err error

	eventOffset
}

func (r *ResponseBodyEnd) print(printer internal.Printer) {
	if r.Err != nil {
		printer.Printf("%s %9.3fms body end (err=%v)", responsePrefix, r.offsetMillis(), r.Err)
	} else {
		printer.Printf("%s %9.3fms body end", responsePrefix, r.offsetMillis())
	}
}

// RequestCanceled represents the instant the request is cancelled by the
// client. No more events will appear after this.
type RequestCanceled struct {
	eventOffset
}

func (r *RequestCanceled) print(printer internal.Printer) {
	printer.Printf("%s %9.3fms canceled", requestPrefix, r.offsetMillis())
}

type traceResult struct {
	trace Trace
	done  chan struct{}
}

type eventOffset struct {
	Offset time.Duration
}

func (o *eventOffset) setEventOffset(offset time.Duration) {
	o.Offset = offset
}

func (o *eventOffset) offsetMillis() float64 {
	return o.Offset.Seconds() * 1000
}

func printHeaders(prefix string, headers http.Header, printer internal.Printer) {
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		for _, val := range headers[key] {
			printer.Printf("%s %11s %s: %s", prefix, "", key, val)
		}
	}
}

func printData(prefix string, offsetMillis float64, index int, env *Envelope, length uint64, printer internal.Printer) {
	if env != nil {
		printer.Printf("%s %9.3fms message #%d: prefix: flags=%d, len=%d", prefix, offsetMillis, index+1, env.Flags, env.Len)
		if length > 0 {
			printer.Printf("%s %11s message #%d: data: %d/%d bytes", prefix, "", index+1, length, env.Len)
		}
	} else {
		printer.Printf("%s %9.3fms message #%d: data: %d bytes", prefix, offsetMillis, index+1, length)
	}
}
