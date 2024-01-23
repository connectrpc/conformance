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
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// TracingRoundTripper applies tracing to the given transport. The returned
// round tripper will record traces of all operations to the given tracer.
func TracingRoundTripper(transport http.RoundTripper, collector Collector) http.RoundTripper {
	return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		builder := newBuilder(req, collector)
		req = req.Clone(req.Context())
		req.Body = newReader(req.Header, req.Body, true, builder)
		resp, err := transport.RoundTrip(req)
		if err != nil {
			builder.add(&ResponseError{Err: err})
			builder.build()
			return nil, err
		}
		builder.add(&ResponseStart{Response: resp})
		respClone := *resp
		respClone.Body = newReader(resp.Header, resp.Body, false, builder)
		return &respClone, nil
	})
}

// TracingHandler applies tracing middleware to the given handler. The returned
// handler will record traces of all operations to the given tracer.
func TracingHandler(handler http.Handler, collector Collector) http.Handler {
	return http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
		builder := newBuilder(req, collector)
		req = req.Clone(req.Context())
		req.Body = newReader(req.Header, req.Body, true, builder)
		traceWriter := &tracingResponseWriter{
			respWriter: respWriter,
			req:        req,
			builder:    builder,
		}

		handler.ServeHTTP(
			traceWriter,
			req,
		)

		traceWriter.tryFinish(nil)
	})
}

type tracingResponseWriter struct {
	respWriter http.ResponseWriter
	req        *http.Request
	builder    *builder
	started    bool
	resp       *http.Response
	finished   bool

	dataTracer dataTracer
}

func (t *tracingResponseWriter) Unwrap() http.ResponseWriter {
	return t.respWriter
}

func (t *tracingResponseWriter) Header() http.Header {
	return t.respWriter.Header()
}

func (t *tracingResponseWriter) Write(data []byte) (int, error) {
	if !t.started {
		t.WriteHeader(http.StatusOK)
	}
	n, err := t.respWriter.Write(data)
	t.dataTracer.trace(data[:n])
	if err != nil {
		t.tryFinish(err)
	}
	return n, err
}

func (t *tracingResponseWriter) WriteHeader(statusCode int) {
	if t.started {
		return
	}
	t.started = true
	t.respWriter.WriteHeader(statusCode)
	isStreamProtocol, decompressor := propertiesFromHeaders(t.Header())
	t.dataTracer = dataTracer{
		isRequest:        false,
		isStreamProtocol: isStreamProtocol,
		decompressor:     decompressor,
		builder:          t.builder,
	}
	contentLenStr := t.Header().Get("Content-Length")
	contentLen := int64(-1)
	if contentLenStr != "" {
		if intVal, err := strconv.ParseInt(contentLenStr, 10, 64); err == nil {
			contentLen = intVal
		}
	}
	t.resp = &http.Response{
		Header:        t.respWriter.Header(),
		Body:          io.NopCloser(bytes.NewBuffer(nil)), // empty body
		Status:        fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		StatusCode:    statusCode,
		Proto:         t.req.Proto,
		ProtoMajor:    t.req.ProtoMajor,
		ProtoMinor:    t.req.ProtoMinor,
		ContentLength: contentLen,
		TLS:           t.req.TLS,
		Trailer:       http.Header{},
	}
	for _, trailerNames := range t.Header().Values("Trailer") {
		for _, trailerName := range strings.Split(trailerNames, ",") {
			trailerName = strings.TrimSpace(trailerName)
			if trailerName == "" {
				continue
			}
			t.resp.Trailer[trailerName] = nil
		}
	}
	t.builder.add(&ResponseStart{Response: t.resp})
}

func (t *tracingResponseWriter) Flush() {
	flusher, ok := t.respWriter.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}

func (t *tracingResponseWriter) tryFinish(err error) {
	if t.finished {
		return // already finished
	}
	if !t.started {
		t.WriteHeader(http.StatusOK)
	}

	t.finished = true
	t.dataTracer.emitUnfinished()
	t.builder.add(&ResponseBodyEnd{Err: err})
	t.setTrailers()
	t.builder.build()
}

func (t *tracingResponseWriter) setTrailers() {
	for trailerName := range t.resp.Trailer {
		t.resp.Trailer[trailerName] = t.resp.Header[trailerName]
	}
	for key, vals := range t.resp.Header {
		trailerKey := strings.TrimPrefix(key, http.TrailerPrefix)
		if trailerKey == key {
			// no prefix trimmed, so not a trailer
			continue
		}
		existing := t.resp.Trailer[trailerKey]
		t.resp.Trailer[trailerKey] = append(existing, vals...)
		delete(t.resp.Header, key)
	}
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (r roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return r(req)
}
