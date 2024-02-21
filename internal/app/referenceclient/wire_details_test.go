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
	"context"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/conformance/internal"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1/conformancev1connect"
	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExamineConnectError(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name             string
		endStream        string
		expectedFeedback []string
	}{
		{
			name: "minimal",
			endStream: `
				{
					"code": "internal"
				}`,
		},
		{
			name: "with message",
			endStream: `
				{
					"code": "internal",
					"message": "blah blah blah"
				}`,
		},
		{
			name: "with details",
			endStream: `
				{
					"code": "internal",
					"details": [
						{
							"type": "blah.blah.Blah",
							"value": "abcdefgh"
						}
					]
				}`,
		},
		{
			name: "complete",
			endStream: `
				{
					"code": "internal",
					"message": "blah blah blah",
					"details": [
						{
							"type": "blah.blah.Blah",
							"value": "abcdefgh"
						}
					]
				}`,
		},
		{
			name:      "empty",
			endStream: `{}`,
			expectedFeedback: []string{
				`connect error JSON: missing required key "code"`,
			},
		},
		{
			name:      "only message",
			endStream: `{"message": "abc"}`,
			expectedFeedback: []string{
				`connect error JSON: missing required key "code"`,
			},
		},
		{
			name:      "incorrect type for code",
			endStream: `{"code": 1, "message": "abc"}`,
			expectedFeedback: []string{
				`connect error JSON: json: cannot unmarshal number into Go struct field connectError.code of type string`,
			},
		},
		{
			name:      "incorrect type for message",
			endStream: `{"code": "unavailable", "message": 12345}`,
			expectedFeedback: []string{
				`connect error JSON: json: cannot unmarshal number into Go struct field connectError.message of type string`,
			},
		},
		{
			name: "incorrect type for details",
			endStream: `
				{
					"code": "unavailable",
					"details": {
						"type": "blah.blah.Blah",
						"value": "abcdefgh"
					}
				}`,
			expectedFeedback: []string{
				`connect error JSON: json: cannot unmarshal object into Go struct field connectError.details of type []json.RawMessage`,
			},
		},
		{
			name:      "unrecognized code",
			endStream: `{"code": "abc"}`,
			expectedFeedback: []string{
				`connect error JSON: value for key "code" is not a recognized error code name: "abc"`,
			},
		},
		{
			name: "incorrect case",
			endStream: `
				{
					"Code": "unavailable",
					"Message": "abc",
					"Details": [
						{
							"Type": "google.protobuf.Empty",
							"Value": "",
							"Debug": {}
						}
					]
				}`,
			expectedFeedback: []string{
				`connect error JSON: invalid key "Code"`,
				`connect error JSON: invalid key "Details"`,
				`connect error JSON: invalid key "Message"`,
				`connect error JSON: missing required key "code"`,
			},
		},
		{
			name: "incorrect detail type", // uses type URL instead of type name
			endStream: `
				{
					"code": "internal",
					"details": [
						{
							"type": "foo.com/blah.blah.Blah",
							"value": "abcdefgh"
						}
					]
				}`,
			expectedFeedback: []string{
				`connect error JSON: details[0]: value for key "type", "foo.com/blah.blah.Blah", is not a valid type name`,
			},
		},
		{
			name: "nulls",
			endStream: `
				{
					"code": null,
					"message": null,
					"details": null
				}`,
			expectedFeedback: []string{
				`connect error JSON: value for key "code" is a <nil> instead of a string`,
				`connect error JSON: value for key "details" is a <nil> instead of a slice`,
				`connect error JSON: value for key "message" is a <nil> instead of a string`,
			},
		},
		{
			name: "null in details",
			endStream: `
				{
					"code": "already_exists",
					"message": "abc",
					"details": [null]
				}`,
			expectedFeedback: []string{
				`connect error JSON: details[0]: expecting an object but got <nil>`,
			},
		},
		{
			name: "empty detail",
			endStream: `
				{
					"code": "already_exists",
					"message": "abc",
					"details": [{}]
				}`,
			expectedFeedback: []string{
				`connect error JSON: details[0]: missing required key "type"`,
				`connect error JSON: details[0]: missing required key "value"`,
			},
		},
		{
			name: "detail has invalid value",
			endStream: `
				{
					"code": "already_exists",
					"message": "abc",
					"details": [{
						"type": "foo.bar.Baz",
						"value": "><@#$%^"
					}]
				}`,
			expectedFeedback: []string{
				`connect error JSON: details[0]: value for key "value", "><@#$%^", is not valid unpadded base64-encoding: illegal base64 data at input byte 0`,
			},
		},
		{
			name: "detail uses wrong base64 alphabet",
			endStream: `
				{
					"code": "already_exists",
					"message": "abc",
					"details": [{
						"type": "foo.bar.Baz",
						"value": "abc-efg"
					}]
				}`,
			expectedFeedback: []string{
				`connect error JSON: details[0]: value for key "value", "abc-efg", is not valid unpadded base64-encoding: illegal base64 data at input byte 3`,
			},
		},
		{
			name: "detail uses padding",
			endStream: `
				{
					"code": "already_exists",
					"message": "abc",
					"details": [{
						"type": "foo.bar.Baz",
						"value": "abcdef="
					}]
				}`,
			expectedFeedback: []string{
				`connect error JSON: details[0]: value for key "value", "abcdef=", is not valid unpadded base64-encoding: illegal base64 data at input byte 6`,
			},
		},
		{
			name: "detail debug disagrees with value",
			// value encoding has a file with path "bar/foo.proto"
			endStream: `
				{
					"code": "already_exists",
					"message": "abc",
					"details": [
						{
							"type": "google.protobuf.FileDescriptorSet",
							"value": "Cg8KDWJhci9mb28ucHJvdG8",
							"debug": {"file": [{"name": "foo.proto"}]}
						}
					]
				}`,
			expectedFeedback: []string{
				`connect error JSON: details[0]: debug data does not match value: - value, + debug
  (*descriptorpb.FileDescriptorSet)(Inverse(protocmp.Transform, protocmp.Message{
  	"@type": s"google.protobuf.FileDescriptorSet",
  	"file": []protocmp.Message{
  		{
  			"@type": s"google.protobuf.FileDescriptorProto",
- 			"name":  string("bar/foo.proto"),
+ 			"name":  string("foo.proto"),
  		},
  	},
  }))`,
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			svr := httptest.NewServer(http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
				respWriter.Header().Set("Content-Type", "application/json")
				respWriter.WriteHeader(http.StatusBadRequest)
				_, _ = respWriter.Write([]byte(testCase.endStream))
			}))
			t.Cleanup(svr.Close)
			client := conformancev1connect.NewConformanceServiceClient(
				&http.Client{Transport: newWireCaptureTransport(svr.Client().Transport, nil)},
				svr.URL,
			)
			ctx := withWireCapture(context.Background())
			req := connect.NewRequest(&conformancev1.UnaryRequest{})
			req.Header().Set("x-test-case-name", "foo") // needed to enable tracing
			_, err := client.Unary(ctx, req)
			require.Error(t, err)
			printer := &internal.SimplePrinter{}
			examineWireDetails(ctx, printer)
			if len(testCase.expectedFeedback) == 0 {
				assert.Empty(t, printer.Messages)
			} else {
				for i := range printer.Messages {
					// Printer output keeps trailing newlines. And some messages come
					// from using cmp.Diff, which uses non-breaking spaces for formatting
					// instead of regular spaces. So we "clean up" the messages so we can
					// easily compare them to the expectations above.
					printer.Messages[i] = strings.ReplaceAll(
						strings.TrimSuffix(printer.Messages[i], "\n"),
						"\u00a0", // non-breaking space
						" ",      // regular space
					)
				}
				assert.Empty(t, cmp.Diff(testCase.expectedFeedback, printer.Messages))
			}
		})
	}
}

