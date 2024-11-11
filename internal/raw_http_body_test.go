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

package internal

import (
	"bytes"
	"encoding/binary"
	"testing"

	"connectrpc.com/conformance/internal/compression"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestWriteRawMessageContents(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		data *conformancev1.MessageContents
	}{
		{
			name: "binary",
			data: &conformancev1.MessageContents{
				Data: &conformancev1.MessageContents_Binary{
					Binary: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
				},
			},
		},
		{
			name: "text",
			data: &conformancev1.MessageContents{
				Data: &conformancev1.MessageContents_Text{
					Text: `Law Blog of Bob Loblaw, Attorney at Law`,
				},
			},
		},
		{
			name: "message",
			data: &conformancev1.MessageContents{
				Data: &conformancev1.MessageContents_BinaryMessage{
					BinaryMessage: messageValue(t),
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// try all flavors of compression
			for compressInt, compressName := range conformancev1.Compression_name {
				compressionType := conformancev1.Compression(compressInt)
				t.Run(compressName, func(t *testing.T) {
					t.Parallel()

					data := proto.Clone(testCase.data).(*conformancev1.MessageContents) //nolint:errcheck,forcetypeassert
					data.Compression = compressionType

					var buf bytes.Buffer
					err := WriteRawMessageContents(data, &buf)
					require.NoError(t, err)

					checkMessageContents(t, data, buf.Bytes())
				})
			}
		})
	}
	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		var buf bytes.Buffer
		err := WriteRawMessageContents(&conformancev1.MessageContents{}, &buf)
		require.NoError(t, err)
		require.Zero(t, buf.Len())
	})
}

func TestWriteRawStreamContents(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		data []*conformancev1.StreamContents_StreamItem
	}{
		{
			name: "empty",
			data: nil,
		},
		{
			name: "simple",
			data: []*conformancev1.StreamContents_StreamItem{
				{
					Flags:  0,
					Length: proto.Uint32(10),
					Payload: &conformancev1.MessageContents{
						Data: &conformancev1.MessageContents_Binary{
							Binary: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
						},
					},
				},
			},
		},
		{
			name: "longer",
			data: []*conformancev1.StreamContents_StreamItem{
				{
					Flags: 123,
					Payload: &conformancev1.MessageContents{
						Data: &conformancev1.MessageContents_Binary{
							Binary: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
						},
						Compression: conformancev1.Compression_COMPRESSION_BR,
					},
				},
				{
					Flags: 22,
					Payload: &conformancev1.MessageContents{
						Data: &conformancev1.MessageContents_Text{
							Text: `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.`,
						},
					},
				},
				{
					Flags: 1,
					Payload: &conformancev1.MessageContents{
						Data: &conformancev1.MessageContents_BinaryMessage{
							BinaryMessage: messageValue(t),
						},
						Compression: conformancev1.Compression_COMPRESSION_ZSTD,
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := WriteRawStreamContents(&conformancev1.StreamContents{Items: testCase.data}, &buf)
			require.NoError(t, err)

			checkStreamContents(t, testCase.data, buf.Bytes())
		})
	}
}

func messageValue(t *testing.T) *anypb.Any {
	t.Helper()
	val, err := structpb.NewValue(map[string]any{
		"abc": "xyz",
		"def": []any{
			1.0, 123, "foo", false,
		},
		"ghi": map[string]any{
			"foo": "bar",
			"baz": -99,
		},
	})
	require.NoError(t, err)
	msgPayload := &anypb.Any{}
	err = anypb.MarshalFrom(msgPayload, val, proto.MarshalOptions{})
	require.NoError(t, err)
	return msgPayload
}

func checkMessageContents(t *testing.T, msg *conformancev1.MessageContents, data []byte) {
	t.Helper()
	decompressor, err := compression.GetDecompressor(msg.Compression)
	require.NoError(t, err)
	err = decompressor.Reset(bytes.NewReader(data))
	require.NoError(t, err)
	var decompressed bytes.Buffer
	_, err = decompressed.ReadFrom(decompressor)
	require.NoError(t, err)
	switch msgData := msg.Data.(type) {
	case *conformancev1.MessageContents_Binary:
		assert.Equal(t, msgData.Binary, decompressed.Bytes())
	case *conformancev1.MessageContents_Text:
		assert.Equal(t, msgData.Text, decompressed.String())
	case *conformancev1.MessageContents_BinaryMessage:
		assert.Equal(t, msgData.BinaryMessage.Value, decompressed.Bytes())
	case nil:
		assert.Zero(t, decompressed.Len())
	default:
		t.Fatalf("unknown type of contents: %T", msg.Data)
	}
}

func checkStreamContents(t *testing.T, items []*conformancev1.StreamContents_StreamItem, data []byte) {
	t.Helper()
	for _, item := range items {
		require.GreaterOrEqual(t, len(data), 5)
		assert.Equal(t, byte(item.Flags), data[0])
		length := binary.BigEndian.Uint32(data[1:5])
		if item.Length != nil {
			assert.Equal(t, item.GetLength(), length)
		}
		data = data[5:]
		require.GreaterOrEqual(t, len(data), int(length))
		msg := data[:length]
		checkMessageContents(t, item.Payload, msg)
		data = data[length:]
	}
	assert.Empty(t, data)
}
