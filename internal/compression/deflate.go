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
	"compress/zlib"
	"errors"
	"io"

	"connectrpc.com/connect"
)

// deflateDecompressor is a thin wrapper around a zlib Reader. Note that due to
// an unfortunate misnomer with the RFC 2616 specification, HTTP deflate is
// actually RFC 1950 with zlib headers, rather than RFC 1951. gRPC uses the same
// nomenclature.
type deflateDecompressor struct {
	reader io.ReadCloser
}

func (c *deflateDecompressor) Read(bytes []byte) (int, error) {
	return c.reader.Read(bytes)
}
func (c *deflateDecompressor) Reset(rdr io.Reader) error {
	resetter, ok := c.reader.(zlib.Resetter)
	if !ok {
		// This should never happen as the returned type from flate should always
		// implement Resetter, but the check is here as a safeguard just in case.
		// This error would be a very exceptional / unexpected occurrence.
		return errors.New("deflate reader is not able to be used as a resetter")
	}
	// Mimics NewReader internal logic, which initializes the internal dict to nil
	return resetter.Reset(rdr, nil)
}
func (c *deflateDecompressor) Close() error {
	return c.reader.Close()
}

// NewDeflateDecompressor returns a new deflate Decompressor.
func NewDeflateDecompressor() connect.Decompressor {
	reader, err := zlib.NewReader(nil)
	if err != nil {
		return &errorDecompressor{err: err}
	}
	return &deflateDecompressor{
		reader: reader,
	}
}

// NewDeflateCompressor returns a new deflate Compressor.
func NewDeflateCompressor() connect.Compressor {
	// Construct a new zlib Writer with default compression level
	return zlib.NewWriter(nil)
}
