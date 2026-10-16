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
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"github.com/stretchr/testify/assert"
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
