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

package referenceserver

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"sync"
	"time"

	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/wsframe"
)

const (
	// wsClientCloseTimeout is how long the reference server waits for a
	// client to close the connection after a raw response.
	wsClientCloseTimeout = 10 * time.Second
	wsCloseNormal        = 1000
)

var errNoWebSocketRawResponseHolder = errors.New(
	"reference server needs to use a raw WebSocket response but the RPC is not on a WebSocket connection")

// wsRawResponseKey stores the *wsRawWriter of a WebSocket handshake in the
// context, so that rawResponseRecorder can direct it to send a raw response.
type wsRawResponseKey struct{}

// wsRawWriter is the response writer of a WebSocket handshake. It hands
// connect-go a connection whose output a raw response can replace.
type wsRawWriter struct {
	http.ResponseWriter

	mu       sync.Mutex
	conn     *wsServerConn
	raw      *conformancev1.RawWebSocketResponse
	feedback *feedbackPrinter
}

func (w *wsRawWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("response writer does not support hijacking")
	}
	netConn, readWriter, err := hijacker.Hijack()
	if err != nil {
		return nil, nil, err
	}
	// The server may have read the client's first frames along with the
	// handshake; they are the first bytes of the connection.
	buffered, _ := readWriter.Peek(readWriter.Reader.Buffered())
	conn := &wsServerConn{
		Conn:    netConn,
		reader:  io.MultiReader(bytes.NewReader(slices.Clone(buffered)), netConn),
		watcher: wsframe.NewCloseWatcher(),
	}
	w.mu.Lock()
	w.conn = conn
	w.mu.Unlock()
	return conn, bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil
}

func (w *wsRawWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *wsRawWriter) setFeedback(feedback *feedbackPrinter) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.feedback = feedback
}

// setRawWebSocketResponse directs the connection of the RPC in ctx to send
// resp, and discards anything connect-go writes to it from now on.
func setRawWebSocketResponse(ctx context.Context, resp *conformancev1.RawWebSocketResponse) error {
	writer, ok := ctx.Value(wsRawResponseKey{}).(*wsRawWriter)
	if !ok {
		return errNoWebSocketRawResponseHolder
	}
	writer.mu.Lock()
	defer writer.mu.Unlock()
	if writer.conn == nil {
		return errNoWebSocketRawResponseHolder
	}
	writer.raw = resp
	writer.conn.discardWrites()
	return nil
}

// writeRawWebSocketResponse sends the raw response that
// setRawWebSocketResponse stored, then ends the connection as it directs.
func writeRawWebSocketResponse(ctx context.Context) {
	writer, ok := ctx.Value(wsRawResponseKey{}).(*wsRawWriter)
	if !ok {
		return
	}
	writer.mu.Lock()
	conn, raw, feedback := writer.conn, writer.raw, writer.feedback
	writer.mu.Unlock()
	if conn == nil || raw == nil {
		return
	}
	frames, err := wsframe.Messages(raw.GetMessages(), false)
	if err != nil && feedback != nil {
		feedback.Printf("invalid raw WebSocket response: %v", err)
	}
	// A client that detects a protocol error may close the connection before
	// every message is written, so a failed write is not a fault.
	_ = conn.writeRaw(frames)
	expected := raw.GetExpectedClientCloseCodes()
	switch {
	case len(expected) > 0:
	case raw.GetOmitClose():
		_ = conn.Conn.Close()
		return
	default:
		code := int(raw.GetCloseCode())
		if code == 0 {
			code = wsCloseNormal
		}
		_ = conn.writeRaw(wsframe.CloseFrame(code, "", false))
	}
	go func() {
		code, reason, ok := conn.awaitClientClose()
		_ = conn.Conn.Close()
		if len(expected) == 0 || feedback == nil {
			return
		}
		if !ok {
			feedback.Printf("client did not close the connection within %v", wsClientCloseTimeout)
			return
		}
		if problem := checkClientClose(expected, code, reason); problem != "" {
			feedback.Printf("%s", problem)
		}
	}()
}

// checkClientClose describes how the client's Close frame falls short of
// expected, or returns "" if it does not.
func checkClientClose(expected []uint32, code int, reason string) string {
	if !slices.Contains(expected, uint32(code)) {
		return fmt.Sprintf("client closed the connection with code %d; expected one of %v", code, expected)
	}
	// §13: the code alone cannot tell a server what went wrong.
	if code == wsframe.CloseClientError && reason == "" {
		return fmt.Sprintf("client closed the connection with code %d but no reason", code)
	}
	return ""
}

// wsServerConn is the connection under connect-go's WebSocket handler. It
// watches what the client sends for a Close frame. Once a raw response has
// taken over, it drops what connect-go writes and leaves closing to the raw
// response.
type wsServerConn struct {
	net.Conn

	reader  io.Reader
	watcher *wsframe.CloseWatcher
	// readMu keeps reads, and so the bytes the watcher observes, in order.
	readMu  sync.Mutex
	writeMu sync.Mutex
	raw     bool
}

func (c *wsServerConn) Read(data []byte) (int, error) {
	c.readMu.Lock()
	defer c.readMu.Unlock()
	n, err := c.reader.Read(data)
	c.watcher.Observe(data[:n])
	return n, err
}

func (c *wsServerConn) Write(data []byte) (int, error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if c.raw {
		return len(data), nil
	}
	return c.Conn.Write(data)
}

func (c *wsServerConn) Close() error {
	c.writeMu.Lock()
	raw := c.raw
	c.writeMu.Unlock()
	if raw {
		return nil // writeRawWebSocketResponse closes the connection
	}
	return c.Conn.Close()
}

func (c *wsServerConn) discardWrites() {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	c.raw = true
}

func (c *wsServerConn) writeRaw(data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_, err := c.Conn.Write(data)
	return err
}

// awaitClientClose reads until the client's Close frame arrives, and returns
// its code and reason. A connection that ends without one reports 1006. It
// reports false if the client neither closed nor ended the connection in time.
func (c *wsServerConn) awaitClientClose() (int, string, bool) {
	_ = c.SetReadDeadline(time.Now().Add(wsClientCloseTimeout))
	buf := make([]byte, 512)
	for {
		select {
		case <-c.watcher.Seen():
			code, _ := c.watcher.Code()
			return code, c.watcher.Reason(), true
		default:
		}
		if _, err := c.Read(buf); err != nil {
			if code, ok := c.watcher.Code(); ok {
				return code, c.watcher.Reason(), true
			}
			if netErr, ok := errors.AsType[net.Error](err); ok && netErr.Timeout() {
				return 0, "", false
			}
			return wsframe.CloseAbnormal, "", true
		}
	}
}
