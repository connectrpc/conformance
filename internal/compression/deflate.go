// Copyright 2023 The Connect Authors
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
	"compress/flate"
	"errors"
	"io"

	"connectrpc.com/connect"
)

// deflateDecompressor is a thin wrapper around a flate Reader.
type deflateDecompressor struct {
	reader io.ReadCloser
}

func (c *deflateDecompressor) Read(bytes []byte) (int, error) {
	return c.reader.Read(bytes)
}
func (c *deflateDecompressor) Reset(rdr io.Reader) error {
	resetter, ok := c.reader.(flate.Resetter)
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
	return &deflateDecompressor{
		reader: flate.NewReader(nil),
	}
}

// NewDeflateCompressor returns a new deflate Compressor.
func NewDeflateCompressor() connect.Compressor {
	// Construct a new flate Writer with default compression level
	w, err := flate.NewWriter(nil, 1)
	if err != nil {
		return &errorCompressor{err: err}
	}
	return w
}
