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

package internal

import (
	"context"
	"fmt"

	v1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CancelTiming struct {
	BeforeCloseSend   *emptypb.Empty
	AfterCloseSendMs  int
	AfterNumResponses int
}

// GetCancelTiming evaluates a Cancel setting and returns a struct with the
// appropriate value set.
func GetCancelTiming(cancel *v1.ClientCompatRequest_Cancel) (*CancelTiming, error) {
	var beforeCloseSend *emptypb.Empty
	afterCloseSendMs := -1
	afterNumResponses := -1
	if cancel != nil {
		switch cancelTiming := cancel.CancelTiming.(type) {
		case *v1.ClientCompatRequest_Cancel_BeforeCloseSend:
			beforeCloseSend = cancelTiming.BeforeCloseSend
		case *v1.ClientCompatRequest_Cancel_AfterCloseSendMs:
			afterCloseSendMs = int(cancelTiming.AfterCloseSendMs)
		case *v1.ClientCompatRequest_Cancel_AfterNumResponses:
			afterNumResponses = int(cancelTiming.AfterNumResponses)
		case nil:
			// If cancel is non-nil, but none of timing values are set, it should
			// be treated as if afterCloseSendMs was set to 0
			afterCloseSendMs = 0
		default:
			return nil, fmt.Errorf("provided CancelTiming has an unexpected type %T", cancelTiming)
		}
	}
	return &CancelTiming{
		BeforeCloseSend:   beforeCloseSend,
		AfterCloseSendMs:  afterCloseSendMs,
		AfterNumResponses: afterNumResponses,
	}, nil
}

// WrapContext wraps the current context. The resulting context and cancel
// function will be dependent on whether the given context has a deadline.
func WrapContext(ctx context.Context) (context.Context, context.CancelFunc) {
	deadline, ok := ctx.Deadline()
	if !ok {
		return context.WithCancel(ctx)
	}
	return context.WithDeadline(ctx, deadline)
}
