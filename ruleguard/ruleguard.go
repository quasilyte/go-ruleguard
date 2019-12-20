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
	Report func(n ast.Node, msg string, s *Suggestion)
}

type Suggestion struct {
	From        token.Pos
	To          token.Pos
	Replacement []byte
}

func ParseRules(filename string, fset *token.FileSet, r io.Reader) (*GoRuleSet, error) {
	p := newRulesParser()
	return p.ParseFile(filename, fset, r)
}

func RunRules(ctx *Context, f *ast.File, rules *GoRuleSet) {
	rr := rulesRunner{ctx: ctx, rules: rules}
	rr.run(f)
}

type GoRuleSet struct {
	universal *scopedGoRuleSet
	local     *scopedGoRuleSet
}
