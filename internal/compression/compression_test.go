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
	"bytes"
	"io"
	"strings"
	"testing"

	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressorsRoundTrip(t *testing.T) {
	t.Parallel()
	payload := []byte(strings.Repeat("connect conformance ", 100))
	for value, name := range conformancev1.Compression_name {
		if value == int32(conformancev1.Compression_COMPRESSION_UNSPECIFIED) {
			continue
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			compressor, err := GetCompressor(conformancev1.Compression(value))
			require.NoError(t, err)
			var compressed bytes.Buffer
			writer, err := compressor.Compress(&compressed)
			require.NoError(t, err)
			_, err = writer.Write(payload)
			require.NoError(t, err)
			require.NoError(t, writer.Close())

			reader, err := compressor.Decompress(&compressed)
			require.NoError(t, err)
			decompressed, err := io.ReadAll(reader)
			require.NoError(t, err)
			require.NoError(t, reader.Close())
			assert.Equal(t, payload, decompressed)
		})
	}
}

func TestAllNames(t *testing.T) {
	t.Parallel()
	names := make([]string, 0, len(All()))
	for _, compressor := range All() {
		names = append(names, compressor.Name())
	}
	assert.Equal(t, []string{Zstd, Snappy, Deflate, Brotli, Gzip}, names)
}