func TestExamineConnectEndStream(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name             string
		endStream        string
		expectedFeedback []string
	}{
		{
			name: "minimal",
			endStream: `
				{}`,
		},
		{
			name: "minimal with metadata",
			endStream: `
				{"metadata":{}}`,
		},
		{
			name: "minimal with error",
			endStream: `
				{"error":{"code":"canceled"}}`,
		},
		{
			name: "nulls",
			endStream: `
				{"error":null, "metadata":null}`,
			expectedFeedback: []string{
				`connect end stream JSON: value for key "error" is a <nil> instead of a map/object`,
				`connect end stream JSON: value for key "metadata" is a <nil> instead of a map/object`,
			},
		},
		{
			name: "empty error",
			endStream: `
				{"error":{}}`,
			expectedFeedback: []string{
				`connect error JSON: missing required key "code"`,
			},
		},
		{
			name: "invalid error", // verifies use of examineConnectError (see above for more test cases of that)
			endStream: `
				{"error":{"code": null}}`,
			expectedFeedback: []string{
				`connect error JSON: value for key "code" is a <nil> instead of a string`,
			},
		},
		{
			name: "invalid metadata",
			endStream: `
				{"metadata":[{"key": "header", "value": "abc"}, {"key": "header", "value": "abc"}]}`,
			expectedFeedback: []string{
				`connect end stream JSON: json: cannot unmarshal array into Go struct field connectEndStream.metadata of type map[string][]string`,
			},
		},
		{
			name: "duplicate metadata key",
			endStream: `
				{"metadata":{
					"header": ["abc", "def"],
					"header": ["xyz"]
				}}`,
			expectedFeedback: []string{
				`connect end stream JSON: metadata: contains duplicate key "header"`,
			},
		},
		{
			name: "invalid metadata key",
			endStream: `
				{"metadata":{
					"abc{'-|-'}def": ["abc", "def"]
				}}`,
			expectedFeedback: []string{
				`connect end stream JSON: metadata["abc{'-|-'}def"]: entry key is not a valid HTTP field name`,
			},
		},
		{
			name: "invalid metadata value type",
			endStream: `
				{"metadata":{
					"abc-def": ["abc", null]
				}}`,
			expectedFeedback: []string{
				`connect end stream JSON: metadata["abc-def"]: value #2 is a <nil> instead of a string`,
			},
		},
		{
			name: "invalid metadata value",
			endStream: `
				{"metadata":{
					"abc-def": ["abc", "def", "abc\u0000def\rxyz"]
				}}`,
			expectedFeedback: []string{
				// strange that it reports the type of the null as []any instead of <nil>...
				`connect end stream JSON: metadata["abc-def"]: value #3 is not a valid HTTP field value: "abc\x00def\rxyz"`,
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			svr := httptest.NewServer(http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
				respWriter.Header().Set("Content-Type", "application/connect+proto")
				_, _ = respWriter.Write([]byte{2}) // just the end-stream flag
				var size [4]byte
				binary.BigEndian.PutUint32(size[:], uint32(len(testCase.endStream)))
				_, _ = respWriter.Write(size[:])
				_, _ = respWriter.Write([]byte(testCase.endStream))
			}))
			t.Cleanup(svr.Close)
			client := conformancev1connect.NewConformanceServiceClient(
				&http.Client{Transport: newWireCaptureTransport(svr.Client().Transport, nil)},
				svr.URL,
			)
			ctx := withWireCapture(context.Background())
			stream := client.ClientStream(ctx)
			stream.RequestHeader().Set("x-test-case-name", "foo") // needed to enable tracing
			_, err := stream.CloseAndReceive()
			require.Error(t, err)
			printer := &internal.SimplePrinter{}
			examineWireDetails(ctx, printer)
			if len(testCase.expectedFeedback) == 0 {
				assert.Empty(t, printer.Messages)
			} else {
				for i := range printer.Messages {
					// Printer output keeps trailing newlines. And some messages come
					// from using cmp.Diff, which uses non-breaking spaces for formatting
					// instead of regular spaces. So we "clean up" the messages so we can
					// easily compare them to the expectations above.
					printer.Messages[i] = strings.ReplaceAll(
						strings.TrimSuffix(printer.Messages[i], "\n"),
						"\u00a0", // non-breaking space
						" ",      // regular space
					)
				}
				assert.Empty(t, cmp.Diff(testCase.expectedFeedback, printer.Messages))
			}
		})
	}
}

