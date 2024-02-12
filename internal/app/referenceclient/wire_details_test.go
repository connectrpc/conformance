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
			pr := &internal.SimplePrinter{}
			examineWireDetails(ctx, pr)
			if len(testCase.expectedFeedback) == 0 {
				assert.Equal(t, connect.CodeAlreadyExists, connect.CodeOf(err), "unexpected error: %v", err)
				assert.Empty(t, pr.Messages)
			} else {
				// When there's feedback, the connect-go client may complain about the end-stream message
				// and report a different code.
				assert.True(t, connect.CodeOf(err) == connect.CodeAlreadyExists ||
					connect.CodeOf(err) == connect.CodeInternal,
					"unexpected error: %v", err)
				for i := range pr.Messages {
					pr.Messages[i] = strings.TrimSuffix(pr.Messages[i], "\n")
				}
				assert.Empty(t, cmp.Diff(testCase.expectedFeedback, pr.Messages))
			}
		})
	}
}
