package ruleguard

import (
	"go/ast"
	"go/token"
	"go/types"
	"io"
)

type Context struct {
	Types  *types.Info
	Fset   *token.FileSet
	Report func(n ast.Node, msg string)
}

func ParseRules(filename string, fset *token.FileSet, r io.Reader) (*GoRuleSet, error) {
	var p rulesParser
	if err := p.init(filename, fset, r); err != nil {
		return nil, err
	}
	err := p.parseTop()
	return p.res, err
}

func RunRules(ctx *Context, f *ast.File, rules *GoRuleSet) {
	rr := rulesRunner{ctx: ctx, rules: rules}
	rr.run(f)
}

type GoRuleSet struct {
	universal *scopedGoRuleSet
	local     *scopedGoRuleSet
}
