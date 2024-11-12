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
	"net/http"
	"net/http/httptrace"
	"net/textproto"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

const testCaseNameHeader = "X-Test-Case-Name"

// builder accumulates events to build a trace.
type builder struct {
	collector Collector
	start     time.Time
	client    bool

	mu                  sync.Mutex
	trace               Trace
	reqCount, respCount int
}

// newBuilder creates a new builder for the given HTTP operation. The
// returned builder will already have a RequestStart event, based on
// the given request, so callers should NOT explicitly call builder.add
// to add such an event.
func newBuilder(req *http.Request, client bool, collector Collector) (*builder, context.Context) {
	ctx := req.Context()
	var getHeaders func() http.Header
	if client {
		// The net/http transport may *add* other headers to the request
		// on the wire. In order to see them, we need to use httptrace
		// instead of looking at the headers on req.
		var headersMu sync.Mutex
		headers := http.Header{}
		trace := httptrace.ClientTrace{
			WroteHeaderField: func(key string, value []string) {
				if strings.HasPrefix(key, ":") {
					// ignore http/2 pseudo-headers
					return
				}
				headersMu.Lock()
				defer headersMu.Unlock()
				vals := make([]string, len(value)) // defensive copy
				copy(vals, value)
				headers[textproto.CanonicalMIMEHeaderKey(key)] = vals
			},
		}
		ctx = httptrace.WithClientTrace(ctx, &trace)
		getHeaders = func() http.Header {
			headersMu.Lock()
			defer headersMu.Unlock()
			return headers.Clone()
		}
	} else {
		// If req.ContentLength is set by net/http server, it must have come
		// from a header. So synthesize the header if it's not present.
		headers := req.Header.Clone()
		if len(headers.Get("Content-Length")) == 0 && req.ContentLength != -1 {
			headers.Set("Content-Length", strconv.FormatInt(req.ContentLength, 10))
		}
		getHeaders = func() http.Header {
			return headers
		}
	}
	testName := req.Header.Get(testCaseNameHeader)
	return &builder{
		collector: collector,
		start:     time.Now(),
		client:    client,
		trace: Trace{
			TestName: testName,
			Request:  req,
			Events:   []Event{&RequestStart{Request: req, getHeaders: getHeaders}},
		},
	}, ctx
}

// add adds the given event to the trace being built.
func (b *builder) add(event Event) {
	var finish bool
	var finishedTrace Trace
	defer func() {
		// We must not call b.collector.Complete  while lock held,
		// So we do so from deferred function, after lock released.
		if finish {
			b.finish(finishedTrace)
		}
	}()
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.trace.TestName == "" {
		return
	}
	switch event := event.(type) {
	case *RequestBodyData:
		event.MessageIndex = b.reqCount
		b.reqCount++
	case *RequestBodyEnd:
		if b.trace.Err == nil {
			b.trace.Err = event.Err
		}
		if event.Err != nil {
			// An error writing the request body means
			// operation has failed.
			finish = true
		}
	case *ResponseStart:
		b.trace.Response = event.Response
		if b.client {
			// for client-side traces, the HTTP version of the request
			// isn't known until we get back the response
			b.trace.Request.Proto = event.Response.Proto
			b.trace.Request.ProtoMajor = event.Response.ProtoMajor
			b.trace.Request.ProtoMinor = event.Response.ProtoMinor
		}
	case *ResponseError:
		b.trace.Err = event.Err
		if b.client {
			// Can't use type assertion to http2.StreamError because the standard library
			// includes vendored copy of the http2 package. So the type assertion would fail
			// since the type in the vendored package != the type in x/net/http2.
			if strings.HasSuffix(reflect.TypeOf(event.Err).String(), "http2.StreamError") {
				b.trace.Request.Proto = "HTTP/2.0"
			} else {
				// We don't conclusively know what version of HTTP was used.
				b.trace.Request.Proto = ""
			}
		}
		finish = true
	case *ResponseBodyData:
		event.MessageIndex = b.respCount
		b.respCount++
	case *ResponseBodyEnd:
		if b.trace.Err == nil {
			b.trace.Err = event.Err
		}
		finish = true
	case *RequestCanceled:
		if b.trace.Err == nil {
			b.trace.Err = context.Canceled
		}
		finish = true
	}
	event.setEventOffset(time.Since(b.start))
	b.trace.Events = append(b.trace.Events, event)
	if finish {
		finishedTrace = b.getAndClearLocked()
	}
}

func (b *builder) getAndClearLocked() Trace {
	trace := b.trace
	b.trace = Trace{} // reset; subsequent calls to add or build ignored
	return trace
}

func (b *builder) finish(trace Trace) {
	if trace.TestName != "" {
		b.collector.Complete(trace)
	}
}

// build builds the trace and provides the data to the given Tracer.
func (b *builder) build() {
	b.mu.Lock()
	trace := b.getAndClearLocked()
	b.mu.Unlock()

	b.finish(trace)
}
