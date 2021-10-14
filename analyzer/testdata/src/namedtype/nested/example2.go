package nested

import (
	"extra"
	"namedtype/y/nested"
)

func example2() {
	sink = &Element{}
	sink = &nested.Element{}

	sink = extra.NewValue() // want `\Qextra Value`
}
