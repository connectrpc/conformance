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
	"io"
	"strings"
	"testing"
	"time"

	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestReadDelimitedMessage(t *testing.T) {
	t.Parallel()
	t.Run("multiple-reads", func(t *testing.T) {
		t.Parallel()
		msg := conformancev1.ClientCompatResponse{
			TestName: "abc/def/xyz",
			Result: &conformancev1.ClientCompatResponse_Response{
				Response: &conformancev1.ClientResponseResult{
					Payloads: []*conformancev1.ConformancePayload{
						{
							Data: make([]byte, 1000),
						},
						{
							Data: make([]byte, 2345),
						},
					},
				},
			},
		}
		data, err := proto.Marshal(&msg)
		require.NoError(t, err)
		var size [4]byte
		binary.BigEndian.PutUint32(size[:], uint32(len(data)))
		require.NoError(t, err)
		in := io.MultiReader(
			bytes.NewReader(size[:]),
			// When trying to read the full message, this will cause first read to
			// return 100 bytes, then 200, 300, 400, etc. So this makes sure we
			// are accumulating the data slice correctly.
			bytes.NewReader(data[:100]),
			bytes.NewReader(data[100:300]),
			bytes.NewReader(data[300:600]),
			bytes.NewReader(data[600:1000]),
			bytes.NewReader(data[1000:]),
		)
		var result conformancev1.ClientCompatResponse
		err = ReadDelimitedMessage(in, &result, "client", time.Second, 16*1024*1024)
		require.NoError(t, err)
		require.Empty(t, cmp.Diff(&msg, &result, protocmp.Transform()))
		// make sure no data left in reader
		n, err := in.Read(make([]byte, 1))
		require.Zero(t, n)
		require.ErrorIs(t, err, io.EOF)
	})
	t.Run("eof", func(t *testing.T) {
		t.Parallel()
		var msg conformancev1.ClientCompatResponse
		err := ReadDelimitedMessage(bytes.NewReader(nil), &msg, "client", time.Second, 16*1024*1024)
		require.ErrorIs(t, err, io.EOF)
	})
	t.Run("unexpected-eof", func(t *testing.T) {
		t.Parallel()
		var data [50]byte
		binary.BigEndian.PutUint32(data[:], 123)
		var msg conformancev1.ClientCompatResponse
		err := ReadDelimitedMessage(bytes.NewReader(data[:]), &msg, "client", time.Second, 16*1024*1024)
		require.ErrorIs(t, err, io.ErrUnexpectedEOF)
	})
	t.Run("read-prefix-but-no-message", func(t *testing.T) {
		t.Parallel()
		var msg conformancev1.ClientCompatResponse
		var size [4]byte
		binary.BigEndian.PutUint32(size[:], 1234)
		in := bytes.NewReader(size[:])
		err := ReadDelimitedMessage(in, &msg, "client", time.Second, 16*1024*1024)
		require.ErrorIs(t, err, io.ErrUnexpectedEOF)
	})
	t.Run("timeout-read-nothing", func(t *testing.T) {
		t.Parallel()
		var msg conformancev1.ClientCompatResponse
		err := ReadDelimitedMessage(stuckReader{}, &msg, "client", time.Second, 16*1024*1024)
		require.ErrorContains(t, err, "timed out waiting for result from client")
	})
	t.Run("timeout-read-partial-prefix", func(t *testing.T) {
		t.Parallel()
		var msg conformancev1.ClientCompatResponse
		in := io.MultiReader(bytes.NewReader([]byte{1, 2, 3}), stuckReader{})
		err := ReadDelimitedMessage(in, &msg, "client", time.Second, 16*1024*1024)
		require.ErrorContains(t, err, "timed out waiting for result from client: read 3/4 bytes of length prefix")
	})
	t.Run("timeout-read-prefix-but-no-message", func(t *testing.T) {
		t.Parallel()
		var msg conformancev1.ClientCompatResponse
		var size [4]byte
		binary.BigEndian.PutUint32(size[:], 1234)
		in := io.MultiReader(bytes.NewReader(size[:]), stuckReader{})
		err := ReadDelimitedMessage(in, &msg, "client", time.Second, 16*1024*1024)
		require.ErrorContains(t, err, "timed out waiting for result from client: read 0/1234 bytes of message")
	})
	t.Run("timeout-read-partial-message", func(t *testing.T) {
		t.Parallel()
		var msg conformancev1.ClientCompatResponse
		var size [4]byte
		binary.BigEndian.PutUint32(size[:], 12345)
		in := io.MultiReader(
			bytes.NewReader(size[:]),
			// When trying to read 12,345 bytes, this will cause first read to
			// return 100 bytes, then 200, 300, and finally 399. So we make sure
			// we are accumulating total across multiple reads correctly.
			bytes.NewReader(make([]byte, 100)),
			bytes.NewReader(make([]byte, 200)),
			bytes.NewReader(make([]byte, 300)),
			bytes.NewReader(make([]byte, 399)),
			stuckReader{},
		)
		err := ReadDelimitedMessage(in, &msg, "client", time.Second, 16*1024*1024)
		require.ErrorContains(t, err, "timed out waiting for result from client: read 999/12345 bytes of message")
	})
	t.Run("max-size", func(t *testing.T) {
		t.Parallel()
		// Sizes are limited so that most-significant byte will be zero, making
		// it likely to distinguish a valid size from process inadvertently
		// writing other stuff to its stdout.
		var msg conformancev1.ClientCompatResponse
		err := ReadDelimitedMessage(strings.NewReader("Poop!"), &msg, "client", time.Second, 16*1024*1024)
		require.ErrorContains(t, err, "but should not exceed")
	})
}

type stuckReader struct{}

func (s stuckReader) Read(_ []byte) (int, error) {
	time.Sleep(5 * time.Second)
	return 0, io.EOF
}
