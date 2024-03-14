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

package tracer

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"connectrpc.com/conformance/internal/compression"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func TestTracer(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name        string
		setupClient func(Collector) http.RoundTripper
		setupServer func(Collector, net.Listener, http.HandlerFunc) (net.Listener, *http.Server)
	}{
		{
			name: "http1-middleware",
			setupClient: func(collector Collector) http.RoundTripper {
				return TracingRoundTripper(http.DefaultTransport, collector)
			},
			setupServer: func(collector Collector, listener net.Listener, handler http.HandlerFunc) (net.Listener, *http.Server) {
				return listener, &http.Server{
					Handler:           TracingHandler(handler, collector),
					ReadHeaderTimeout: 5 * time.Second,
				}
			},
		},
		{
			name: "http2-middleware",
			// Uses H2C to force HTTP/2 w/out dealing with TLS.
			setupClient: func(collector Collector) http.RoundTripper {
				h2cTransport := &http2.Transport{
					AllowHTTP: true,
					DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
						return (&net.Dialer{}).DialContext(ctx, network, addr)
					},
				}
				return TracingRoundTripper(h2cTransport, collector)
			},
			setupServer: func(collector Collector, listener net.Listener, handler http.HandlerFunc) (net.Listener, *http.Server) {
				return listener, &http.Server{
					Handler:           h2c.NewHandler(TracingHandler(handler, collector), &http2.Server{}),
					ReadHeaderTimeout: 5 * time.Second,
				}
			},
		},
		{
			name: "http2-conn",
			// Wraps net.Conn instances instead of using net/http middleware.
			// Like above, uses H2C to force HTTP/2 w/out dealing with TLS.
			setupClient: func(collector Collector) http.RoundTripper {
				return &http2.Transport{
					AllowHTTP: true,
					DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
						conn, err := (&net.Dialer{}).DialContext(ctx, network, addr)
						if err != nil {
							return nil, err
						}
						return TracingHTTP2Conn(conn, false, collector), nil
					},
				}
			},
			setupServer: func(collector Collector, listener net.Listener, handler http.HandlerFunc) (net.Listener, *http.Server) {
				return TracingHTTP2Listener(listener, collector), &http.Server{
					Handler:           h2c.NewHandler(handler, &http2.Server{}),
					ReadHeaderTimeout: 5 * time.Second,
				}
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var clientTracer, serverTracer Tracer
			client := testCase.setupClient(&clientTracer)
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)

			responseData := []byte(`{"abc": "def","foo": "bar"}`)
			var responseLenPrefix [4]byte
			binary.BigEndian.PutUint32(responseLenPrefix[:], uint32(len(responseData)))

			comp := compression.NewSnappyCompressor()
			var compressedResponseBuffer bytes.Buffer
			comp.Reset(&compressedResponseBuffer)
			_, _ = comp.Write(responseData)
			_ = comp.Close()
			compressedResponseData := compressedResponseBuffer.Bytes()
			var compressedResponseLenPrefix [4]byte
			binary.BigEndian.PutUint32(compressedResponseLenPrefix[:], uint32(len(compressedResponseData)))

			listener, server := testCase.setupServer(
				&serverTracer,
				listener,
				func(respWriter http.ResponseWriter, req *http.Request) {
					_, _ = io.Copy(io.Discard, req.Body)
					_ = req.Body.Close()

					switch req.Header.Get("Content-Type") {
					case "application/proto":
						respWriter.Header().Set("Custom-Header", "ABC")
						respWriter.Header().Set("Content-Type", "application/json")
						respWriter.WriteHeader(http.StatusConflict)
						_, _ = respWriter.Write(responseData)
					case "application/connect+proto":
						respWriter.Header().Set("Custom-Header", "ABC")
						respWriter.Header().Set("Content-Type", "application/connect+proto")
						respWriter.Header().Set("Connect-Content-Encoding", "snappy")
						respWriter.WriteHeader(http.StatusOK)
						_, _ = respWriter.Write([]byte{0})
						_, _ = respWriter.Write(responseLenPrefix[:])
						_, _ = respWriter.Write(responseData)
						_, _ = respWriter.Write([]byte{3})
						_, _ = respWriter.Write(compressedResponseLenPrefix[:])
						_, _ = respWriter.Write(compressedResponseData)
					case "application/grpc-web+proto":
						respWriter.Header().Set("Custom-Header", "ABC")
						respWriter.Header().Set("Content-Type", "application/grpc-web")
						_, _ = respWriter.Write([]byte{0})
						_, _ = respWriter.Write(responseLenPrefix[:])
						_, _ = respWriter.Write(responseData)
						_, _ = respWriter.Write([]byte{128})
						_, _ = respWriter.Write(responseLenPrefix[:])
						_, _ = respWriter.Write(responseData)
					case "application/grpc+proto":
						respWriter.Header().Set("Custom-Header", "ABC")
						respWriter.Header().Set("Content-Type", "application/grpc")
						respWriter.Header().Set("Trailer", "Foo, Bar")
						respWriter.WriteHeader(http.StatusOK)
						_, _ = respWriter.Write([]byte{0})
						_, _ = respWriter.Write(responseLenPrefix[:])
						_, _ = respWriter.Write(responseData)
						respWriter.Header().Set(http.TrailerPrefix+"Foo", "One")
						respWriter.Header().Set(http.TrailerPrefix+"Bar", "Two")
					default:
						respWriter.Header().Set("Custom-Header", "ABC")
						respWriter.Header().Set("Content-Type", "application/json")
						respWriter.WriteHeader(http.StatusAccepted)
						_, _ = respWriter.Write(responseData)
					}
				},
			)
			go func() {
				err := server.Serve(listener)
				// not using require since we're not on the main goroutine
				assert.ErrorIs(t, err, http.ErrServerClosed)
			}()
			t.Cleanup(func() {
				_ = server.Close()
			})

			serverAddr := listener.Addr().String()
			requestData := []byte(`{"query": "I can haz cheezberder?"}`)
			var requestLenPrefix [4]byte
			binary.BigEndian.PutUint32(requestLenPrefix[:], uint32(len(requestData)))
			testCalls := []struct {
				name string
				// This trace's request must be complete as a client-side
				// request (and usable by an http.RoundTripper).
				// The trace's response does not need to indicate body
				// data and basically just reiterates the status code,
				// headers, and trailers that the handler above emits.
				// The trace events must be complete in terms of event
				// types and order, but only body data and end stream
				// events need their attributes populated.
				expectTrace *Trace
			}{
				{
					name: "connect-unary-post",
					expectTrace: &Trace{
						Request: &http.Request{
							Method: http.MethodPost,
							URL: &url.URL{
								Path: "/com.foo.Service/Bar",
							},
							Header: headers(
								"Content-Type", "application/proto",
							),
							Body: io.NopCloser(bytes.NewReader(requestData)),
						},
						Response: &http.Response{
							StatusCode: http.StatusConflict,
							Header: headers(
								"Custom-Header", "ABC",
								"Content-Type", "application/json",
							),
						},
						Events: []Event{
							&RequestStart{},
							&RequestBodyData{
								Len: uint64(len(requestData)),
							},
							&RequestBodyEnd{},
							&ResponseStart{},
							&ResponseBodyData{
								Len: uint64(len(responseData)),
							},
							&ResponseBodyEnd{},
						},
					},
				},
				{
					name: "connect-unary-get",
					expectTrace: &Trace{
						Request: &http.Request{
							Method: http.MethodGet,
							URL: &url.URL{
								Path:     "/com.foo.Service/Bar",
								RawQuery: "encoding=json&msg={}",
							},
							Header: headers(),
							Body:   http.NoBody,
						},
						Response: &http.Response{
							StatusCode: http.StatusAccepted,
							Header: headers(
								"Custom-Header", "ABC",
								"Content-Type", "application/json",
							),
						},
						Events: []Event{
							&RequestStart{},
							&RequestBodyEnd{},
							&ResponseStart{},
							&ResponseBodyData{
								Len: uint64(len(responseData)),
							},
							&ResponseBodyEnd{},
						},
					},
				},
				{
					name: "connect-stream",
					expectTrace: &Trace{
						Request: &http.Request{
							Method: http.MethodPost,
							URL: &url.URL{
								Path: "/com.foo.Service/Bar",
							},
							Header: headers(
								"Content-Type", "application/connect+proto",
							),
							Body: io.NopCloser(bytes.NewReader(concat(
								[]byte{0},
								requestLenPrefix[:],
								requestData,
							))),
						},
						Response: &http.Response{
							StatusCode: http.StatusOK,
							Header: headers(
								"Custom-Header", "ABC",
								"Content-Type", "application/connect+proto",
								"Connect-Content-Encoding", "snappy",
							),
						},
						Events: []Event{
							&RequestStart{},
							&RequestBodyData{
								Envelope: &Envelope{
									Flags: 0,
									Len:   uint32(len(requestData)),
								},
								Len: uint64(len(requestData)),
							},
							&RequestBodyEnd{},
							&ResponseStart{},
							&ResponseBodyData{
								Envelope: &Envelope{
									Flags: 0,
									Len:   uint32(len(responseData)),
								},
								Len: uint64(len(responseData)),
							},
							&ResponseBodyData{
								Envelope: &Envelope{
									Flags: 3,
									Len:   uint32(len(compressedResponseData)),
								},
								Len: uint64(len(compressedResponseData)),
							},
							&ResponseBodyEndStream{
								Content: string(responseData),
							},
							&ResponseBodyEnd{},
						},
					},
				},
				{
					name: "grpc-web",
					expectTrace: &Trace{
						Request: &http.Request{
							Method: http.MethodPost,
							URL: &url.URL{
								Path: "/com.foo.Service/Bar",
							},
							Header: headers(
								"Content-Type", "application/grpc-web+proto",
							),
							Body: io.NopCloser(bytes.NewReader(concat(
								[]byte{0},
								requestLenPrefix[:],
								requestData,
							))),
						},
						Response: &http.Response{
							StatusCode: http.StatusOK,
							Header: headers(
								"Custom-Header", "ABC",
								"Content-Type", "application/grpc-web",
							),
						},
						Events: []Event{
							&RequestStart{},
							&RequestBodyData{
								Envelope: &Envelope{
									Flags: 0,
									Len:   uint32(len(requestData)),
								},
								Len: uint64(len(requestData)),
							},
							&RequestBodyEnd{},
							&ResponseStart{},
							&ResponseBodyData{
								Envelope: &Envelope{
									Flags: 0,
									Len:   uint32(len(responseData)),
								},
								Len: uint64(len(responseData)),
							},
							&ResponseBodyData{
								Envelope: &Envelope{
									Flags: 128,
									Len:   uint32(len(responseData)),
								},
								Len: uint64(len(responseData)),
							},
							&ResponseBodyEndStream{
								Content: string(responseData),
							},
							&ResponseBodyEnd{},
						},
					},
				},
				{
					name: "grpc",
					expectTrace: &Trace{
						Request: &http.Request{
							Method: http.MethodPost,
							URL: &url.URL{
								Path: "/com.foo.Service/Bar",
							},
							Header: headers(
								"Content-Type", "application/grpc+proto",
								"TE", "trailers",
							),
							Body: io.NopCloser(bytes.NewReader(concat(
								[]byte{0},
								requestLenPrefix[:],
								requestData,
							))),
						},
						Response: &http.Response{
							StatusCode: http.StatusOK,
							Header: headers(
								"Custom-Header", "ABC",
								"Content-Type", "application/grpc",
							),
							Trailer: headers(
								"Foo", "One",
								"Bar", "Two",
							),
						},
						Events: []Event{
							&RequestStart{},
							&RequestBodyData{
								Envelope: &Envelope{
									Flags: 0,
									Len:   uint32(len(requestData)),
								},
								Len: uint64(len(requestData)),
							},
							&RequestBodyEnd{},
							&ResponseStart{},
							&ResponseBodyData{
								Envelope: &Envelope{
									Flags: 0,
									Len:   uint32(len(responseData)),
								},
								Len: uint64(len(responseData)),
							},
							&ResponseBodyEnd{},
						},
					},
				},
			}
			for _, testCall := range testCalls {
				testCall := testCall
				t.Run(testCall.name, func(t *testing.T) {
					t.Parallel()
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					req := testCall.expectTrace.Request.Clone(ctx)
					req.URL.Host = serverAddr
					req.URL.Scheme = "http"
					req.Header.Set(testCaseNameHeader, t.Name())
					clientTracer.Init(t.Name())
					serverTracer.Init(t.Name())
					resp, err := client.RoundTrip(req)
					require.NoError(t, err)
					_, _ = io.Copy(io.Discard, resp.Body)
					_ = resp.Body.Close()
					checkResponse(t, testCall.expectTrace.Response, resp)

					serverTrace, err := serverTracer.Await(ctx, t.Name())
					require.NoError(t, err)
					checkTrace(t, testCall.expectTrace, serverTrace)

					clientTrace, err := clientTracer.Await(ctx, t.Name())
					require.NoError(t, err)
					checkTrace(t, testCall.expectTrace, clientTrace)
				})
			}
		})
	}
}

