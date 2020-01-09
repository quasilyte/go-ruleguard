package nested

import (
	"container/list"
	listcont "container/list"
	htmltemplate "html/template"
	texttemplate "text/template"

	"extra"

	xnested "namedtype/x/nested"
	ynested "namedtype/y/nested"
)

var sink interface{}

type Element struct{}

func example() {
	sink = &list.Element{}     // want `list Element`
	sink = &listcont.Element{} // want `list Element`

	sink = &htmltemplate.Template{} // want `html Template`
	sink = &texttemplate.Template{} // want `text Template`

	sink = &xnested.Element{} // want `x/nested Element`
	sink = &ynested.Element{}

	sink = extra.NewValue() // want `extra Value`
}
