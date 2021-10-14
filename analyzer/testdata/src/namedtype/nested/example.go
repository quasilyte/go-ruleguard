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
	sink = &list.Element{}     // want `\Qlist Element`
	sink = &listcont.Element{} // want `\Qlist Element`

	sink = &htmltemplate.Template{} // want `\Qhtml Template`
	sink = &texttemplate.Template{} // want `\Qtext Template`

	sink = &xnested.Element{} // want `\Qx/nested Element`
	sink = &ynested.Element{}

	sink = extra.NewValue() // want `\Qextra Value`
}
