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

// Package wsframe writes and observes RFC 6455 WebSocket frames, so that raw
// test cases can put exact bytes on a connection that connect-go otherwise
// manages. It knows nothing of Connect over WebSocket: the markers and
// payloads of the messages it writes come from the test case definition.
package wsframe

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"sync"

	"connectrpc.com/conformance/internal"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
)

// Frame opcodes.
const (
	OpText   = 0x1
	OpBinary = 0x2
	OpClose  = 0x8
)

// Close codes with a meaning to the conformance suite rather than a peer.
const (
	// CloseNoStatus is reported for a Close frame without a code.
	CloseNoStatus = 1005
	// CloseAbnormal is reported for a connection that ended without a Close
	// frame. A peer never sends it.
	CloseAbnormal = 1006
	// CloseClientError is the only error code a Connect-over-WebSocket client
	// may send (§13). The code says nothing of the cause, so it needs a reason.
	CloseClientError = 3102
)

// maxClosePayload is the most a Close frame carries: RFC 6455 limits control
// frames to 125 bytes, two of them the code.
const maxClosePayload = 125

// AppendFrame appends a single final frame to dst. A client masks its frames
// and a server does not.
func AppendFrame(dst []byte, opcode byte, payload []byte, masked bool) []byte {
	dst = append(dst, 0x80|opcode)
	var maskBit byte
	if masked {
		maskBit = 0x80
	}
	switch length := len(payload); {
	case length < 126:
		dst = append(dst, maskBit|byte(length))
	case length <= 0xFFFF:
		dst = append(dst, maskBit|126)
		dst = binary.BigEndian.AppendUint16(dst, uint16(length))
	default:
		dst = append(dst, maskBit|127)
		dst = binary.BigEndian.AppendUint64(dst, uint64(length))
	}
	if !masked {
		return append(dst, payload...)
	}
	var key [4]byte
	_, _ = rand.Read(key[:])
	dst = append(dst, key[:]...)
	for i, b := range payload {
		dst = append(dst, b^key[i%4])
	}
	return dst
}

// CloseFrame returns a Close frame that carries code and reason.
func CloseFrame(code int, reason string, masked bool) []byte {
	payload := binary.BigEndian.AppendUint16(nil, uint16(code))
	return AppendFrame(nil, OpClose, append(payload, reason...), masked)
}

// Messages returns the frames for the messages of a raw request or response
// definition, one frame per message.
func Messages(messages []*conformancev1.WebSocketMessage, masked bool) ([]byte, error) {
	var frames []byte
	for i, message := range messages {
		var opcode byte
		switch message.GetType() {
		case conformancev1.WebSocketMessageType_WEB_SOCKET_MESSAGE_TYPE_TEXT:
			opcode = OpText
		case conformancev1.WebSocketMessageType_WEB_SOCKET_MESSAGE_TYPE_BINARY:
			opcode = OpBinary
		default:
			return nil, fmt.Errorf("raw message #%d: no frame type specified", i+1)
		}
		var payload bytes.Buffer
		payload.WriteString(message.GetMarker())
		if message.GetContents() != nil {
			if err := internal.WriteRawMessageContents(message.GetContents(), &payload); err != nil {
				return nil, fmt.Errorf("raw message #%d: %w", i+1, err)
			}
		}
		frames = AppendFrame(frames, opcode, payload.Bytes(), masked)
	}
	return frames, nil
}

// CloseWatcher follows the frames of one direction of a connection, as they
// are read, and records the code and reason of the first Close frame among
// them.
type CloseWatcher struct {
	mu        sync.Mutex
	header    []byte
	inPayload bool
	opcode    byte
	masked    bool
	maskKey   [4]byte
	remaining uint64
	offset    uint64
	payload   []byte // the payload of a Close frame, up to maxClosePayload
	code      int
	reason    string
	closed    bool
	seen      chan struct{}
}

// NewCloseWatcher returns a CloseWatcher positioned at the start of a frame.
func NewCloseWatcher() *CloseWatcher {
	return &CloseWatcher{seen: make(chan struct{})}
}

// Observe advances the watcher past data, the next bytes of the connection.
func (w *CloseWatcher) Observe(data []byte) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for len(data) > 0 {
		if !w.inPayload {
			w.header = append(w.header, data[0])
			data = data[1:]
			if length := headerLength(w.header); length > 0 && len(w.header) == length {
				w.startPayload()
			}
			continue
		}
		chunk := data[:min(uint64(len(data)), w.remaining)]
		if w.opcode == OpClose {
			for _, b := range chunk {
				if w.offset < maxClosePayload {
					w.payload = append(w.payload, b^w.maskByte(w.offset))
				}
				w.offset++
			}
		}
		data = data[len(chunk):]
		w.remaining -= uint64(len(chunk))
		if w.remaining == 0 {
			w.endFrame()
		}
	}
}

// Seen returns a channel that is closed once a Close frame has been read.
func (w *CloseWatcher) Seen() <-chan struct{} {
	return w.seen
}

// Code returns the code of the first Close frame, and false if none has been
// read.
func (w *CloseWatcher) Code() (int, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.code, w.closed
}

// Reason returns the reason of the first Close frame, or "" if it has none or
// none has been read.
func (w *CloseWatcher) Reason() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.reason
}

func (w *CloseWatcher) maskByte(offset uint64) byte {
	if !w.masked {
		return 0
	}
	return w.maskKey[offset%4]
}

func (w *CloseWatcher) startPayload() {
	header := w.header
	w.opcode = header[0] & 0x0F
	w.masked = header[1]&0x80 != 0
	rest := header[2:]
	switch length := header[1] & 0x7F; length {
	case 126:
		w.remaining = uint64(binary.BigEndian.Uint16(rest))
		rest = rest[2:]
	case 127:
		w.remaining = binary.BigEndian.Uint64(rest)
		rest = rest[8:]
	default:
		w.remaining = uint64(length)
	}
	if w.masked {
		copy(w.maskKey[:], rest)
	}
	w.header, w.offset, w.payload = w.header[:0], 0, nil
	w.inPayload = true
	if w.remaining == 0 {
		w.endFrame()
	}
}

func (w *CloseWatcher) endFrame() {
	w.inPayload = false
	if w.opcode != OpClose || w.closed {
		return
	}
	w.code = CloseNoStatus
	if len(w.payload) >= 2 {
		w.code = int(binary.BigEndian.Uint16(w.payload))
		w.reason = string(w.payload[2:])
	}
	w.closed = true
	close(w.seen)
}

// headerLength returns the length of a frame header that begins with header,
// or zero if header is too short to tell.
func headerLength(header []byte) int {
	if len(header) < 2 {
		return 0
	}
	length := 2
	switch header[1] & 0x7F {
	case 126:
		length += 2
	case 127:
		length += 8
	}
	if header[1]&0x80 != 0 {
		length += 4
	}
	return length
}
