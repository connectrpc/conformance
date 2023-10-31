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

package client

import (
	"io"

	"connectrpc.com/connect"
	"github.com/klauspost/compress/zstd"
)

type ZstdDecompressor struct {
	decoder *zstd.Decoder
}

func NewZstdDecompressor() connect.Decompressor {
	d, _ := zstd.NewReader(nil)
	return &ZstdDecompressor{
		decoder: d,
	}
}
func (c *ZstdDecompressor) Read(bytes []byte) (int, error) {
	return c.decoder.Read(bytes)
}
func (c *ZstdDecompressor) Reset(rdr io.Reader) error {
	return c.decoder.Reset(rdr)
}
func (c *ZstdDecompressor) Close() error {
	c.decoder.Close()
	return nil
}

type ZstdCompressor struct {
	*zstd.Encoder
}

func NewZstdCompressor() connect.Compressor {
	w, _ := zstd.NewWriter(nil)
	return w
}
