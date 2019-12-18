package ruleguard

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"

	"github.com/quasilyte/go-ruleguard/internal/mvdan.cc/gogrep"
	"github.com/quasilyte/go-ruleguard/ruleguard/typematch"
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
		return nil, fmt.Errorf("parser error: %v", err)
	}

	typechecker := types.Config{Importer: importer.Default()}
	_, err = typechecker.Check("gorules", fset, []*ast.File{f}, nil)
	if err != nil {
		return nil, fmt.Errorf("typechecker error: %v", err)
	}

	for _, decl := range f.Decls {
		decl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if err := p.parseRuleGroup(decl); err != nil {
			return nil, err
		}
	}

	return p.res, nil
}

func (p *rulesParser) parseRuleGroup(f *ast.FuncDecl) error {
	if f.Body == nil {
		return p.errorf(f, "unexpected empty function body")
	}
	if f.Type.Results != nil {
		return p.errorf(f.Type.Results, "rule group function should not return anything")
	}
	params := f.Type.Params.List
	if len(params) != 1 || len(params[0].Names) != 1 {
		return p.errorf(f.Type.Params, "rule group function should accept exactly 1 Matcher param")
	}
	// TODO(quasilyte): do an actual matcher param type check?
	matcher := params[0].Names[0].Name

	for _, stmt := range f.Body.List {
		if err := p.parseRule(matcher, stmt); err != nil {
			return err
		}
	}

	return nil
}

func (p *rulesParser) parseRule(matcher string, stmt ast.Stmt) error {
	stmtExpr, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return p.errorf(stmt, "expected a %s.Match method call, found %s", matcher, sprintNode(p.fset, stmt))
	}
	call, ok := stmtExpr.X.(*ast.CallExpr)
	if !ok {
		return p.errorf(stmt, "expected a %s.Match method call, found %s", matcher, sprintNode(p.fset, stmt))
	}

	var (
		matchArgs  *[]ast.Expr
		whereArgs  *[]ast.Expr
		reportArgs *[]ast.Expr
	)
	for {
		chain, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			break
		}
		switch chain.Sel.Name {
		case "Report":
			reportArgs = &call.Args
		case "Where":
			whereArgs = &call.Args
		case "Match":
			matchArgs = &call.Args
		}
		call, ok = chain.X.(*ast.CallExpr)
		if !ok {
			break
		}
	}

	dst := p.res.universal
	filters := map[string]submatchFilter{}
	proto := goRule{
		filters: filters,
	}
	var alternatives []string

	if matchArgs == nil {
		return p.errorf(call, "missing Match() call")
	}
	for _, arg := range *matchArgs {
		alt, ok := p.toStringValue(arg)
		if !ok {
			return p.errorf(arg, "expected a string literal argument")
		}
		alternatives = append(alternatives, alt)
	}

	if whereArgs != nil {
		if err := p.walkFilter(filters, (*whereArgs)[0], false); err != nil {
			return err
		}
	}

	if reportArgs == nil {
		return p.errorf(call, "missing Report() call")
	}
	message, ok := p.toStringValue((*reportArgs)[0])
	if !ok {
		return p.errorf((*reportArgs)[0], "expected string literal argument")
	}
	proto.msg = message

	for i, alt := range alternatives {
		rule := proto
		pat, err := gogrep.Parse(p.fset, alt)
		if err != nil {
			return p.errorf((*matchArgs)[i], "gogrep parse: %v", err)
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

	return nil
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
		pat, err := typematch.Parse(typeString)
		if err != nil {
			return p.errorf(args[0], "parse type expr: %v", err)
		}
		wantIdentical := !negate
		filter.typePred = func(x types.Type) bool {
			return wantIdentical == pat.MatchIdentical(x)
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

func (p *rulesParser) toStringValue(x ast.Node) (string, bool) {
	lit, ok := x.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	return unquoteNode(lit), true
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
