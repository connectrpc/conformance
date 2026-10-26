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
	"github.com/golang/snappy"
)

type snappyCompressor struct{}

// NewSnappyCompressor returns a snappy [connect.Compressor].
func NewSnappyCompressor() connect.Compressor {
	return snappyCompressor{}
}

func (snappyCompressor) Name() string { return Snappy }

func (snappyCompressor) Compress(dst io.Writer) (io.WriteCloser, error) {
	return snappy.NewBufferedWriter(dst), nil
}

func (snappyCompressor) Decompress(src io.Reader) (io.ReadCloser, error) {
	return io.NopCloser(snappy.NewReader(src)), nil
}
