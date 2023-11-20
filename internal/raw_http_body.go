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

package internal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"connectrpc.com/conformance/internal/compression"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/connect"
	"github.com/klauspost/compress/gzip"
)

// WriteRawMessageContents writes the given message contents to w.
func WriteRawMessageContents(contents *conformancev1.MessageContents, writer io.Writer) error {
	var msgBytes []byte
	switch data := contents.Data.(type) {
	case nil:
		// empty, so nothing to write
		return nil
	case *conformancev1.MessageContents_Binary:
		msgBytes = data.Binary
	case *conformancev1.MessageContents_BinaryMessage:
		msgBytes = data.BinaryMessage.Value
	case *conformancev1.MessageContents_Text:
		msgBytes = []byte(data.Text)
	default:
		return fmt.Errorf("invalid message contents data type: %T", data)
	}

	var compressor connect.Compressor
	switch contents.Compression {
	case conformancev1.Compression_COMPRESSION_IDENTITY, conformancev1.Compression_COMPRESSION_UNSPECIFIED:
		// no compression
		_, err := writer.Write(msgBytes)
		return err
	case conformancev1.Compression_COMPRESSION_GZIP:
		compressor = gzip.NewWriter(nil)
	case conformancev1.Compression_COMPRESSION_BR:
		compressor = compression.NewBrotliCompressor()
	case conformancev1.Compression_COMPRESSION_ZSTD:
		compressor = compression.NewZstdCompressor()
	case conformancev1.Compression_COMPRESSION_SNAPPY:
		compressor = compression.NewSnappyCompressor()
	case conformancev1.Compression_COMPRESSION_DEFLATE:
		compressor = compression.NewDeflateCompressor()
	default:
		return fmt.Errorf("unknown compression type: %v", contents.Compression)
	}
	compressor.Reset(writer)
	_, err := compressor.Write(msgBytes)
	if err == nil {
		err = compressor.Close()
	}
	return err
}

// WriteRawStreamContents writes the given stream contents to w.
func WriteRawStreamContents(contents *conformancev1.StreamContents, writer io.Writer) error {
	for i, item := range contents.Items {
		var prefix [5]byte
		if item.Flags > 255 {
			return fmt.Errorf("message #%d: flags is out of range: %d, should be [0,255]", i+1, item.Flags)
		}
		prefix[0] = byte(item.Flags)
		if item.Length != nil {
			binary.BigEndian.PutUint32(prefix[1:], item.GetLength())
			_, err := writer.Write(prefix[:])
			if err == nil {
				err = WriteRawMessageContents(item.Payload, writer)
			}
			if err != nil {
				return fmt.Errorf("message #%d: %w", i+1, err)
			}
			continue
		}

		var buf bytes.Buffer
		if err := WriteRawMessageContents(item.Payload, &buf); err != nil {
			return fmt.Errorf("message #%d: %w", i+1, err)
		}
		binary.BigEndian.PutUint32(prefix[1:], uint32(buf.Len()))
		_, err := writer.Write(prefix[:])
		if err == nil {
			_, err = buf.WriteTo(writer)
		}
		if err != nil {
			return fmt.Errorf("message #%d: %w", i+1, err)
		}
	}
	return nil
}
