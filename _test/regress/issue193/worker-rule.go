// +build ignore

package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
	"github.com/quasilyte/go-ruleguard/dsl/types"
)

func implementsWorker(ctx *dsl.VarFilterContext) bool {
	worker := ctx.GetInterface(`github.com/go-ruleguard/rg1.Worker`)
	return types.Implements(ctx.Type, worker) ||
		types.Implements(types.NewPointer(ctx.Type), worker)
}

func workerLiteral(m dsl.Matcher) {
	m.Match(`$x{$*_}`).
		Where(m["x"].Filter(implementsWorker)).
		Report("$x implements worker")
}
