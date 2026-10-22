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
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/connect/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestCheckDeadlineError(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	t.Cleanup(cancel)

	// A canceled error sent by the server has no cause and must keep its code.
	serverErr := connect.NewError(connect.CodeCanceled, "canceled by server")
	assert.Same(t, serverErr, checkDeadlineError(ctx, serverErr))

	otherErr := connect.NewError(connect.CodeCanceled, "canceled").
		WithCause(errors.New("stream error: stream ID 3; INTERNAL_ERROR; received from peer"))
	assert.Same(t, otherErr, checkDeadlineError(ctx, otherErr))

	rstErr := connect.NewError(connect.CodeCanceled, "canceled").
		WithCause(errors.New("stream error: stream ID 3; CANCEL; received from peer"))
	assert.Equal(t, connect.CodeDeadlineExceeded, connect.CodeOf(checkDeadlineError(ctx, rstErr)))
}

func TestBidiStreamReturnsErrorForInvalidRequestMessage(t *testing.T) {
	t.Parallel()
	svr := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(svr.Close)
	svrURL, err := url.Parse(svr.URL)
	require.NoError(t, err)
	invoker := newInvoker(svr.Client().Transport, false, svrURL, nil)
	wrongType, err := anypb.New(&conformancev1.UnaryRequest{})
	require.NoError(t, err)
	for _, streamType := range []conformancev1.StreamType{
		conformancev1.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM,
		conformancev1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
	} {
		t.Run(streamType.String(), func(t *testing.T) {
			t.Parallel()
			result, err := invoker.bidiStream(t.Context(), &conformancev1.ClientCompatRequest{
				StreamType:      streamType,
				RequestMessages: []*anypb.Any{wrongType},
			})
			require.Error(t, err)
			assert.Nil(t, result)
		})
	}
}
