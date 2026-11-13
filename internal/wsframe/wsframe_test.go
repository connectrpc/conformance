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

package wsframe

import (
	"bytes"
	"testing"

	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"github.com/stretchr/testify/require"
)

func TestAppendFrame(t *testing.T) {
	t.Parallel()
	require.Equal(t, []byte{0x81, 0x03, 'B', '{', '}'}, AppendFrame(nil, OpText, []byte("B{}"), false))

	masked := AppendFrame(nil, OpBinary, []byte("B"), true)
	require.Len(t, masked, 7)
	require.Equal(t, []byte{0x82, 0x81}, masked[:2])
	require.Equal(t, byte('B'), masked[6]^masked[2])

	long := AppendFrame(nil, OpBinary, make([]byte, 300), false)
	require.Equal(t, []byte{0x82, 126, 0x01, 0x2C}, long[:4])
	require.Len(t, long, 304)
}

func TestMessages(t *testing.T) {
	t.Parallel()
	frames, err := Messages([]*conformancev1.WebSocketMessage{
		{
			Type:     conformancev1.WebSocketMessageType_WEB_SOCKET_MESSAGE_TYPE_TEXT,
			Marker:   "M",
			Contents: &conformancev1.MessageContents{Data: &conformancev1.MessageContents_Text{Text: "{}"}},
		},
		{Type: conformancev1.WebSocketMessageType_WEB_SOCKET_MESSAGE_TYPE_BINARY, Marker: "C"},
	}, false)
	require.NoError(t, err)
	require.Equal(t, []byte{0x81, 0x03, 'M', '{', '}', 0x82, 0x01, 'C'}, frames)

	_, err = Messages([]*conformancev1.WebSocketMessage{{Marker: "B"}}, false)
	require.Error(t, err)
}

func TestCloseWatcher(t *testing.T) {
	t.Parallel()
	for _, masked := range []bool{false, true} {
		var stream bytes.Buffer
		stream.Write(AppendFrame(nil, OpText, []byte("B{}"), masked))
		stream.Write(AppendFrame(nil, OpBinary, make([]byte, 70000), masked))
		stream.Write(CloseFrame(3102, "bad marker", masked))
		stream.Write(CloseFrame(1000, "", masked))

		watcher := NewCloseWatcher()
		// One byte at a time, as the worst case of how reads may split frames.
		for _, b := range stream.Bytes() {
			_, ok := watcher.Code()
			require.False(t, ok)
			watcher.Observe([]byte{b})
			if _, ok := watcher.Code(); ok {
				break
			}
		}
		code, ok := watcher.Code()
		require.True(t, ok)
		require.Equal(t, 3102, code)
		require.Equal(t, "bad marker", watcher.Reason())
		select {
		case <-watcher.Seen():
		default:
			t.Fatal("Seen not closed")
		}
		watcher.Observe(stream.Bytes())
		code, _ = watcher.Code()
		require.Equal(t, 3102, code, "first Close frame wins")
		require.Equal(t, "bad marker", watcher.Reason(), "first Close frame wins")
	}

	watcher := NewCloseWatcher()
	watcher.Observe(CloseFrame(1000, "", true))
	code, ok := watcher.Code()
	require.True(t, ok)
	require.Equal(t, 1000, code)
	require.Empty(t, watcher.Reason())

	watcher = NewCloseWatcher()
	watcher.Observe(AppendFrame(nil, OpClose, nil, false))
	code, ok = watcher.Code()
	require.True(t, ok)
	require.Equal(t, CloseNoStatus, code)
}
