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

type tb struct {
	internal *testing.T
	failed   bool
}

func NewCrossTestT(t *testing.T) TB {
	t.Helper()
	return &tb{
		internal: t,
	}
}

func (t *tb) Helper() {
	t.internal.Helper()
}

func (t *tb) Errorf(format string, args ...any) {
	// t.Errorf was called at least once, so a failed test case
	// was found.
	t.failed = true
	t.internal.Errorf(format, args...)
}

func (t *tb) Fatalf(format string, args ...any) {
	t.internal.Fatalf(format, args...)
}

func (t *tb) Successf(format string, args ...any) {
	// Only log a success message if no instances of `t.Errorf` was
	// ever called.
	if !t.failed {
		t.internal.Logf(format, args...)
	}
}

func (t *tb) FailNow() {
	t.internal.FailNow()
}
