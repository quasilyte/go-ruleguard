package nested

import (
	"extra"
)

func example2() {
	sink = &Element{} // want `Element`

	sink = extra.NewValue() // want `extra Value`
}
