// Copyright 2022-2023 The Connect Authors
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

package conformancetesting

import "testing"

// TB is a testing interface that the conformance suite depends on. It is trimmed down
// from the standard library testing.TB interface and adds a Successf method.
type TB interface {
	Helper()
	Errorf(string, ...any)
	Fatalf(string, ...any)
	Successf(string, ...any)
	FailNow()
}

// NewTB returns a new TB.
func NewTB(t *testing.T) TB {
	t.Helper()
	return &tb{
		t: t,
	}
}
