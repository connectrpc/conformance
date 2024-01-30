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
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"

	"connectrpc.com/conformance/internal/compression"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/connect"
)

const prefixLen = 5

type tracingReader struct {
	reader    io.ReadCloser
	builder   *builder
	isRequest bool
	whenDone  func()
	closed    atomic.Bool

	dataTracer dataTracer
}

func newRequestReader(headers http.Header, reader io.ReadCloser, isRequest bool, builder *builder) io.ReadCloser {
	// no action to take when request body is done
	whenDone := func() {}
	return newReader(headers, reader, isRequest, builder, whenDone)
}

func newReader(headers http.Header, reader io.ReadCloser, isRequest bool, builder *builder, whenDone func()) io.ReadCloser {
	isStream, decompressor := propertiesFromHeaders(headers)
	return &tracingReader{
		reader:    reader,
		isRequest: isRequest,
		builder:   builder,
		whenDone:  whenDone,
		dataTracer: dataTracer{
			isRequest:        isRequest,
			isStreamProtocol: isStream,
			decompressor:     decompressor,
			builder:          builder,
		},
	}
}

func (t *tracingReader) Read(data []byte) (n int, err error) {
	n, err = t.reader.Read(data)
	t.dataTracer.trace(data[:n])
	if err != nil {
		if errors.Is(err, io.EOF) {
			t.tryFinish(nil)
		} else {
			t.tryFinish(err)
		}
	}
	return n, err
}

func (t *tracingReader) Close() error {
	err := t.reader.Close()
	if err != nil {
		t.tryFinish(fmt.Errorf("close: %w", err))
	} else {
		t.tryFinish(errors.New("closed before fully consumed"))
	}
	return err
}

func (t *tracingReader) tryFinish(err error) {
	if !t.closed.CompareAndSwap(false, true) {
		return // already finished
	}
	defer t.whenDone()

	t.dataTracer.emitUnfinished()

	if t.isRequest {
		t.builder.add(&RequestBodyEnd{Err: err})
	} else {
		t.builder.add(&ResponseBodyEnd{Err: err})
	}
}

// dataTracer is responsible for translating bytes read/written into trace events.
type dataTracer struct {
	isRequest        bool
	isStreamProtocol bool
	decompressor     connect.Decompressor
	builder          *builder

	prefix    []byte
	env       *Envelope
	expecting uint32
	actual    uint64
	endStream *bytes.Buffer
}

func (d *dataTracer) trace(data []byte) {
	if !d.isStreamProtocol {
		d.actual += uint64(len(data))
		return
	}
	for {
		if len(data) == 0 {
			return
		}

		if d.expecting == 0 {
			// still reading envelope prefix
			n, done := d.tracePrefix(data)
			if !done {
				// need to read more data to finish prefix
				return
			}
			data = data[n:]
			continue
		}

		n, done := d.traceMessage(data)
		if !done {
			// need to read more data to finish message
			return
		}
		data = data[n:]
	}
}

func (d *dataTracer) tracePrefix(data []byte) (int, bool) {
	need := prefixLen - len(d.prefix)
	if len(data) < need {
		// envelope still not complete...
		d.prefix = append(d.prefix, data...)
		return need, false
	}

	d.prefix = append(d.prefix, data[:need]...)
	d.env = &Envelope{
		Len:   binary.BigEndian.Uint32(d.prefix[1:]),
		Flags: d.prefix[0],
	}
	d.expecting = d.env.Len
	d.prefix = d.prefix[:0]
	if d.expecting == 0 {
		// If we're not expecting any more data for this message, go
		// ahead and emit event.
		if d.isRequest {
			d.builder.add(&RequestBodyData{
				Envelope: d.env,
				Len:      0,
			})
		} else {
			d.builder.add(&ResponseBodyData{
				Envelope: d.env,
				Len:      0,
			})
		}
		d.env = nil
	} else if !d.isRequest && (d.env.Flags&0x82) != 0 {
		// This is a response end-stream message. Capture the contents.
		d.endStream = bytes.NewBuffer(make([]byte, 0, d.env.Len))
	}
	return need, true
}

