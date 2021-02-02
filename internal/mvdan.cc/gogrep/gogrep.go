package gogrep

import (
	"go/ast"
	"go/token"
)

type StmtList = stmtList
type ExprList = exprList

// Parse creates a gogrep pattern out of a given string expression.
func Parse(fset *token.FileSet, expr string) (*Pattern, error) {
	m := matcher{fset: fset}
	node, err := m.parseExpr(expr)
	if err != nil {
		return nil, err
	}
	return &Pattern{m: &m, Expr: node}, nil
}

// Pattern is a compiled gogrep pattern.
type Pattern struct {
	Expr ast.Node
	m    *matcher
}

// MatchData describes a successful pattern match.
type MatchData struct {
	Node   ast.Node
	Values map[string]ast.Node
}

// Clone creates a pattern copy.
func (p *Pattern) Clone() *Pattern {
	clone := *p
	clone.m = &matcher{}
	*clone.m = *p.m
	clone.m.values = make(map[string]ast.Node)
	return &clone
}

// MatchNode calls cb if n matches a pattern.
func (p *Pattern) MatchNode(n ast.Node, cb func(MatchData)) {
	p.m.values = map[string]ast.Node{}
	if p.m.node(p.Expr, n) {
		cb(MatchData{
			Values: p.m.values,
			Node:   n,
		})
	}
}

func (p *Pattern) MatchStmtList(stmts []ast.Stmt, cb func(MatchData)) {
	p.matchNodeList(p.Expr.(stmtList), stmtList(stmts), cb)
}

func (p *Pattern) MatchExprList(exprs []ast.Expr, cb func(MatchData)) {
	p.matchNodeList(p.Expr.(exprList), exprList(exprs), cb)
}

func (p *Pattern) matchNodeList(pattern, list nodeList, cb func(MatchData)) {
	listLen := list.len()
	from := 0
	for {
		p.m.values = map[string]ast.Node{}
		matched, offset := p.m.nodes(pattern, list.slice(from, listLen), true)
		if matched == nil {
			break
		}
		cb(MatchData{
			Values: p.m.values,
			Node:   matched,
		})
		from += offset - 1
		if from >= listLen {
			break
		}
	}
}