func TestExamineGRPCEndStream(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name             string
		endStream        string
		expectedFeedback []string
	}{
		{
			name: "correct",
			endStream: "grpc-status: 6\r\n" +
				"grpc-message: foo\r\n" +
				"blah-blah: foobar\r\n",
		},
		{
			name: "allowed special chars in key",
			endStream: "grpc-status: 6\r\n" +
				"grpc-message: foo\r\n" +
				"`~blah.blah-#blah|blah%blah's~`: foobar\r\n",
		},
		{
			name: "allowed special chars in value",
			endStream: "grpc-status: 6\r\n" +
				"grpc-message: foo\r\n" +
				"blah-blah: {foobar} \"baz\" 🤷🦸\r\n",
		},
		{
			name: "mixed case",
			endStream: "Grpc-Status: 6\r\n" +
				"Grpc-Message: foo\r\n" +
				"Blah-Blah: foobar\r\n",
			expectedFeedback: []string{
				`grpc-web trailers include non-lower-case field key: "Grpc-Status"`,
				`grpc-web trailers include non-lower-case field key: "Grpc-Message"`,
				`grpc-web trailers include non-lower-case field key: "Blah-Blah"`,
			},
		},
		{
			name: "no trailing crlf",
			endStream: "grpc-status: 6\r\n" +
				"grpc-message: foo\r\n" +
				"blah-blah: foobar",
			expectedFeedback: []string{
				"grpc-web trailers should end with CRLF but does not",
			},
		},
		{
			name: "key without value",
			endStream: "grpc-status: 6\r\n" +
				"grpc-message: foo\r\n" +
				"blah-blah\r\n",
			expectedFeedback: []string{
				`grpc-web trailers include invalid field (missing colon): "blah-blah"`,
			},
		},
		{
			name: "extra end line",
			endStream: "grpc-status: 6\r\n" +
				"grpc-message: foo\r\n" +
				"blah-blah: foobar\r\n" +
				"\r\n",
			expectedFeedback: []string{
				"grpc-web trailers ends in extra blank line",
			},
		},
		{
			name: "blank line at start",
			endStream: "\r\n" +
				"grpc-status: 6\r\n" +
				"grpc-message: foo\r\n" +
				"blah-blah: foobar\r\n",
			expectedFeedback: []string{
				"grpc-web trailers include blank lines",
			},
		},
		{
			name: "blank line in middle",
			endStream: "grpc-status: 6\r\n" +
				"grpc-message: foo\r\n" +
				"\r\n" +
				"blah-blah: foobar\r\n",
			expectedFeedback: []string{
				"grpc-web trailers include blank lines",
			},
		},
		{
			name: "obsolete line folding",
			endStream: "grpc-status: 6\r\n" +
				"grpc-message: foo\r\n" +
				" \tbar\r\n" +
				" \tbaz\r\n" +
				"blah-blah: foobar\r\n",
			expectedFeedback: []string{
				"grpc-web trailers use obsolete line-folding",
			},
		},
		{
			name: "invalid char in value",
			endStream: "grpc-status: 6\r\n" +
				"grpc-message: foo\r\n" +
				"blah-blah: foo\x00bar\r\n",
			expectedFeedback: []string{
				`grpc-web trailers include invalid field; value contains invalid characters: "blah-blah: foo\x00bar"`,
			},
		},
		{
			name: "invalid char in key",
			endStream: "grpc-status: 6\r\n" +
				"grpc-message: foo\r\n" +
				"blah[blah]blah: foo bar\r\n",
			expectedFeedback: []string{
				`grpc-web trailers include invalid field; name contains invalid characters: "blah[blah]blah: foo bar"`,
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			svr := httptest.NewServer(http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
				respWriter.Header().Set("Content-Type", "application/grpc-web")
				_, _ = respWriter.Write([]byte{128}) // just the end-stream flag
				var size [4]byte
				binary.BigEndian.PutUint32(size[:], uint32(len(testCase.endStream)))
				_, _ = respWriter.Write(size[:])
				_, _ = respWriter.Write([]byte(testCase.endStream))
			}))
			t.Cleanup(svr.Close)
			client := conformancev1connect.NewConformanceServiceClient(
				&http.Client{Transport: newWireCaptureTransport(svr.Client().Transport, nil)},
				svr.URL,
				connect.WithGRPCWeb(),
			)
			ctx := withWireCapture(context.Background())
			req := connect.NewRequest(&conformancev1.UnaryRequest{})
			req.Header().Set("x-test-case-name", "foo") // needed to enable tracing
			_, err := client.Unary(ctx, req)
			require.Error(t, err)
			printer := &internal.SimplePrinter{}
			examineWireDetails(ctx, printer)
			if len(testCase.expectedFeedback) == 0 {
				assert.Equal(t, connect.CodeAlreadyExists, connect.CodeOf(err), "unexpected error: %v", err)
				assert.Empty(t, printer.Messages)
			} else {
				// When there's feedback, the connect-go client may complain about the end-stream message
				// and report a different code.
				assert.Contains(t,
					[]connect.Code{connect.CodeAlreadyExists, connect.CodeInternal, connect.CodeUnknown},
					connect.CodeOf(err),
					"unexpected error: %v", err,
				)
				for i := range printer.Messages {
					printer.Messages[i] = strings.TrimSuffix(printer.Messages[i], "\n")
				}
				assert.Empty(t, cmp.Diff(testCase.expectedFeedback, printer.Messages))
			}
		})
	}
}

