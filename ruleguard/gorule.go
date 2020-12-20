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

type matchFilterResult string

func (s matchFilterResult) Matched() bool { return s == "" }

func (s matchFilterResult) RejectReason() string { return string(s) }

type filterFunc func(*filterParams) matchFilterResult

type matchFilter struct {
	src string
	fn  func(*filterParams) matchFilterResult
}

type filterParams struct {
	ctx      *Context
	filename string
	imports  map[string]struct{}

	values map[string]ast.Node

	nodeText func(n ast.Node) []byte
}

func (params *filterParams) subExpr(name string) ast.Expr {
	switch n := params.values[name].(type) {
	case ast.Expr:
		return n
	case *ast.ExprStmt:
		return n.X
	default:
		return nil
	}
}

func (params *filterParams) typeofNode(n ast.Node) types.Type {
	if e, ok := n.(ast.Expr); ok {
		if typ := params.ctx.Types.TypeOf(e); typ != nil {
			return typ
		}
	}

	return types.Typ[types.Invalid]
}