func (d *dataTracer) traceMessage(data []byte) (int, bool) {
	need := int(d.expecting - uint32(d.actual))
	if len(data) < need {
		// message still not complete...
		d.actual += uint64(len(data))
		if d.endStream != nil {
			_, _ = d.endStream.Write(data)
		}
		return need, false
	}

	if d.isRequest {
		d.builder.add(&RequestBodyData{
			Envelope: d.env,
			Len:      uint64(d.expecting),
		})
	} else {
		d.builder.add(&ResponseBodyData{
			Envelope: d.env,
			Len:      uint64(d.expecting),
		})
	}
	if d.endStream != nil { //nolint:nestif
		_, _ = d.endStream.Write(data[:need])
		var content string
		if d.decompressor == nil {
			content = d.endStream.String()
		} else {
			var uncompressed bytes.Buffer
			if err := d.decompressor.Reset(d.endStream); err == nil {
				_, err := uncompressed.ReadFrom(d.decompressor)
				if err == nil {
					content = uncompressed.String()
				}
			}
		}
		if content != "" {
			d.builder.add(&ResponseBodyEndStream{
				Content: content,
			})
		}
		d.endStream = nil
	}
	d.env = nil
	d.expecting = 0
	d.actual = 0
	return need, true
}

func (d *dataTracer) emitUnfinished() {
	var unfinished uint64
	if d.expecting == 0 && len(d.prefix) > 0 {
		unfinished = uint64(len(d.prefix))
	} else {
		unfinished = d.actual
	}

	if unfinished > 0 {
		if d.isRequest {
			d.builder.add(&RequestBodyData{
				Envelope: d.env,
				Len:      unfinished,
			})
		} else {
			d.builder.add(&ResponseBodyData{
				Envelope: d.env,
				Len:      unfinished,
			})
		}
	}

	d.endStream = nil // we didn't finish reading end-stream message; discard what we got
	d.env = nil
	d.expecting = 0
	d.actual = 0
	d.prefix = d.prefix[:0]
}

// brokenDecompressor is a no-op implementation that treats all compressed
// messages as if they were empty.
type brokenDecompressor struct{}

func (brokenDecompressor) Read([]byte) (n int, err error) {
	return 0, io.EOF
}

func (brokenDecompressor) Close() error {
	return nil
}

func (brokenDecompressor) Reset(io.Reader) error {
	return nil
}

func propertiesFromHeaders(headers http.Header) (isStream bool, decomp connect.Decompressor) {
	contentType := strings.ToLower(headers.Get("Content-Type"))
	if headers.Get("Content-Encoding") != "" {
		// full body is encoded, so don't bother trying to parse stream
		return false, brokenDecompressor{}
	}
	switch {
	case strings.HasPrefix(contentType, "application/connect"):
		return true, getDecompressor(headers.Get("Connect-Content-Encoding"))
	case strings.HasPrefix(contentType, "application/grpc"):
		return true, getDecompressor(headers.Get("Grpc-Encoding"))
	default:
		// We should only need a decompressor for streams (to decompress the end-stream message)
		// So for non-stream protocols, this no-op decompressor should suffice.
		return false, brokenDecompressor{}
	}
}

func getDecompressor(encoding string) connect.Decompressor {
	var comp conformancev1.Compression
	switch strings.ToLower(encoding) {
	case "", "identity":
		comp = conformancev1.Compression_COMPRESSION_IDENTITY
	case "gzip":
		comp = conformancev1.Compression_COMPRESSION_GZIP
	case "br":
		comp = conformancev1.Compression_COMPRESSION_BR
	case "zstd":
		comp = conformancev1.Compression_COMPRESSION_ZSTD
	case "deflate":
		comp = conformancev1.Compression_COMPRESSION_DEFLATE
	case "snappy":
		comp = conformancev1.Compression_COMPRESSION_SNAPPY
	default:
		return brokenDecompressor{}
	}
	decomp, err := compression.GetDecompressor(comp)
	if err != nil {
		return brokenDecompressor{}
	}
	return decomp
}
