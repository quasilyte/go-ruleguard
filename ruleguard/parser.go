package ruleguard

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io"

	"github.com/quasilyte/go-ruleguard/internal/mvdan.cc/gogrep"
)

type rulesParser struct {
	fset *token.FileSet
	res  *GoRuleSet
}

func newRulesParser() *rulesParser {
	return &rulesParser{}
}

func (p *rulesParser) ParseFile(filename string, fset *token.FileSet, r io.Reader) (*GoRuleSet, error) {
	p.fset = fset
	p.res = &GoRuleSet{
		local:     &scopedGoRuleSet{},
		universal: &scopedGoRuleSet{},
	}

	parserFlags := parser.Mode(0)
	f, err := parser.ParseFile(fset, filename, r, parserFlags)
	if err != nil {
		return nil, err
	}

	for _, decl := range f.Decls {
		if err := p.parseDecl(decl); err != nil {
			return nil, err
		}
	}

	return p.res, nil
}

func (p *rulesParser) parseDecl(d ast.Decl) error {
	switch d := d.(type) {
	case *ast.FuncDecl:
		return p.parseFuncDecl(d)
	}
	return nil
}

func (p *rulesParser) parseFuncDecl(f *ast.FuncDecl) error {
	if f.Body == nil {
		return p.errorf(f, "unexpected empty function body")
	}

	list := f.Body.List
	var err error
	for len(list) > 0 {
		list, err = p.parseRule(list)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *rulesParser) parseRule(list []ast.Stmt) ([]ast.Stmt, error) {
	dst := p.res.universal
	filters := map[string]submatchFilter{}
	proto := goRule{filters: filters}
	var alternatives []string

	match := p.toCallExpr(list[0])
	if match == nil || p.funcName(match) != "Match" {
		return list, p.errorf(list[0], "expected a Match call, found %s", sprintNode(p.fset, list[0]))
	}
	if len(match.Args) == 0 {
		return list, p.errorf(match.Fun, "Match expects at least 1 argument")
	}
	for _, arg := range match.Args {
		alt, ok := p.toStringValue(arg)
		if !ok {
			return list, p.errorf(arg, "expected a string literal argument")
		}
		alternatives = append(alternatives, alt)
	}
	list = list[1:]

	if len(list) == 0 {
		return list, p.errorf(match, "expected filter or yield clause, found nothing")
	}
	filter := p.toCallExpr(list[0])
	if filter == nil {
		return list, p.errorf(list[0], "expected filter or yield clause, found %s", sprintNode(p.fset, list[0]))
	}
	if p.funcName(filter) == "Filter" {
		if len(filter.Args) != 1 {
			return list, p.errorf(filter.Fun, "Filter expects exactly 1 argument, %d given", len(filter.Args))
		}
		if err := p.walkFilter(filters, filter.Args[0], false); err != nil {
			return list, err
		}
		list = list[1:]
	}

	if len(list) == 0 {
		lastNode := filter
		if lastNode == nil {
			lastNode = match
		}
		return list, p.errorf(lastNode, "expected yield clause, found nothing")
	}
	yield := p.toCallExpr(list[0])
	yieldFunc := p.funcName(yield)
	switch yieldFunc {
	case "Error":
		proto.severity = "error"
	case "Warn":
		proto.severity = "warn"
	case "Info":
		proto.severity = "info"
	case "Hint":
		proto.severity = "hint"
	default:
		return list, p.errorf(list[0], "expected a Error/Warn/Info/Hint call, found %s", sprintNode(p.fset, list[0]))
	}
	if len(yield.Args) != 1 {
		return list, p.errorf(yield.Fun, "%s expects exactly 1 argument, %d given", yieldFunc, len(yield.Args))
	}
	message, ok := p.toStringValue(yield.Args[0])
	if !ok {
		return list, p.errorf(yield.Args[0], "expected string literal argument")
	}
	proto.msg = message
	list = list[1:]

	for i, alt := range alternatives {
		rule := proto
		pat, err := gogrep.Parse(p.fset, alt)
		if err != nil {
			return list, p.errorf(match.Args[i], "gogrep parse: %v", err)
		}
		rule.pat = pat
		cat := categorizeNode(pat.Expr)
		if cat == nodeUnknown {
			dst.uncategorized = append(dst.uncategorized, rule)
		} else {
			dst.categorizedNum++
			dst.rulesByCategory[cat] = append(dst.rulesByCategory[cat], rule)
		}
	}

	return list, nil
}

func (p *rulesParser) walkFilter(dst map[string]submatchFilter, e ast.Expr, negate bool) error {
	switch e := e.(type) {
	case *ast.UnaryExpr:
		if e.Op == token.NOT {
			return p.walkFilter(dst, e.X, !negate)
		}
	case *ast.BinaryExpr:
		if e.Op == token.LAND {
			err := p.walkFilter(dst, e.X, negate)
			if err != nil {
				return err
			}
			return p.walkFilter(dst, e.Y, negate)
		}
	}

	// TODO(quasilyte): refactor and extend.
	operand := p.toFilterOperand(e)
	args := operand.args
	filter := dst[operand.varName]
	switch operand.path {
	default:
		return p.errorf(e, "%s is not a valid filter expression", sprintNode(p.fset, e))
	case "Pure":
		if negate {
			filter.pure = bool3false
		} else {
			filter.pure = bool3true
		}
		dst[operand.varName] = filter
	case "Const":
		if negate {
			filter.constant = bool3false
		} else {
			filter.constant = bool3true
		}
		dst[operand.varName] = filter
	case "Type.Is":
		if len(args) != 1 {
			return p.errorf(e, "Type.Is() expects exactly 1 argument, %d given", len(args))
		}
		typeString, ok := p.toStringValue(args[0])
		if !ok {
			return p.errorf(args[0], "expected a string literal argument")
		}
		y, err := typeFromString(typeString)
		if err != nil {
			return p.errorf(args[0], "parse type expr: %v", err)
		}
		if y == nil {
			return p.errorf(args[0], "can't convert %s into a type constraint yet", typeString)
		}
		wantIdentical := !negate
		filter.typePred = func(x types.Type) bool {
			return wantIdentical == types.Identical(x, y)
		}
		dst[operand.varName] = filter
	case "Type.AssignableTo":
		if len(args) != 1 {
			return p.errorf(e, "Type.Implements() expects exactly 1 argument, %d given", len(args))
		}
		typeString, ok := p.toStringValue(args[0])
		if !ok {
			return p.errorf(args[0], "expected a string literal argument")
		}
		y, err := typeFromString(typeString)
		if err != nil {
			return p.errorf(args[0], "parse type expr: %v", err)
		}
		if y == nil {
			return p.errorf(args[0], "can't convert %s into a type constraint yet", typeString)
		}
		wantAssignable := !negate
		filter.typePred = func(x types.Type) bool {
			return wantAssignable == types.AssignableTo(x, y)
		}
		dst[operand.varName] = filter
	}

	return nil
}

func (p *rulesParser) funcName(call *ast.CallExpr) string {
	if call == nil {
		return ""
	}
	x := call.Fun
	switch x := x.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		left, ok := x.X.(*ast.Ident)
		right := x.Sel
		if ok {
			return left.Name + "." + right.Name
		}
	}
	return ""
}

