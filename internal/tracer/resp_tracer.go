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
	"context"
	"net/http"

	"sync/atomic"
)

type respKey struct{}

type RespWrapper struct {
	val atomic.Pointer[http.Response]
}

func WithResponseCapture(ctx context.Context) (context.Context, *RespWrapper) {
	wrappers := &RespWrapper{}
	ctx = context.WithValue(ctx, respKey{}, wrappers)
	return ctx, wrappers
}

// Get returns the resp captured. Resps are not captured until the response body is
// exhausted.
func (t *RespWrapper) Get() *http.Response {
	respPtr := t.val.Load()
	if respPtr == nil {
		return nil
	}
	return respPtr
}

type contextTracer struct {
	tracer *Tracer
}

func NewContextTracer(trace *Tracer) *contextTracer {
	return &contextTracer{
		tracer: trace,
	}
}

func (t *contextTracer) Complete(trace Trace) {
	if t != nil {
		t.tracer.Complete(trace)
	}
}