func TestCheckNoDuplicateKeys(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name:  "okay",
			input: `{ "a": true, "b": false }`,
		},
		{
			name:        "duplicate key",
			input:       `{ "a": true, "a": false }`,
			expectedErr: `contains duplicate key "a"`,
		},
		{
			name:  "top-level scalar",
			input: `123.456`,
		},
		{
			name:  "top-level array",
			input: `[1, 2, 3, 4]`,
		},
		{
			name:        "top-level array, interior duplicates",
			input:       `[1, 2, {"a": 1, "b": 2, "c": 3, "b": 4}, 4]`,
			expectedErr: `[2]: contains duplicate key "b"`,
		},
		{
			name:  "nested arrays, okay",
			input: `{ "a": [ [ [ "a", "b" ] ], 123 ], "b": false }`,
		},
		{
			name:        "nested arrays, interior duplicate",
			input:       `{ "a": [ [ [ "a", "b" ], {"foo": "bar", "foo": "baz"} ], 123 ], "b": false }`,
			expectedErr: `a[0][1]: contains duplicate key "foo"`,
		},
		{
			name:  "nested objects, okay",
			input: `{ "a": { "b": { "c": { "d": { }, "e": ["foo", "bar"] } } } }`,
		},
		{
			name:        "nested objects, interior duplicate",
			input:       `{ "a": { "b": { "c": { "d": { }, "d": "foobar" } } } }`,
			expectedErr: `a.b.c: contains duplicate key "d"`,
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			_, err := checkNoDuplicateKeys("", json.NewDecoder(strings.NewReader(testCase.input)))
			if testCase.expectedErr == "" {
				require.NoError(t, err, "expected input to succeed: %q", testCase.input)
			} else {
				require.ErrorContains(t, err, testCase.expectedErr, "expected input to fail with particular error: %q", testCase.input)
			}
		})
	}
}
