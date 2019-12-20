package namedtypes

import (
	"container/list"
	listcont "container/list"
)

var sink interface{}

type Element struct{}

func example() {
	sink = &list.Element{}     // want `list Element`
	sink = &listcont.Element{} // want `list Element`

	sink = &Element{} // want `Element`
}
