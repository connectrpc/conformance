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
	"io"

	"connectrpc.com/connect/v2"
)

// deflateCompressor uses zlib (RFC 1950) rather than raw deflate (RFC 1951),
// because HTTP and gRPC both use "deflate" to name the zlib format.
type deflateCompressor struct{}

// NewDeflateCompressor returns a deflate [connect.Compressor].
func NewDeflateCompressor() connect.Compressor {
	return deflateCompressor{}
}

func (deflateCompressor) Name() string { return Deflate }

func (deflateCompressor) Compress(dst io.Writer) (io.WriteCloser, error) {
	return zlib.NewWriter(dst), nil
}

func (deflateCompressor) Decompress(src io.Reader) (io.ReadCloser, error) {
	return zlib.NewReader(src)
}
