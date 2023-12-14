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

package referenceserver

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/conformance/internal"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1/conformancev1connect"
	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestRawResponseRecorder(t *testing.T) {
	t.Parallel()

	_, handler := conformancev1connect.NewConformanceServiceHandler(
		conformancev1connect.UnimplementedConformanceServiceHandler{},
		connect.WithInterceptors(rawResponseRecorder{}),
	)
	svr := httptest.NewUnstartedServer(rawResponder(handler, internal.NewPrinter(&bytes.Buffer{})))
	svr.EnableHTTP2 = true // so we can test bidi stream
	svr.StartTLS()
	t.Cleanup(svr.Close)

	invokeUnary := func(ctx context.Context, rawResponse *conformancev1.RawHTTPResponse, client conformancev1connect.ConformanceServiceClient) {
		_, _ = client.Unary(ctx, connect.NewRequest(&conformancev1.UnaryRequest{
			RequestData: []byte{0, 1, 2, 3, 4},
			ResponseDefinition: &conformancev1.UnaryResponseDefinition{
				Response: &conformancev1.UnaryResponseDefinition_ResponseData{
					ResponseData: []byte{4, 3, 2, 1, 0},
				},
				RawResponse: rawResponse,
			},
		}))
	}
	invokeClientStream := func(ctx context.Context, rawResponse *conformancev1.RawHTTPResponse, client conformancev1connect.ConformanceServiceClient) {
		stream := client.ClientStream(ctx)
		_ = stream.Send(&conformancev1.ClientStreamRequest{
			RequestData: []byte{0, 1, 2, 3, 4},
			ResponseDefinition: &conformancev1.UnaryResponseDefinition{
				Response: &conformancev1.UnaryResponseDefinition_ResponseData{
					ResponseData: []byte{4, 3, 2, 1, 0},
				},
				RawResponse: rawResponse,
			},
		})
		_ = stream.Send(&conformancev1.ClientStreamRequest{
			RequestData: []byte{5, 6, 7, 8, 9},
		})
		_ = stream.Send(&conformancev1.ClientStreamRequest{
			RequestData: []byte{0, 1, 2, 3, 4},
		})
		_, _ = stream.CloseAndReceive()
	}
	invokeServerStream := func(ctx context.Context, rawResponse *conformancev1.RawHTTPResponse, client conformancev1connect.ConformanceServiceClient) {
		stream, err := client.ServerStream(ctx, connect.NewRequest(&conformancev1.ServerStreamRequest{
			RequestData: []byte{0, 1, 2, 3, 4},
			ResponseDefinition: &conformancev1.StreamResponseDefinition{
				ResponseData: [][]byte{
					{0, 1, 2, 3},
					{4, 5, 6, 7},
					{8, 9, 0, 1},
				},
				RawResponse: rawResponse,
			},
		}))
		if err != nil {
			return
		}
		defer func() {
			_ = stream.Close()
		}()
		for stream.Receive() {
			// exhaust stream
		}
	}
	invokeBidiStream := func(ctx context.Context, rawResponse *conformancev1.RawHTTPResponse, client conformancev1connect.ConformanceServiceClient) {
		stream := client.BidiStream(ctx)
		done := make(chan struct{})
		go func() {
			defer close(done)
			defer func() {
				_ = stream.CloseResponse()
			}()
			for {
				// exhaust stream
				_, err := stream.Receive()
				if err != nil {
					return
				}
			}
		}()
		_ = stream.Send(&conformancev1.BidiStreamRequest{
			RequestData: []byte{0, 1, 2, 3, 4},
			ResponseDefinition: &conformancev1.StreamResponseDefinition{
				ResponseData: [][]byte{
					{0, 1, 2, 3},
					{4, 5, 6, 7},
					{8, 9, 0, 1},
				},
				RawResponse: rawResponse,
			},
		})
		_ = stream.Send(&conformancev1.BidiStreamRequest{
			RequestData: []byte{5, 6, 7, 8, 9},
		})
		_ = stream.Send(&conformancev1.BidiStreamRequest{
			RequestData: []byte{0, 1, 2, 3, 4},
		})
		_ = stream.CloseRequest()
		<-done // make sure we've finished exhausting response stream
	}

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

	testCases := []struct {
		name string
		resp *conformancev1.RawHTTPResponse
	}{
		{
			name: "simple",
			resp: &conformancev1.RawHTTPResponse{
				StatusCode: 200,
				Headers: []*conformancev1.Header{
					{
						Name:  "foo",
						Value: []string{"bar", "baz"},
					},
					{
						Name:  "content-type",
						Value: []string{"foo/bar"},
					},
				},
				Body: &conformancev1.RawHTTPResponse_Unary{
					Unary: &conformancev1.MessageContents{
						Data: &conformancev1.MessageContents_Text{
							Text: `{"foo":"bar"}`,
						},
					},
				},
				Trailers: []*conformancev1.Header{
					{
						Name:  "x",
						Value: []string{"123", "456"},
					},
				},
			},
		},
		{
			name: "empty",
			resp: &conformancev1.RawHTTPResponse{
				StatusCode: 404,
			},
		},
		{
			name: "stream body",
			resp: &conformancev1.RawHTTPResponse{
				StatusCode: 505,
				Headers: []*conformancev1.Header{
					{
						Name:  "foo",
						Value: []string{"bar", "baz"},
					},
					{
						Name:  "content-type",
						Value: []string{"foo/bar"},
					},
				},
				Body: &conformancev1.RawHTTPResponse_Stream{
					Stream: &conformancev1.StreamContents{
						Items: []*conformancev1.StreamContents_StreamItem{
							{
								Length: proto.Uint32(10),
								Payload: &conformancev1.MessageContents{
									Data: &conformancev1.MessageContents_Binary{
										Binary: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
									},
								},
							},
							{
								Flags: 1,
								Payload: &conformancev1.MessageContents{
									Data: &conformancev1.MessageContents_BinaryMessage{
										BinaryMessage: msgPayload,
									},
									Compression: conformancev1.Compression_COMPRESSION_SNAPPY,
								},
							},
							{
								Flags: 0x80,
								Payload: &conformancev1.MessageContents{
									Data: &conformancev1.MessageContents_Text{
										Text: `{"error": "not_found", "trailers": {"a": ["1", "2", "3"], "b": ["xyz"]}}"`,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	invokers := []struct {
		name   string
		invoke func(ctx context.Context, rawResponse *conformancev1.RawHTTPResponse, client conformancev1connect.ConformanceServiceClient)
	}{
		{
			name:   "unary",
			invoke: invokeUnary,
		},
		{
			name:   "client stream",
			invoke: invokeClientStream,
		},
		{
			name:   "server stream",
			invoke: invokeServerStream,
		},
		{
			name:   "bidi stream",
			invoke: invokeBidiStream,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			for _, invoker := range invokers {
				invoker := invoker
				t.Run(invoker.name, func(t *testing.T) {
					t.Parallel()

					var err error
					var resp *http.Response
					var body bytes.Buffer
					testCaseName := t.Name()
					transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
						req.Header.Set("x-test-case-name", testCaseName)
						resp, err = svr.Client().Transport.RoundTrip(req)
						if err != nil {
							return nil, err
						}
						resp.Body = &captureBody{buf: &body, r: resp.Body}
						return resp, nil
					})

					client := conformancev1connect.NewConformanceServiceClient(&http.Client{Transport: transport}, svr.URL)
					invoker.invoke(context.Background(), testCase.resp, client)

					require.NoError(t, err)

					// First we check the status code
					assert.Equal(t, int(testCase.resp.StatusCode), resp.StatusCode)

					// Then headers and trailers
					expectedHeaders := http.Header{}
					internal.AddHeaders(testCase.resp.Headers, expectedHeaders)
					assert.Equal(t, expectedHeaders, resp.Header)

					expectedTrailers := http.Header{}
					if resp.Trailer == nil {
						assert.Empty(t, testCase.resp.Trailers)
					} else {
						internal.AddHeaders(testCase.resp.Trailers, expectedTrailers)
						assert.Equal(t, expectedTrailers, resp.Trailer)
					}

					// Finally, the body
					expectedBody := bytes.NewBuffer([]byte{})
					switch contents := testCase.resp.Body.(type) {
					case *conformancev1.RawHTTPResponse_Unary:
						err = internal.WriteRawMessageContents(contents.Unary, expectedBody)
					case *conformancev1.RawHTTPResponse_Stream:
						err = internal.WriteRawStreamContents(contents.Stream, expectedBody)
					}
					require.NoError(t, err)
					assert.Equal(t, expectedBody.Bytes(), body.Bytes())
				})
			}
		})
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type captureBody struct {
	buf *bytes.Buffer
	r   io.ReadCloser
}

func (c *captureBody) Read(data []byte) (n int, err error) {
	n, err = c.r.Read(data)
	// save all bytes read to c.buf
	c.buf.Write(data[:n])
	return n, err
}

func (c *captureBody) Close() error {
	// make sure reader is drained before closing
	// so that we capture entire body into c.buf
	_, _ = io.Copy(c.buf, c.r)
	return c.r.Close()
}
