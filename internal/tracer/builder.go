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
	"net/http"
	"sync"
	"time"
)

// builder accumulates events to build a trace.
type builder struct {
	collector Collector
	start     time.Time

	mu                  sync.Mutex
	trace               Trace
	reqCount, respCount int
}

// newBuilder creates a new builder for the given HTTP operation. The
// returned builder will already have a RequestStart event, based on
// the given request, so callers should NOT explicitly call builder.add
// to add such an event.
func newBuilder(req *http.Request, collector Collector) *builder {
	testName := req.Header.Get("x-test-case-name")
	return &builder{
		collector: collector,
		start:     time.Now(),
		trace: Trace{
			TestName: testName,
			Events:   []Event{&RequestStart{Request: req}},
		},
	}
}

// add adds the given event to the trace being built.
func (b *builder) add(event Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.trace.TestName == "" {
		return
	}
	switch e := event.(type) {
	case *ResponseStart:
		b.trace.Response = e.Response
	case *ResponseError:
		b.trace.Err = e.Err
	case *RequestBodyEnd:
		if b.trace.Err != nil {
			b.trace.Err = e.Err
		}
	case *ResponseBodyEnd:
		if b.trace.Err != nil {
			b.trace.Err = e.Err
		}
	case *RequestBodyData:
		e.MessageIndex = b.reqCount
		b.reqCount++
	case *ResponseBodyData:
		e.MessageIndex = b.respCount
		b.respCount++
	}
	event.setEventOffset(time.Since(b.start))
	b.trace.Events = append(b.trace.Events, event)
}

// build builds the trace and provides the data to the given Tracer.
func (b *builder) build() {
	b.mu.Lock()
	trace := b.trace
	b.trace = Trace{} // reset; subsequent calls to add or build ignored
	b.mu.Unlock()

	if trace.TestName == "" {
		return
	}
	b.collector.Complete(trace)
}
