package gogrep

import (
	"math/bits"
	"testing"
	"unsafe"
)

func TestInstructionSize(t *testing.T) {
	if bits.UintSize != 64 {
		t.Skip("not 64-bit platform")
	}
	wantSize := 3
	haveSize := int(unsafe.Sizeof(instruction{}))
	if wantSize != haveSize {
		t.Errorf("sizeof(instruction): have %d, want %d", haveSize, wantSize)
	}
}
