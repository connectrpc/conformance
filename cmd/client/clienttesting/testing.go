// Package clienttesting implements the internal/testing.TB interface.
package clienttesting

import (
	"log"

	"github.com/bufbuild/connect-crosstest/internal/testing"
)

type t struct {
}

func NewClientTestingT() testing.TB {
	return &t{}
}

func (t *t) Helper() {
}

func (t *t) Errorf(format string, args ...any) {
	log.Printf("ERROR: "+format, args...)
}

func (t *t) Fatalf(format string, args ...any) {
	log.Fatalf("FAIL: "+format, args...)
}

func (t *t) Successf(format string, args ...any) {
	log.Printf("SUCCESS: "+format, args...)
}