func (p *rulesParser) toStringValue(x ast.Node) (string, bool) {
	lit, ok := x.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	return unquoteNode(lit), true
}

func (p *rulesParser) toCallExpr(x ast.Node) *ast.CallExpr {
	stmt, ok := x.(*ast.ExprStmt)
	if !ok {
		return nil
	}
	call, _ := stmt.X.(*ast.CallExpr)
	return call
}

func (p *rulesParser) toFilterOperand(e ast.Expr) filterOperand {
	var o filterOperand

	if call, ok := e.(*ast.CallExpr); ok {
		o.args = call.Args
		e = call.Fun
	}
	var path string
	for {
		selector, ok := e.(*ast.SelectorExpr)
		if !ok {
			break
		}
		if path == "" {
			path = selector.Sel.Name
		} else {
			path = selector.Sel.Name + "." + path
		}
		e = selector.X
	}
	indexing, ok := e.(*ast.IndexExpr)
	if !ok {
		return o
	}
	mapIdent, ok := indexing.X.(*ast.Ident)
	if !ok {
		return o
	}
	indexString, ok := p.toStringValue(indexing.Index)
	if !ok {
		return o
	}

	o.mapName = mapIdent.Name
	o.varName = indexString
	o.path = path
	return o
}

func (p *rulesParser) errorf(n ast.Node, format string, args ...interface{}) error {
	loc := p.fset.Position(n.Pos())
	return fmt.Errorf("%s:%d: %s",
		loc.Filename, loc.Line, fmt.Sprintf(format, args...))
}

type filterOperand struct {
	mapName string
	varName string
	path    string
	args    []ast.Expr
}
