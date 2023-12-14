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

package referenceclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/conformance/internal"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestRawRequestSender(t *testing.T) {
	t.Parallel()

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

	var requests sync.Map // map[string]chan *http.Request
	svr := httptest.NewServer(http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
		testCaseName := req.Header.Get("x-test-case-name")
		val, ok := requests.Load(testCaseName)
		reqChan, isChan := val.(chan *http.Request)
		if !ok || !isChan {
			respWriter.WriteHeader(http.StatusBadRequest)
		}
		reqCopy := req.Clone(req.Context())
		var data bytes.Buffer
		_, err := data.ReadFrom(req.Body)
		if err != nil {
			respWriter.WriteHeader(http.StatusInternalServerError)
		}
		reqCopy.Body = io.NopCloser(&data)
		reqChan <- reqCopy
		// no response means implicit 200 okay w/ no body
	}))
	t.Cleanup(svr.Close)

	testCases := []struct {
		name string
		req  *conformancev1.RawHTTPRequest
	}{
		{
			name: "basic",
			req: &conformancev1.RawHTTPRequest{
				Verb: http.MethodPost,
				Uri:  "/foo/bar.baz",
				Body: &conformancev1.RawHTTPRequest_Unary{
					Unary: &conformancev1.MessageContents{
						Data: &conformancev1.MessageContents_Binary{
							Binary: []byte{0, 1, 2, 3, 4, 5, 6, 7},
						},
					},
				},
				Headers: []*conformancev1.Header{
					{
						Name:  "Content-Type",
						Value: []string{"foo/bar"},
					},
					{
						Name:  "Content-Length",
						Value: []string{"8"},
					},
				},
			},
		},
		{
			name: "get-with-query-params",
			req: &conformancev1.RawHTTPRequest{
				Verb: http.MethodGet,
				Uri:  "/foo/bar.baz",
				RawQueryParams: []*conformancev1.Header{
					{
						Name:  "q",
						Value: []string{"a", "b", "c"},
					},
					{
						Name:  "x",
						Value: []string{"123"},
					},
				},
				Body: &conformancev1.RawHTTPRequest_Unary{
					Unary: &conformancev1.MessageContents{
						Data: &conformancev1.MessageContents_Text{
							Text: `{"foo": "bar", "baz": [0,1,2,3]}`,
						},
					},
				},
				Headers: []*conformancev1.Header{
					{
						Name:  "Content-Type",
						Value: []string{"foo/bar"},
					},
					{
						Name:  "X-Custom-Header",
						Value: []string{"abc", "def", "xyz"},
					},
				},
			},
		},
		{
			name: "encoded-query-params",
			req: &conformancev1.RawHTTPRequest{
				Verb: http.MethodPut,
				Uri:  "/foo/bar.baz",
				EncodedQueryParams: []*conformancev1.RawHTTPRequest_EncodedQueryParam{
					{
						Name: "q",
						Value: &conformancev1.MessageContents{
							Data: &conformancev1.MessageContents_Binary{
								Binary: []byte{0, 1, 2, 3, 4, 5, 6, 7},
							},
							Compression: conformancev1.Compression_COMPRESSION_GZIP,
						},
						Base64Encode: true,
					},
					{
						Name: "x",
						Value: &conformancev1.MessageContents{
							Data: &conformancev1.MessageContents_Text{
								Text: `{"foo": "bar"}`,
							},
						},
					},
				},
				Body: &conformancev1.RawHTTPRequest_Unary{
					Unary: &conformancev1.MessageContents{
						Data: &conformancev1.MessageContents_Text{
							Text: `{"foo": "bar", "baz": [0,1,2,3]}`,
						},
					},
				},
				Headers: []*conformancev1.Header{
					{
						Name:  "Content-Type",
						Value: []string{"foo/bar"},
					},
					{
						Name:  "X-Custom-Header",
						Value: []string{"abc", "def", "xyz"},
					},
				},
			},
		},
		{
			name: "mix-of-query-params",
			req: &conformancev1.RawHTTPRequest{
				Verb: http.MethodGet,
				Uri:  "/foo/bar.baz?q=q&x=456",
				RawQueryParams: []*conformancev1.Header{
					{
						Name:  "q",
						Value: []string{"a", "b", "c"},
					},
					{
						Name:  "x",
						Value: []string{"123"},
					},
				},
				EncodedQueryParams: []*conformancev1.RawHTTPRequest_EncodedQueryParam{
					{
						Name: "q",
						Value: &conformancev1.MessageContents{
							Data: &conformancev1.MessageContents_Binary{
								Binary: []byte{0, 1, 2, 3, 4, 5, 6, 7},
							},
							Compression: conformancev1.Compression_COMPRESSION_GZIP,
						},
						Base64Encode: true,
					},
					{
						Name: "x",
						Value: &conformancev1.MessageContents{
							Data: &conformancev1.MessageContents_Text{
								Text: `{"foo": "bar"}`,
							},
						},
					},
				},
				Body: &conformancev1.RawHTTPRequest_Unary{
					Unary: &conformancev1.MessageContents{
						Data: &conformancev1.MessageContents_Text{
							Text: `{"foo": "bar", "baz": [0,1,2,3]}`,
						},
					},
				},
				Headers: []*conformancev1.Header{
					{
						Name:  "Content-Type",
						Value: []string{"application/json"},
					},
				},
			},
		},
		{
			name: "empty-body",
			req: &conformancev1.RawHTTPRequest{
				Verb: http.MethodGet,
				Uri:  "/foo/bar.baz",
			},
		},
		{
			name: "stream-body",
			req: &conformancev1.RawHTTPRequest{
				Verb: http.MethodGet,
				Uri:  "/foo/bar.baz",
				RawQueryParams: []*conformancev1.Header{
					{
						Name:  "q",
						Value: []string{"a", "b", "c"},
					},
					{
						Name:  "x",
						Value: []string{"123"},
					},
				},
				Body: &conformancev1.RawHTTPRequest_Stream{
					Stream: &conformancev1.StreamContents{
						Items: []*conformancev1.StreamContents_StreamItem{
							{
								Payload: &conformancev1.MessageContents{
									Data: &conformancev1.MessageContents_Text{
										Text: `{"foo": "bar", "baz": [0,1,2,3]}`,
									},
								},
							},
							{
								Flags: 0x01,
								Payload: &conformancev1.MessageContents{
									Data: &conformancev1.MessageContents_BinaryMessage{
										BinaryMessage: msgPayload,
									},
									Compression: conformancev1.Compression_COMPRESSION_DEFLATE,
								},
							},
							{
								Flags:  0x3,
								Length: proto.Uint32(8),
								Payload: &conformancev1.MessageContents{
									Data: &conformancev1.MessageContents_Binary{
										Binary: []byte{0, 1, 2, 3, 4, 5, 6, 7},
									},
								},
							},
						},
					},
				},
				Headers: []*conformancev1.Header{
					{
						Name:  "Content-Type",
						Value: []string{"foo/bar"},
					},
					{
						Name:  "X-Custom-Header",
						Value: []string{"abc", "def", "xyz"},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			reqChan := make(chan *http.Request, 1)
			testCaseName := t.Name()
			requests.Store(testCaseName, reqChan)

			sender := &rawRequestSender{
				transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
					req.Header.Set("x-test-case-name", testCaseName)
					transport := &http.Transport{DisableCompression: true}
					return transport.RoundTrip(req)
				}),
				rawRequest: testCase.req,
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, svr.URL+"/some.random.endpoint/", strings.NewReader("foobar"))
			require.NoError(t, err)
			resp, err := sender.RoundTrip(req)
			require.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()
			require.Equal(t, http.StatusOK, resp.StatusCode)

			requests.Delete(testCaseName)
			select {
			case req = <-reqChan:
			default:
				t.Fatal("server did not send request to channel")
			}

			// First we check the URI path
			expectedURL, err := url.ParseRequestURI(testCase.req.Uri)
			require.NoError(t, err)
			assert.Equal(t, expectedURL.Path, req.URL.Path)

			// Then query params
			expectedInitialQueryParams := expectedURL.Query()
			expectedQueryParamCounts := map[string]int{}
			for k, v := range expectedInitialQueryParams {
				expectedQueryParamCounts[k] = len(v)
			}
			for _, param := range testCase.req.RawQueryParams {
				expectedQueryParamCounts[param.Name] += len(param.Value)
			}
			for _, param := range testCase.req.EncodedQueryParams {
				expectedQueryParamCounts[param.Name]++
			}

			actualQueryParams := req.URL.Query()
			actualQueryParamCounts := map[string]int{}
			for k, v := range actualQueryParams {
				actualQueryParamCounts[k] = len(v)
			}

			assert.Equal(t, expectedQueryParamCounts, actualQueryParamCounts)
			for k, expectedVals := range expectedInitialQueryParams {
				actualVals := actualQueryParams[k]
				if len(actualVals) > len(expectedVals) {
					actualVals = actualVals[:len(expectedVals)]
				}
				assert.Equal(t, expectedVals, actualVals, "inline query param values for %q", k)
			}
			rawParams := make(map[string][]string, len(testCase.req.RawQueryParams))
			for _, param := range testCase.req.RawQueryParams {
				rawParams[param.Name] = param.Value
				actualVals := actualQueryParams[param.Name]
				// remove any initial expected values that were in the URI string
				initVals := expectedInitialQueryParams[param.Name]
				if len(actualVals) > len(initVals) {
					actualVals = actualVals[len(initVals):]
				} else {
					actualVals = nil
				}
				if len(actualVals) > len(param.Value) {
					actualVals = actualVals[:len(param.Value)]
				}
				assert.Equal(t, param.Value, actualVals, "raw query param values for %q", param.Name)
			}
			for _, param := range testCase.req.EncodedQueryParams {
				actualVals := actualQueryParams[param.Name]
				// remove any initial expected values that were in the URI string
				initVals := expectedInitialQueryParams[param.Name]
				if len(actualVals) > len(initVals) {
					actualVals = actualVals[len(initVals):]
				} else {
					actualVals = nil
				}
				rawVals := rawParams[param.Name]
				if len(actualVals) > len(rawVals) {
					actualVals = actualVals[len(rawVals):]
				} else {
					actualVals = nil
				}
				var buf bytes.Buffer
				err := internal.WriteRawMessageContents(param.Value, &buf)
				require.NoError(t, err)
				var expectedVals []string
				if param.Base64Encode {
					expectedVals = []string{base64.URLEncoding.EncodeToString(buf.Bytes())}
				} else {
					expectedVals = []string{buf.String()}
				}
				assert.Equal(t, expectedVals, actualVals, "encoded query param values for %q", param.Name)
			}

			// Then HTTP method/verb
			assert.Equal(t, testCase.req.Verb, req.Method)

			// Then headers
			req.Header.Del("x-test-case-name") // added by the round tripper above; not in raw request
			req.Header.Del("user-agent")       // added by http.Transport
			expectedHeaders := http.Header{}
			internal.AddHeaders(testCase.req.Headers, expectedHeaders)
			assert.Equal(t, expectedHeaders, req.Header)

			// Finally, the body
			var expected, actual bytes.Buffer
			_, err = io.Copy(&actual, req.Body)
			require.NoError(t, err)
			switch contents := testCase.req.Body.(type) {
			case *conformancev1.RawHTTPRequest_Unary:
				err = internal.WriteRawMessageContents(contents.Unary, &expected)
			case *conformancev1.RawHTTPRequest_Stream:
				err = internal.WriteRawStreamContents(contents.Stream, &expected)
			}
			require.NoError(t, err)
			assert.Equal(t, expected.Bytes(), actual.Bytes())
		})
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