func checkHeaders(t *testing.T, expected, actual http.Header) {
	t.Helper()
	require.GreaterOrEqual(t, len(actual), len(expected))
	for k, expectedVals := range expected {
		actualVals, ok := actual[k]
		require.True(t, ok, "actual metadata missing key %q", k)
		require.Equal(t, expectedVals, actualVals)
	}
}

func checkRequest(t *testing.T, expected, actual *http.Request) {
	t.Helper()
	require.NotNil(t, actual)
	require.Equal(t, expected.Method, actual.Method)
	require.Equal(t, expected.URL.Path, actual.URL.Path)
	require.Equal(t, expected.URL.RawQuery, actual.URL.RawQuery)
	checkHeaders(t, expected.Header, actual.Header)
	require.Empty(t, actual.Trailer)
}

func checkResponse(t *testing.T, expected, actual *http.Response) {
	t.Helper()
	require.NotNil(t, actual)
	require.Equal(t, expected.StatusCode, actual.StatusCode)
	checkHeaders(t, expected.Header, actual.Header)
	checkHeaders(t, expected.Trailer, actual.Trailer)
}

func checkMessageEnvelope(t *testing.T, expected, actual *Envelope) {
	t.Helper()
	if expected == nil {
		require.Nil(t, actual)
		return
	}
	require.NotNil(t, actual)
	require.Equal(t, expected.Flags, actual.Flags)
	require.Equal(t, expected.Len, actual.Len)
}

