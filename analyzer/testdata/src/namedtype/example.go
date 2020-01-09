package namedtype

import (
	"namedtype/y/nested"
)

var sink interface{}

func example() {
	sink = &nested.Element{}
}
