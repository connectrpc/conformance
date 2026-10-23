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
	"fmt"
	"io"

	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connectgzip"
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

// All returns a compressor for every supported algorithm other than identity.
// Servers advertise them in this order, which matches earlier releases.
func All() []connect.Compressor {
	return []connect.Compressor{
		NewZstdCompressor(),
		NewSnappyCompressor(),
		NewDeflateCompressor(),
		NewBrotliCompressor(),
		connectgzip.New(),
	}
}

// GetCompressor returns a compressor for the given compression algorithm.
func GetCompressor(compression conformancev1.Compression) (connect.Compressor, error) {
	switch compression {
	case conformancev1.Compression_COMPRESSION_UNSPECIFIED, conformancev1.Compression_COMPRESSION_IDENTITY:
		return noOpCompressor{}, nil
	case conformancev1.Compression_COMPRESSION_GZIP:
		return connectgzip.New(), nil
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
// A [connect.Compressor] also decompresses, so this is the same as
// [GetCompressor].
func GetDecompressor(compression conformancev1.Compression) (connect.Compressor, error) {
	return GetCompressor(compression)
}

type noOpCompressor struct{}

func (noOpCompressor) Name() string { return Identity }

func (noOpCompressor) Compress(dst io.Writer) (io.WriteCloser, error) {
	return nopWriteCloser{dst}, nil
}

func (noOpCompressor) Decompress(src io.Reader) (io.ReadCloser, error) {
	return io.NopCloser(src), nil
}

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error { return nil }
