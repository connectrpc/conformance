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

	"connectrpc.com/connect/v2"
	"github.com/klauspost/compress/zstd"
)

type zstdCompressor struct{}

// NewZstdCompressor returns a Zstandard [connect.Compressor].
func NewZstdCompressor() connect.Compressor {
	return zstdCompressor{}
}

func (zstdCompressor) Name() string { return Zstd }

func (zstdCompressor) Compress(dst io.Writer) (io.WriteCloser, error) {
	return zstd.NewWriter(dst)
}

func (zstdCompressor) Decompress(src io.Reader) (io.ReadCloser, error) {
	decoder, err := zstd.NewReader(src)
	if err != nil {
		return nil, err
	}
	return decoder.IOReadCloser(), nil
}
