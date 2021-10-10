package stdlib

import "io"

type foo struct{}

func (foo) WriteString(args ...interface{}) {}

func (foo) Sprint(args ...interface{}) string { return "" }

func sink(args ...interface{}) {}

func test(w io.Writer) {
	io.WriteString(w, "") // want `\QWriteString from stdlib`

	{
		var io foo
		io.WriteString(w, "")
	}
}
