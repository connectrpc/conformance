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
	"io"

	"connectrpc.com/connect"
	"github.com/andybalholm/brotli"
)

// brotliDecompressor is a thin wrapper around a brotli Reader.
type brotliDecompressor struct {
	reader *brotli.Reader
}

func (c *brotliDecompressor) Read(bytes []byte) (int, error) {
	return c.reader.Read(bytes)
}
func (c *brotliDecompressor) Reset(rdr io.Reader) error {
	return c.reader.Reset(rdr)
}
func (c *brotliDecompressor) Close() error {
	// brotli's Reader does not expose a Close function
	return nil
}

// NewBrotliDecompressor returns a new Brotli Decompressor.
func NewBrotliDecompressor() connect.Decompressor {
	return &brotliDecompressor{
		reader: brotli.NewReader(nil),
	}
}

// NewBrotliCompressor returns a new Brotli Compressor.
func NewBrotliCompressor() connect.Compressor {
	return brotli.NewWriter(nil)
}
