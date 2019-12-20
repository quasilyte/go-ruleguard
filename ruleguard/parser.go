package ruleguard

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"path"

	"github.com/quasilyte/go-ruleguard/internal/mvdan.cc/gogrep"
	"github.com/quasilyte/go-ruleguard/ruleguard/typematch"
)

type rulesParser struct {
	fset *token.FileSet
	res  *GoRuleSet

	groupImports map[string]string
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

	if f.Name.Name != "gorules" {
		return nil, fmt.Errorf("expected a gorules package name, found %s", f.Name.Name)
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
	p.groupImports = map[string]string{}

	for _, stmt := range f.Body.List {
		stmtExpr, ok := stmt.(*ast.ExprStmt)
		if !ok {
			return p.errorf(stmt, "expected a %s method call, found %s", matcher, sprintNode(p.fset, stmt))
		}
		call, ok := stmtExpr.X.(*ast.CallExpr)
		if !ok {
			return p.errorf(stmt, "expected a %s method call, found %s", matcher, sprintNode(p.fset, stmt))
		}
		if err := p.parseCall(matcher, call); err != nil {
			return err
		}

	}

	return nil
}

func (p *rulesParser) parseCall(matcher string, call *ast.CallExpr) error {
	f := call.Fun.(*ast.SelectorExpr)
	x, ok := f.X.(*ast.Ident)
	if ok && x.Name == matcher {
		return p.parseStmt(f.Sel, call.Args)
	}

	return p.parseRule(matcher, call)
}

func (p *rulesParser) parseStmt(fn *ast.Ident, args []ast.Expr) error {
	switch fn.Name {
	case "Import":
		pkgPath, ok := p.toStringValue(args[0])
		if !ok {
			return p.errorf(args[0], "expected a string literal argument")
		}
		pkgName := path.Base(pkgPath)
		p.groupImports[pkgName] = pkgPath
		return nil
	default:
		return p.errorf(fn, "unexpected %s method", fn.Name)
	}
}

func (p *rulesParser) parseRule(matcher string, call *ast.CallExpr) error {
	origCall := call
	var (
		matchArgs   *[]ast.Expr
		whereArgs   *[]ast.Expr
		suggestArgs *[]ast.Expr
		reportArgs  *[]ast.Expr
		atArgs      *[]ast.Expr
	)
	for {
		chain, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			break
		}
		switch chain.Sel.Name {
		case "Match":
			matchArgs = &call.Args
		case "Where":
			whereArgs = &call.Args
		case "Suggest":
			suggestArgs = &call.Args
		case "Report":
			reportArgs = &call.Args
		case "At":
			atArgs = &call.Args
		default:
			return p.errorf(chain.Sel, "unexpected %s method", chain.Sel.Name)
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
		return p.errorf(origCall, "missing Match() call")
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

	if suggestArgs != nil {
		s, ok := p.toStringValue((*suggestArgs)[0])
		if !ok {
			return p.errorf((*suggestArgs)[0], "expected string literal argument")
		}
		proto.suggestion = s
	}

	if reportArgs == nil {
		if suggestArgs == nil {
			return p.errorf(origCall, "missing Report() or Suggest() call")
		}
		proto.msg = "suggestion: " + proto.suggestion
	} else {
		message, ok := p.toStringValue((*reportArgs)[0])
		if !ok {
			return p.errorf((*reportArgs)[0], "expected string literal argument")
		}
		proto.msg = message
	}

	if atArgs != nil {
		index, ok := (*atArgs)[0].(*ast.IndexExpr)
		if !ok {
			return p.errorf((*atArgs)[0], "expected %s[`varname`] expression", matcher)
		}
		arg, ok := p.toStringValue(index.Index)
		if !ok {
			return p.errorf(index.Index, "expected a string literal index")
		}
		proto.location = arg
	}

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
		ctx := typematch.Context{Imports: p.groupImports}
		pat, err := typematch.Parse(&ctx, typeString)
		if err != nil {
			return p.errorf(args[0], "parse type expr: %v", err)
		}
		wantIdentical := !negate
		filter.typePred = func(x types.Type) bool {
			return wantIdentical == pat.MatchIdentical(x)
		}
		dst[operand.varName] = filter
	case "Type.ConvertibleTo":
		if len(args) != 1 {
			return p.errorf(e, "Type.ConvertibleTo() expects exactly 1 argument, %d given", len(args))
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
		wantConvertible := !negate
		filter.typePred = func(x types.Type) bool {
			return wantConvertible == types.ConvertibleTo(x, y)
		}
		dst[operand.varName] = filter
	case "Type.AssignableTo":
		if len(args) != 1 {
			return p.errorf(e, "Type.AssignableTo() expects exactly 1 argument, %d given", len(args))
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
