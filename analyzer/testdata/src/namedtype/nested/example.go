package nested

import (
	"container/list"
	listcont "container/list"
	htmltemplate "html/template"
	texttemplate "text/template"

	"extra"
)

var sink interface{}

type Element struct{}

func example() {
	sink = &list.Element{}     // want `list Element`
	sink = &listcont.Element{} // want `list Element`

	sink = &htmltemplate.Template{} // want `html Template`
	sink = &texttemplate.Template{} // want `text Template`

	sink = &Element{} // want `Element`

	sink = extra.NewValue() // want `extra Value`
}
