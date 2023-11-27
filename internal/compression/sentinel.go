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
	"io"
)

// errorDecompressor is a sentinel type for a connect.Decompressor
// which will return an error upon first use.
type errorDecompressor struct {
	err error
}

func (c *errorDecompressor) Read(_ []byte) (int, error) {
	return 0, c.err
}
func (c *errorDecompressor) Reset(_ io.Reader) error {
	return c.err
}
func (c *errorDecompressor) Close() error {
	return c.err
}

// errorCompressor is a sentinel type for a connect.Compressor
// which will return an error upon first use.
type errorCompressor struct {
	err error
}

func (c *errorCompressor) Write(_ []byte) (int, error) {
	return 0, c.err
}

func (c *errorCompressor) Reset(_ io.Writer) {}

func (c *errorCompressor) Close() error {
	return c.err
}
