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
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
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
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/proto"
	_ "google.golang.org/protobuf/types/descriptorpb" // needed in global registry for test case
	"google.golang.org/protobuf/types/known/anypb"
)

func TestExamineConnectError(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name             string
		compressed       bool
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
			name:       "compressed",
			compressed: true,
			endStream: `
				{
					"code": "internal",
					"message": "blah blah blah"
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
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			svr := httptest.NewServer(http.HandlerFunc(func(respWriter http.ResponseWriter, _ *http.Request) {
				respWriter.Header().Set("Content-Type", "application/json")
				if testCase.compressed {
					respWriter.Header().Set("Content-Encoding", "gzip")
				}
				respWriter.WriteHeader(http.StatusBadRequest)
				if testCase.compressed {
					w := gzip.NewWriter(respWriter)
					_, _ = w.Write([]byte(testCase.endStream))
					_ = w.Close()
				} else {
					_, _ = respWriter.Write([]byte(testCase.endStream))
				}
			}))
			t.Cleanup(svr.Close)
			client := conformancev1connect.NewConformanceServiceClient(
				&http.Client{Transport: newWireCaptureTransport(svr.Client().Transport, nil)},
				svr.URL,
			)
			ctx := withWireCapture(context.Background())
			req := connect.NewRequest(&conformancev1.UnaryRequest{})
			req.Header().Set("X-Test-Case-Name", "foo") // needed to enable tracing
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
		compressed       bool
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
			name:       "compressed",
			compressed: true,
			endStream: `
				{
					"error": {"code": "invalid_argument", "message": "foobar"},
					"metadata":{
						"foo": ["bar", "baz"]
					}
				}`,
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
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			svr := httptest.NewServer(http.HandlerFunc(func(respWriter http.ResponseWriter, _ *http.Request) {
				respWriter.Header().Set("Content-Type", "application/connect+proto")
				if testCase.compressed {
					respWriter.Header().Set("Connect-Content-Encoding", "gzip")
					_, _ = respWriter.Write([]byte{3}) // end-stream + compressed flags
				} else {
					_, _ = respWriter.Write([]byte{2}) // just the end-stream flag
				}
				writeStreamFrame([]byte(testCase.endStream), testCase.compressed, respWriter)
			}))
			t.Cleanup(svr.Close)
			client := conformancev1connect.NewConformanceServiceClient(
				&http.Client{Transport: newWireCaptureTransport(svr.Client().Transport, nil)},
				svr.URL,
			)
			ctx := withWireCapture(context.Background())
			stream := client.ClientStream(ctx)
			stream.RequestHeader().Set("X-Test-Case-Name", "foo") // needed to enable tracing
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
	stat := &status.Status{
		Code:    3,
		Message: "foo",
	}
	data, err := proto.Marshal(stat)
	require.NoError(t, err)
	statusCode3MessageFooNoDetails := base64.RawStdEncoding.EncodeToString(data)
	statusCode3MessageFooNoDetailsPadded := base64.StdEncoding.EncodeToString(data)
	stat = &status.Status{
		Code:    3,
		Message: "foo",
		Details: []*anypb.Any{
			{TypeUrl: "type.googleapis.com/google.rpc.Status", Value: data},
		},
	}
	data, err = proto.Marshal(stat)
	require.NoError(t, err)
	statusCode3MessageFooDetails := base64.RawStdEncoding.EncodeToString(data)
	stat = &status.Status{
		Code: 0,
		Details: []*anypb.Any{
			{TypeUrl: "type.googleapis.com/google.rpc.Status", Value: data},
		},
	}
	data, err = proto.Marshal(stat)
	require.NoError(t, err)
	statusCode0EmptyMessageDetails := base64.RawStdEncoding.EncodeToString(data)
	stat = &status.Status{Code: 0}
	data, err = proto.Marshal(stat)
	require.NoError(t, err)
	statusCode0EmptyMessageNoDetails := base64.RawStdEncoding.EncodeToString(data)

	testCases := []struct {
		name             string
		compressed       bool
		endStream        string
		expectedCode     connect.Code
		expectedFeedback []string
	}{
		{
			name: "correct",
			endStream: "grpc-status: 6\r\n" +
				"grpc-message: foo\r\n" +
				"blah-blah: foobar\r\n",
			expectedCode: 6,
		},
		{
			name:       "compressed",
			compressed: true,
			endStream: "grpc-status: 6\r\n" +
				"grpc-message: foo\r\n" +
				"blah-blah: foobar\r\n",
			expectedCode: 6,
		},
		{
			name: "allowed special chars in key",
			endStream: "grpc-status: 6\r\n" +
				"grpc-message: foo\r\n" +
				"`~blah.blah-#blah|blah%blah's~`: foobar\r\n",
			expectedCode: 6,
		},
		{
			name: "allowed special chars in value",
			endStream: "grpc-status: 6\r\n" +
				"grpc-message: foo\r\n" +
				"blah-blah: {foobar} \"baz\" 🤷🦸\r\n",
			expectedCode: 6,
		},
		{
			name: "mixed case",
			endStream: "Grpc-Status: 6\r\n" +
				"Grpc-Message: foo\r\n" +
				"Blah-Blah: foobar\r\n",
			expectedCode: 6,
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
			expectedCode: 6,
			expectedFeedback: []string{
				"grpc-web trailers should end with CRLF but does not",
			},
		},
		{
			name: "key without value",
			endStream: "grpc-status: 6\r\n" +
				"grpc-message: foo\r\n" +
				"blah-blah\r\n",
			expectedCode: 6,
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
			expectedCode: 6,
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
			expectedCode: 6,
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
			expectedCode: 6,
			expectedFeedback: []string{
				"grpc-web trailers include blank lines",
			},
		},
		{
			name: "obsolete line folding",
			endStream: "grpc-status: 6\r\n" +
				"grpc-message: foobar\r\n" +
				"blah-blah: foo\r\n" +
				" \tbar\r\n" +
				" \tbaz\r\n",
			expectedCode: 6,
			expectedFeedback: []string{
				"grpc-web trailers use obsolete line-folding",
			},
		},
		{
			name: "invalid char in value",
			endStream: "grpc-status: 6\r\n" +
				"grpc-message: foo\r\n" +
				"blah-blah: foo\x00bar\r\n",
			expectedCode: 6,
			expectedFeedback: []string{
				`grpc-web trailers include invalid field; value contains invalid characters: "blah-blah: foo\x00bar"`,
			},
		},
		{
			name: "invalid char in key",
			endStream: "grpc-status: 6\r\n" +
				"grpc-message: foo\r\n" +
				"blah[blah]blah: foo bar\r\n",
			expectedCode: 6,
			expectedFeedback: []string{
				`grpc-web trailers include invalid field; name contains invalid characters: "blah[blah]blah: foo bar"`,
			},
		},
		{
			name: "missing grpc-status",
			endStream: "grpc-message: foo\r\n" +
				"blah-blah: foo bar\r\n",
			expectedCode: connect.CodeUnknown,
			expectedFeedback: []string{
				`trailers did not include 'grpc-status' key`,
			},
		},
		{
			name: "multiple grpc-status",
			endStream: "grpc-status: 1\r\n" +
				"grpc-status: 1\r\n" +
				"grpc-message: foo\r\n" +
				"blah-blah: foo bar\r\n",
			expectedCode: 1,
			expectedFeedback: []string{
				`trailers include multiple 'grpc-status' keys (2)`,
			},
		},
		{
			name: "invalid grpc-status",
			endStream: "grpc-status: abc\r\n" +
				"grpc-message: foo\r\n" +
				"blah-blah: foo bar\r\n",
			expectedCode: connect.CodeUnknown,
			expectedFeedback: []string{
				`trailers include invalid 'grpc-status' value "abc": strconv.Atoi: parsing "abc": invalid syntax`,
			},
		},
		{
			name: "negative grpc-status",
			endStream: "grpc-status: -1\r\n" +
				"grpc-message: foo\r\n" +
				"blah-blah: foo bar\r\n",
			expectedCode: connect.CodeUnknown,
			expectedFeedback: []string{
				`trailers include invalid 'grpc-status' value -1: should be >= 0 && <= 16`,
			},
		},
		{
			name: "out-of-range grpc-status",
			endStream: "grpc-status: 17\r\n" +
				"grpc-message: foo\r\n" +
				"blah-blah: foo bar\r\n",
			expectedCode: 17,
			expectedFeedback: []string{
				`trailers include invalid 'grpc-status' value 17: should be >= 0 && <= 16`,
			},
		},
		{
			name: "multiple grpc-message",
			endStream: "grpc-status: 1\r\n" +
				"grpc-message: foo\r\n" +
				"grpc-message: foo\r\n" +
				"grpc-message: foo\r\n" +
				"blah-blah: foo bar\r\n",
			expectedCode: 1,
			expectedFeedback: []string{
				`trailers include multiple 'grpc-message' keys (3)`,
			},
		},
		{
			name: "invalid grpc-message - not percent encoded",
			endStream: "grpc-status: 3\r\n" +
				"grpc-message: abc\tdef\r\n" +
				"blah-blah: foo bar\r\n",
			expectedCode: 3,
			expectedFeedback: []string{
				`trailers include incorrectly-encoded 'grpc-message' value "abc\tdef": byte at position 3 (0x09) should be percent-encoded`,
			},
		},
		{
			name: "invalid grpc-message - not percent encoded 2",
			endStream: "grpc-status: 3\r\n" +
				"grpc-message: abc def 🧐\r\n" +
				"blah-blah: foo bar\r\n",
			expectedCode: 3,
			expectedFeedback: []string{
				`trailers include incorrectly-encoded 'grpc-message' value "abc def 🧐": byte at position 8 (0xf0) should be percent-encoded`,
			},
		},
		{
			name: "invalid grpc-message - illegal percent encoding",
			endStream: "grpc-status: 3\r\n" +
				"grpc-message: abc def %x12\r\n" +
				"blah-blah: foo bar\r\n",
			expectedCode: 3,
			expectedFeedback: []string{
				`trailers include incorrectly-encoded 'grpc-message' value "abc def %x12": byte at position 9 (0x78) should be hexadecimal digit`,
			},
		},
		{
			name: "invalid grpc-message - incomplete percent encoding",
			endStream: "grpc-status: 3\r\n" +
				"grpc-message: %ab %de %f0 %1\r\n" +
				"blah-blah: foo bar\r\n",
			expectedCode: 3,
			expectedFeedback: []string{
				`trailers include incorrectly-encoded 'grpc-message' value "%ab %de %f0 %1": incomplete percent-encoded character at the end`,
			},
		},
		{
			name: "unnecessary grpc-message okay if blank",
			endStream: "grpc-status: 0\r\n" +
				"grpc-message: \r\n" +
				"blah-blah: foo bar\r\n",
		},
		{
			name: "unnecessary grpc-message",
			endStream: "grpc-status: 0\r\n" +
				"grpc-message: foo\r\n" +
				"blah-blah: foo bar\r\n",
			expectedFeedback: []string{
				`trailers include a non-empty 'grpc-message' value with zero/okay 'grpc-status'`,
			},
		},
		{
			name: "with grpc-status-details-bin",
			endStream: "grpc-status: 3\r\n" +
				"grpc-message: foo\r\n" +
				"grpc-status-details-bin: " + statusCode3MessageFooDetails + "\r\n",
			expectedCode: 3,
		},
		{
			name: "invalid grpc-status-details-bin - not base64",
			endStream: "grpc-status: 12\r\n" +
				"grpc-message: foo\r\n" +
				"grpc-status-details-bin: foo bar\r\n",
			expectedCode: 12,
			expectedFeedback: []string{
				`trailers include incorrectly-encoded 'grpc-status-details-bin' value: illegal base64 data at input byte 3`,
			},
		},
		{
			name: "invalid grpc-status-details-bin - padded",
			endStream: "grpc-status: 3\r\n" +
				"grpc-message: foo\r\n" +
				"grpc-status-details-bin: " + statusCode3MessageFooNoDetailsPadded + "\r\n",
			expectedCode: 3,
			expectedFeedback: []string{
				`trailers include 'grpc-status-details-bin' value with padding but servers should emit unpadded: ` + statusCode3MessageFooNoDetailsPadded,
			},
		},
		{
			name: "invalid grpc-status-details-bin - not proto",
			endStream: "grpc-status: 3\r\n" +
				"grpc-message: foo\r\n" +
				"grpc-status-details-bin: AbCdEfG\r\n",
			expectedCode: 3,
			expectedFeedback: []string{
				`trailers include un-parseable 'grpc-status-details-bin' value: proto: cannot parse invalid wire-format data`,
			},
		},
		{
			name: "grpc-status-details-bin disagrees with grpc-status",
			endStream: "grpc-status: 12\r\n" +
				"grpc-message: foo\r\n" +
				"grpc-status-details-bin: " + statusCode3MessageFooNoDetails + "\r\n",
			expectedCode: 3,
			expectedFeedback: []string{
				`trailers include 'grpc-status-details-bin' value that disagrees with 'grpc-status' value: 3 != 12`,
			},
		},
		{
			name: "grpc-status-details-bin disagrees with grpc-message",
			endStream: "grpc-status: 3\r\n" +
				"grpc-message: bar\r\n" +
				"grpc-status-details-bin: " + statusCode3MessageFooDetails + "\r\n",
			expectedCode: 3,
			expectedFeedback: []string{
				`trailers include 'grpc-status-details-bin' value that disagrees with 'grpc-message' value: "foo" != "bar"`,
			},
		},
		{
			name: "unnecessary grpc-status-details-bin",
			endStream: "grpc-status: 0\r\n" +
				"grpc-status-details-bin: " + statusCode0EmptyMessageDetails + "\r\n",
			expectedFeedback: []string{
				`trailers include 'grpc-status-details-bin' value with zero/okay 'grpc-status' and non-empty details`,
			},
		},
		{
			name: "unnecessary grpc-status-details-bin okay if details empty",
			endStream: "grpc-status: 0\r\n" +
				"grpc-status-details-bin: " + statusCode0EmptyMessageNoDetails + "\r\n",
		},
		{
			name: "invalid binary metadata - not base64",
			endStream: "grpc-status: 0\r\n" +
				"foo-bar-bin: foo bar\r\n",
			expectedFeedback: []string{
				`trailers include incorrectly-encoded 'Foo-Bar-Bin' value: illegal base64 data at input byte 3`,
			},
		},
		{
			name: "invalid binary metadata - padded",
			endStream: "grpc-status: 3\r\n" +
				"grpc-message: foo\r\n" +
				"foo-bar-bin: AbCcDdEeFfGg01==\r\n",
			expectedCode: 3,
			expectedFeedback: []string{
				`metadata include 'Foo-Bar-Bin' value with padding but servers should emit unpadded: AbCcDdEeFfGg01==`,
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			svr := httptest.NewServer(http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
				respWriter.Header().Set("Content-Type", "application/grpc-web")
				if req.Header.Get("Expect-Code") == "" {
					// Not expecting an error, so send a message.
					_, _ = respWriter.Write([]byte{0}) // flags indicate simple message
					writeStreamFrame(nil, false, respWriter)
				}
				if testCase.compressed {
					respWriter.Header().Set("Grpc-Encoding", "gzip")
					_, _ = respWriter.Write([]byte{129}) // end-stream + compressed flags
				} else {
					_, _ = respWriter.Write([]byte{128}) // just the end-stream flag
				}
				writeStreamFrame([]byte(testCase.endStream), testCase.compressed, respWriter)
			}))
			t.Cleanup(svr.Close)
			client := conformancev1connect.NewConformanceServiceClient(
				&http.Client{Transport: newWireCaptureTransport(svr.Client().Transport, nil)},
				svr.URL,
				connect.WithGRPCWeb(),
			)
			ctx := withWireCapture(context.Background())
			req := connect.NewRequest(&conformancev1.UnaryRequest{})
			req.Header().Set("X-Test-Case-Name", "foo") // needed to enable tracing
			if testCase.expectedCode != 0 {
				req.Header().Set("Expect-Code", testCase.expectedCode.String())
			}
			resp, err := client.Unary(ctx, req)

			// Now we can check the details and report issues:
			printer := &internal.SimplePrinter{}
			examineWireDetails(ctx, printer)
			// Check binary headers/trailers, too
			if err != nil {
				var connErr *connect.Error
				if errors.As(err, &connErr) {
					checkBinaryMetadata("metadata", internal.ConvertToProtoHeader(connErr.Meta()), printer)
				}
			} else {
				checkBinaryMetadata("headers", internal.ConvertToProtoHeader(resp.Header()), printer)
				checkBinaryMetadata("trailers", internal.ConvertToProtoHeader(resp.Trailer()), printer)
			}

			if len(testCase.expectedFeedback) == 0 {
				if testCase.expectedCode == 0 {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
					assert.Equal(t, testCase.expectedCode, connect.CodeOf(err), "unexpected error: %v", err)
				}
				assert.Empty(t, printer.Messages)
				return
			}

			// When there's feedback, the connect-go client may complain about the end-stream message
			// and report a different code.
			options := []connect.Code{connect.CodeInternal, connect.CodeUnknown}
			if testCase.expectedCode != 0 {
				require.Error(t, err)
				options = append(options, testCase.expectedCode)
			}
			if err != nil {
				assert.Contains(t,
					options,
					connect.CodeOf(err),
					"unexpected error: %v", err,
				)
			}
			for i := range printer.Messages {
				printer.Messages[i] = strings.TrimSuffix(printer.Messages[i], "\n")
			}
			// Ugh, proto library doesn't want its error messages used in tests so makes the
			// output non-deterministic, sometimes using a non-breaking space instead of space :/
			for i, msg := range printer.Messages {
				printer.Messages[i] = strings.ReplaceAll(msg, "\u00a0", " ")
			}
			assert.Empty(t, cmp.Diff(testCase.expectedFeedback, printer.Messages))
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

func writeStreamFrame(data []byte, compressed bool, writer io.Writer) {
	if compressed {
		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)
		_, _ = w.Write(data)
		_ = w.Close()
		data = buf.Bytes()
	}
	var size [4]byte
	binary.BigEndian.PutUint32(size[:], uint32(len(data)))
	_, _ = writer.Write(size[:])
	_, _ = writer.Write(data)
}
