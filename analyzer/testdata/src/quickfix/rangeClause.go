package quickfix

import (
	"bytes"
	"io"
)

func rangeRuneSlice(s string) {
	for _, ch := range []rune(s) { // want `\Qsuggestion: range s`
		println(ch)
	}

	{
		var ch rune
		for _, ch = range []rune(s[:]) { // want `\Qsuggestion: range s[:]`
			println(ch)
		}
	}

	{
		var ch rune
		for _, ch = range []rune(getString()) { // want `\Qsuggestion: range getString()`
			println(ch)
		}
	}

	{
		for _, ch1 := range []rune("foo") { // want `\Qsuggestion: range "foo"`
			for _, ch2 := range []rune("123") { // want `\Qsuggestion: range "123"`
				println(ch1, ch2)
			}
		}
	}
}

func writeString() {
	{
		var buf bytes.Buffer
		io.WriteString(&buf, "example") // want `\Qbuf.WriteString("example")`
	}
	{
		var buf bytes.Buffer
		io.WriteString((&buf), "example") // want `\Q(&buf).WriteString("example")`
		(&buf).WriteString("example")
	}
	{
		buf := &bytes.Buffer{}
		io.WriteString(buf, "example") // want `\Qbuf.WriteString("example")`
	}
	{
		var buffers [4]bytes.Buffer
		io.WriteString(&buffers[0], "str") // want `\Qbuffers[0].WriteString("str")`
		buffers[0].WriteString("str")
	}
	{
		type withBuffer struct {
			buf bytes.Buffer
		}
		var o withBuffer
		io.WriteString(&o.buf, "foo") // want `\Qo.buf.WriteString("foo")`
		o.buf.WriteString("foo")
	}
	{
		type withBufferPtr struct {
			buf *bytes.Buffer
		}
		var o withBufferPtr
		io.WriteString(o.buf, "foo") // want `\Qo.buf.WriteString("foo")`
		o.buf.WriteString("foo")
	}
}

func getString() string {
	return "123"
}
