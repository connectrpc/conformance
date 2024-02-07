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

package compression

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"

	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/connect"
)

// The IANA names for supported compression algorithms.
const (
	Identity = "identity"
	Gzip     = "gzip"
	Brotli   = "br"
	Deflate  = "deflate"
	Snappy   = "snappy"
	Zstd     = "zstd"
)

// GetCompressor returns a compressor for the given compression algorithm.
func GetCompressor(compression conformancev1.Compression) (connect.Compressor, error) {
	switch compression {
	case conformancev1.Compression_COMPRESSION_UNSPECIFIED, conformancev1.Compression_COMPRESSION_IDENTITY:
		fmt.Fprintf(os.Stderr, "no op comp")
		return &noOpCompressor{}, nil
	case conformancev1.Compression_COMPRESSION_GZIP:
		return gzip.NewWriter(nil), nil
	case conformancev1.Compression_COMPRESSION_BR:
		return NewBrotliCompressor(), nil
	case conformancev1.Compression_COMPRESSION_ZSTD:
		return NewZstdCompressor(), nil
	case conformancev1.Compression_COMPRESSION_DEFLATE:
		return NewDeflateCompressor(), nil
	case conformancev1.Compression_COMPRESSION_SNAPPY:
		return NewSnappyCompressor(), nil
	default:
		return nil, fmt.Errorf("unsupported compression scheme %v", compression)
	}
}

// GetDecompressor returns a decompressor for the given compression algorithm.
func GetDecompressor(compression conformancev1.Compression) (connect.Decompressor, error) {
	switch compression {
	case conformancev1.Compression_COMPRESSION_UNSPECIFIED, conformancev1.Compression_COMPRESSION_IDENTITY:
		return &noOpDecompressor{}, nil
	case conformancev1.Compression_COMPRESSION_GZIP:
		return &gzip.Reader{}, nil
	case conformancev1.Compression_COMPRESSION_BR:
		return NewBrotliDecompressor(), nil
	case conformancev1.Compression_COMPRESSION_ZSTD:
		return NewZstdDecompressor(), nil
	case conformancev1.Compression_COMPRESSION_DEFLATE:
		return NewDeflateDecompressor(), nil
	case conformancev1.Compression_COMPRESSION_SNAPPY:
		return NewSnappyDecompressor(), nil
	default:
		return nil, fmt.Errorf("unsupported compression scheme %v", compression)
	}
}

type noOpCompressor struct {
	io.WriteCloser
}

func (c *noOpCompressor) Reset(writer io.Writer) {
	wc, ok := writer.(io.WriteCloser)
	if !ok {
		wc = &noOpCloser{writer}
	}
	c.WriteCloser = wc
}

type noOpDecompressor struct {
	io.ReadCloser
}

func (c *noOpDecompressor) Reset(reader io.Reader) error {
	rc, ok := reader.(io.ReadCloser)
	if !ok {
		rc = io.NopCloser(reader)
	}
	c.ReadCloser = rc
	return nil
}

type noOpCloser struct {
	io.Writer
}

func (n *noOpCloser) Close() error {
	return nil
}
