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

// Package console implements the internal/testing.TB interface.
package console

import (
	"log"
	"os"
)

type TB struct {
	failed bool
}

func NewTB() *TB {
	return &TB{}
}

func (t *TB) Helper() {}

func (t *TB) Errorf(format string, args ...any) {
	// t.Errorf was called at least once, so a failed test case
	// was found.
	t.failed = true
	log.Printf("ERROR: "+format, args...)
}

func (t *TB) Fatalf(format string, args ...any) {
	log.Printf("FAIL: "+format, args...)
}

func (t *TB) Successf(format string, args ...any) {
	if t.failed {
		t.FailNow()
	}
	log.Printf("SUCCESS: "+format, args...)
}

func (t *TB) FailNow() {
	os.Exit(1)
}
