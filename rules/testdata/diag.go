package target

import (
	"bytes"
	"strings"
)

func badCond() {
	var s string
	var b []byte

	_ = strings.Count(s, "/") >= 0       // want `\QbadCond: statement always true`
	_ = bytes.Count(b, []byte("/")) >= 0 // want `\QbadCond: statement always true`
}
