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
	"github.com/golang/snappy"
)

// snappyDecompressor is a thin wrapper around a snappy Reader.
type snappyDecompressor struct {
	reader *snappy.Reader
}

func (c *snappyDecompressor) Read(bytes []byte) (int, error) {
	return c.reader.Read(bytes)
}
func (c *snappyDecompressor) Reset(rdr io.Reader) error {
	c.reader.Reset(rdr)
	// snappy's Reader does not return an error on Reset
	return nil
}
func (c *snappyDecompressor) Close() error {
	// snappy's Reader does not expose a Close function
	return nil
}

// NewSnappyDecompressor returns a new snappy Decompressor.
func NewSnappyDecompressor() connect.Decompressor {
	return &snappyDecompressor{
		reader: snappy.NewReader(nil),
	}
}

// NewSnappyCompressor returns a new snappy Compressor.
func NewSnappyCompressor() connect.Compressor {
	return snappy.NewBufferedWriter(nil)
}
