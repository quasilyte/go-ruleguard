package ruleguard

import (
	"go/ast"
	"go/types"

	"github.com/quasilyte/go-ruleguard/internal/mvdan.cc/gogrep"
)

type scopedGoRuleSet struct {
	uncategorized   []goRule
	categorizedNum  int
	rulesByCategory [nodeCategoriesCount][]goRule
}

type goRule struct {
	group      string
	filename   string
	line       int
	severity   string
	pat        *gogrep.Pattern
	msg        string
	location   string
	suggestion string
	filter     matchFilter
}

type matchFilter struct {
	fileFilters []fileFilter
	subFilters  map[string][]nodeFilter
}

type fileFilter struct {
	src  string
	pred func(*fileFilterParams) bool
}

type fileFilterParams struct {
	ctx      *Context
	filename string
	imports  map[string]struct{}
}

type nodeFilter struct {
	src  string
	pred func(*nodeFilterParams) bool
}

type nodeFilterParams struct {
	ctx *Context
	n   ast.Expr

	nodeText func(n ast.Node) []byte
}

func (params *nodeFilterParams) nodeType() types.Type {
	return params.typeofNode(params.n)
}

func (params *nodeFilterParams) typeofNode(n ast.Node) types.Type {
	if e, ok := n.(ast.Expr); ok {
		if typ := params.ctx.Types.TypeOf(e); typ != nil {
			return typ
		}
	}

	return types.Typ[types.Invalid]
}
