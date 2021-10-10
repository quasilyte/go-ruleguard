package stdlib

import (
	iorenamed "io"
)

func _(w iorenamed.Writer) {
	iorenamed.WriteString(w, "") // want `\QWriteString from stdlib`

	{
		var io foo
		io.WriteString(w, "")
	}
	{
		var iorenamed foo
		iorenamed.WriteString(w, "")
	}
}
