// Copyright 2022 Buf Technologies, Inc.
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

package testing

import "testing"

type t struct {
	internal *testing.T
}

func NewCrossTestT(internal *testing.T) TB {
	return &t{
		internal: internal,
	}
}

func (t *t) Helper() {
	t.internal.Helper()
}

func (t *t) Errorf(format string, args ...any) {
	t.internal.Errorf(format, args...)
}

func (t *t) Fatalf(format string, args ...any) {
	t.internal.Fatalf(format, args...)
}

func (t *t) Successf(format string, args ...any) {
	t.internal.Logf(format, args...)
}