func checkTrace(t *testing.T, expected, actual *Trace) {
	t.Helper()
	require.NotNil(t, actual)
	require.NoError(t, actual.Err)
	require.Len(t, actual.Events, len(expected.Events),
		"want [%s]; got [%s]",
		eventTypes(expected.Events), eventTypes(actual.Events))
	for i, actualEvent := range actual.Events {
		expectedEvent := expected.Events[i]
		require.Equal(t, fmt.Sprintf("%T", expectedEvent), fmt.Sprintf("%T", actualEvent),
			"event #%d; want [%s]; got [%s]",
			i, eventTypes(expected.Events), eventTypes(actual.Events))
		switch actualEvent := actualEvent.(type) {
		case *RequestStart:
			require.Same(t, actual.Request, actualEvent.Request)
		case *RequestBodyData:
			expectedEvent := expectedEvent.(*RequestBodyData) //nolint:errcheck,forcetypeassert // already checked type above
			checkMessageEnvelope(t, expectedEvent.Envelope, actualEvent.Envelope)
			require.Equal(t, expectedEvent.Len, actualEvent.Len)
		case *ResponseStart:
			require.Same(t, actual.Response, actualEvent.Response)
		case *ResponseBodyData:
			expectedEvent := expectedEvent.(*ResponseBodyData) //nolint:errcheck,forcetypeassert // already checked type above
			checkMessageEnvelope(t, expectedEvent.Envelope, actualEvent.Envelope)
			require.Equal(t, expectedEvent.Len, actualEvent.Len)
		case *ResponseBodyEndStream:
			expectedEvent := expectedEvent.(*ResponseBodyEndStream) //nolint:errcheck,forcetypeassert // already checked type above
			require.Equal(t, expectedEvent.Content, actualEvent.Content)
		default:
			// we already checked event type above; nothing else to check for other events
		}
	}
	checkRequest(t, expected.Request, actual.Request)
	checkResponse(t, expected.Response, actual.Response)
}

func headers(kv ...string) http.Header {
	result := make(http.Header, len(kv)/2)
	for i := 0; i < len(kv); i += 2 {
		result.Add(kv[i], kv[i+1])
	}
	return result
}

func concat(data ...[]byte) []byte {
	result := data[0]
	for i := 1; i < len(data); i++ {
		result = append(result, data[i]...)
	}
	return result
}

func eventTypes(events []Event) string {
	types := make([]string, len(events))
	for i := range events {
		types[i] = fmt.Sprintf("%T", events[i])
	}
	return strings.Join(types, ",")
}
