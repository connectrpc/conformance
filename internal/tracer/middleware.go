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
	"context"
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
		ctx, cancel := context.WithCancel(req.Context())
		go func() {
			<-ctx.Done()
			builder.add(&RequestCanceled{})
		}()
		req = req.Clone(ctx)
		req.Body = newRequestReader(req.Header, req.Body, true, builder)
		resp, err := transport.RoundTrip(req)
		if err != nil {
			builder.add(&ResponseError{Err: err})
			cancel()
			return nil, err
		}
		// body, _ := io.ReadAll(resp.Body)
		// fmt.Fprintf(os.Stderr, "dagsboro: ", string(body))
		// resp.Body = io.NopCloser(bytes.NewReader(body))
		// booty, _ := io.ReadAll(resp.Body)
		// fmt.Fprintf(os.Stderr, "dagster.io: ", string(booty))

		respect := *resp
		cardi, _ := io.ReadAll(respect.Body)
		respect.Body = io.NopCloser(bytes.NewReader(cardi))

		// bardy, _ := io.ReadAll(evt.Response.Body)
		// fmt.Fprintf(os.Stderr, "dagestan", string(bardy))

		builder.add(&ResponseStart{
			Response: &respect,
		})
		respClone := *resp

		respClone.Body = newReader(resp.Header, resp.Body, false, builder, cancel)
		return &respClone, nil
	})
}

// TracingHandler applies tracing middleware to the given handler. The returned
// handler will record traces of all operations to the given tracer.
func TracingHandler(handler http.Handler, collector Collector) http.Handler {
	return http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		builder := newBuilder(req, collector)
		defer builder.build() // make sure the trace is complete before returning
		go func() {
			<-ctx.Done()
			builder.add(&RequestCanceled{})
		}()

		req = req.Clone(ctx)
		req.Body = newRequestReader(req.Header, req.Body, true, builder)
		traceWriter := &tracingResponseWriter{
			respWriter: respWriter,
			req:        req,
			builder:    builder,
		}
		defer func() {
			var err error
			panicVal := recover()
			if panicVal != nil {
				err = fmt.Errorf("panic: %v", panicVal)
			}
			traceWriter.tryFinish(err)
			if panicVal != nil {
				//nolint:forbidigo // just propagating existing panic
				panic(panicVal)
			}
		}()

		handler.ServeHTTP(
			traceWriter,
			req,
		)
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
		Body:          io.NopCloser(bytes.NewBuffer(nil)), // empty body
		Status:        fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		StatusCode:    statusCode,
		Proto:         t.req.Proto,
		ProtoMajor:    t.req.ProtoMajor,
		ProtoMinor:    t.req.ProtoMinor,
		ContentLength: contentLen,
		TLS:           t.req.TLS,
	}
	// Create snapshot of the headers
	t.resp.Header = make(http.Header, len(t.Header()))
	for headerName, headerVals := range t.Header() {
		if strings.HasPrefix(headerName, http.TrailerPrefix) {
			continue
		}
		headerVals = append([]string(nil), headerVals...) // snapshot the slice
		t.resp.Header[headerName] = headerVals
	}
	// And also seed the trailers with expected trailer keys
	trailerHeaders := t.Header().Values("Trailer")
	t.resp.Trailer = make(http.Header, len(trailerHeaders))
	for _, trailerNames := range trailerHeaders {
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
	t.setTrailers()
	t.builder.add(&ResponseBodyEnd{Err: err})
}

func (t *tracingResponseWriter) setTrailers() {
	headersAndTrailers := t.Header() // response writer's headers (counter-intuitively) include trailers, too
	// First extract any known trailer keys (that were advertised in "Trailer" header).
	for trailerName := range t.resp.Trailer {
		t.resp.Trailer[trailerName] = headersAndTrailers[trailerName]
	}
	// Then get any others, identified by a special prefix in the name.
	for key, vals := range headersAndTrailers {
		trailerKey := strings.TrimPrefix(key, http.TrailerPrefix)
		if trailerKey == key {
			// no prefix trimmed, so not a trailer
			continue
		}
		existing := t.resp.Trailer[trailerKey]
		if len(existing) > 0 {
			// defensive copy, so we don't accidentally mutate the slice in response writer's headers
			existing = append([]string(nil), existing...)
		}
		t.resp.Trailer[trailerKey] = append(existing, vals...)
	}
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (r roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return r(req)
}
