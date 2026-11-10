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

package referenceclient

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"connectrpc.com/conformance/internal"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/wsframe"
	"connectrpc.com/connect/v2/connecthttp"
)

const (
	// wsCloseWait bounds how long the client waits for the server's Close
	// frame once the RPC is over, so that its code can be reported.
	wsCloseWait = 2 * time.Second
	// wsCloseCompression is the client's close code for a server that
	// negotiated compression with a shared context.
	wsCloseCompression = 3110
	wsTimeoutParam     = "connect-timeout-ms"
)

// webSocketWire is the round-tripper for one RPC that connect-go carries over
// WebSocket. It records the handshake's status code and the code of the
// server's Close frame. For a raw request, it also replaces the handshake
// details and the messages that connect-go sends with those the request
// defines; connect-go still reads and decodes the server's response.
type webSocketWire struct {
	transport http.RoundTripper
	raw       *conformancev1.RawWebSocketRequest

	mu         sync.Mutex
	statusCode *int32
	conn       *wsClientConn
}

func (w *webSocketWire) RoundTrip(req *http.Request) (*http.Response, error) {
	if !connecthttp.IsWebSocketUpgrade(req) {
		return w.transport.RoundTrip(req)
	}
	if w.raw != nil {
		req = w.rewrite(req)
	}
	resp, err := w.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	w.mu.Lock()
	w.statusCode = new(int32(resp.StatusCode))
	w.mu.Unlock()
	rwc, ok := resp.Body.(io.ReadWriteCloser)
	if resp.StatusCode != http.StatusSwitchingProtocols || !ok {
		return resp, nil
	}
	conn := &wsClientConn{
		rwc:           rwc,
		watcher:       wsframe.NewCloseWatcher(),
		discardWrites: w.raw != nil,
		ended:         make(chan struct{}),
	}
	w.mu.Lock()
	w.conn = conn
	w.mu.Unlock()
	resp.Body = conn
	if w.raw == nil {
		return resp, nil
	}
	if extension, ok := sharedCompressionContext(resp.Header); ok {
		_, _ = rwc.Write(wsframe.CloseFrame(wsCloseCompression, true))
		_ = rwc.Close()
		return nil, fmt.Errorf("server negotiated %q without no_context_takeover in both directions", extension)
	}
	frames, err := wsframe.Messages(w.raw.GetMessages(), true)
	if err != nil {
		_ = rwc.Close()
		return nil, err
	}
	// Asynchronous, because a server may end the stream before it has read
	// every message, for example when it rejects the deadline.
	go func() { _, _ = rwc.Write(frames) }()
	return resp, nil
}

// rewrite applies the handshake details of the raw request to req.
func (w *webSocketWire) rewrite(req *http.Request) *http.Request {
	req = req.Clone(req.Context())
	if path := w.raw.GetPath(); path != "" {
		req.URL.Path, req.URL.RawPath = path, ""
	}
	// Only the raw request decides whether there is a deadline.
	query := req.URL.Query()
	query.Del(wsTimeoutParam)
	for _, param := range w.raw.GetQueryParams() {
		for _, value := range param.Value {
			query.Add(param.Name, value)
		}
	}
	req.URL.RawQuery = query.Encode()
	internal.AddHeaders(w.raw.GetHeaders(), req.Header)
	if subprotocols := w.raw.GetSubprotocols(); len(subprotocols) > 0 {
		req.Header.Set("Sec-WebSocket-Protocol", strings.Join(subprotocols, ", "))
	}
	// The scripted messages are written uncompressed, which a server must
	// accept whether or not it negotiated compression.
	req.Header.Del("Sec-WebSocket-Extensions")
	if w.raw.GetOfferContextTakeover() {
		req.Header.Set("Sec-WebSocket-Extensions", "permessage-deflate")
	}
	return req
}

func (w *webSocketWire) status() *int32 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.statusCode
}

// closeCode returns the code of the server's Close frame, 1006 if the
// connection ended without one, or nil if there was no connection or it has
// not ended.
func (w *webSocketWire) closeCode() *int32 {
	w.mu.Lock()
	conn := w.conn
	w.mu.Unlock()
	if conn == nil {
		return nil
	}
	select {
	case <-conn.watcher.Seen():
	case <-conn.ended:
	case <-time.After(wsCloseWait):
		return nil
	}
	code, ok := conn.watcher.Code()
	if !ok {
		code = wsframe.CloseAbnormal
	}
	return new(int32(code))
}

// sharedCompressionContext reports a permessage-deflate extension in the
// handshake response that does not reset its context for every message.
func sharedCompressionContext(header http.Header) (string, bool) {
	for _, value := range header.Values("Sec-WebSocket-Extensions") {
		for extension := range strings.SplitSeq(value, ",") {
			params := strings.Split(extension, ";")
			if strings.TrimSpace(params[0]) != "permessage-deflate" {
				continue
			}
			var clientReset, serverReset bool
			for _, param := range params[1:] {
				switch strings.TrimSpace(param) {
				case "client_no_context_takeover":
					clientReset = true
				case "server_no_context_takeover":
					serverReset = true
				}
			}
			if !clientReset || !serverReset {
				return strings.TrimSpace(extension), true
			}
		}
	}
	return "", false
}

// wsClientConn is the connection under connect-go's WebSocket client. It
// watches what the server sends for a Close frame, and drops what connect-go
// writes when a raw request has replaced it.
type wsClientConn struct {
	rwc           io.ReadWriteCloser
	watcher       *wsframe.CloseWatcher
	discardWrites bool

	// readMu keeps reads, and so the bytes the watcher observes, in order.
	readMu    sync.Mutex
	closeOnce sync.Once
	endOnce   sync.Once
	ended     chan struct{}
}

func (c *wsClientConn) Read(data []byte) (int, error) {
	c.readMu.Lock()
	defer c.readMu.Unlock()
	n, err := c.rwc.Read(data)
	c.watcher.Observe(data[:n])
	if err != nil {
		c.end()
	}
	return n, err
}

func (c *wsClientConn) Write(data []byte) (int, error) {
	if c.discardWrites {
		return len(data), nil
	}
	return c.rwc.Write(data)
}

// Close keeps reading until the server's Close frame arrives, so that its
// code can be reported even when connect-go hangs up without waiting for it.
func (c *wsClientConn) Close() error {
	c.closeOnce.Do(func() {
		go func() {
			timer := time.AfterFunc(wsCloseWait, func() { _ = c.rwc.Close() })
			defer timer.Stop()
			buf := make([]byte, 512)
			for {
				select {
				case <-c.watcher.Seen():
					_ = c.rwc.Close()
					c.end()
					return
				default:
				}
				if _, err := c.Read(buf); err != nil {
					_ = c.rwc.Close()
					return
				}
			}
		}()
	})
	return nil
}

func (c *wsClientConn) end() {
	c.endOnce.Do(func() { close(c.ended) })
}
