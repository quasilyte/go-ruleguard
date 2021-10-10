package stdlib

import (
	"fmt"
	io "myio"
)

func _(w io.Writer) {
	io.WriteString(w, "")

	sink(fmt.Sprint(1), fmt.Sprint("ok")) // want `\Qsink with two Sprint from stdlib`

	{
		var fmt foo
		sink(fmt.Sprint(1), fmt.Sprint("ok"))
	}
}
