package irconv

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func TestConvFilterExpr(t *testing.T) {
	tests := []struct {
		expr string
		want string
	}{
		// Bool-typed matcher var expressions.
		{`m["x"].Pure`, `(VarPure ["x"])`},
		{`m["x"].Const`, `(VarConst ["x"])`},
		{`m["x"].Addressable`, `(VarAddressable ["x"])`},
		{`m["x"].Comparable`, `(VarComparable ["x"])`},

		// Parens should not break the conversion.
		{`(m["x"].Pure)`, `(VarPure ["x"])`},
		{`((m["x"]).Pure)`, `(VarPure ["x"])`},
		{`((m)["x"].Pure)`, `(VarPure ["x"])`},
		{`(m["x"].Value).Int() == 0`, `(Eq (VarValueInt ["x"]) 0)`},

		// Unary not.
		{`!m["x"].Pure`, `(Not (VarPure ["x"]))`},
		{`!(m["x"].Pure && m["y"].Pure)`, `(Not (And (VarPure ["x"]) (VarPure ["y"])))`},

		// AND & OR.
		{`m["x"].Pure && m["y"].Const`, `(And (VarPure ["x"]) (VarConst ["y"]))`},
		{`m["x"].Pure || m["y"].Const`, `(Or (VarPure ["x"]) (VarConst ["y"]))`},

		// Simple matcher var expressions.
		{`m["x"].Text != ""`, `(Neq (VarText ["x"]) "")`},
		{`m["x"].Value.Int() == 0`, `(Eq (VarValueInt ["x"]) 0)`},
		{`m["x"].Line > 10`, `(Gt (VarLine ["x"]) 10)`},
		{`m["x"].Line < (20 + 10)`, `(Lt (VarLine ["x"]) 30)`},
		{`m["x"].Line >= 10`, `(GtEq (VarLine ["x"]) 10)`},
		{`m["x"].Line <= (20 + 10)`, `(LtEq (VarLine ["x"]) 30)`},
		{`m["x"].Type.Size == 0`, `(Eq (VarTypeSize ["x"]) 0)`},

		// Matcher var expressions on both sides.
		{`m["x"].Type.Size == m["y"].Type.Size`, `(Eq (VarTypeSize ["x"]) (VarTypeSize ["y"]))`},

		// Matcher var methods.
		{`m["x"].Node.Is("Ident")`, `(VarNodeIs ["x"] "Ident")`},
		{`m["x"].Object.Is("Func")`, `(VarObjectIs ["x"] "Func")`},
		{`m["x"].Type.Is("foo")`, `(VarTypeIs ["x"] "foo")`},
		{`m["x"].Type.Underlying().Is("foo")`, `(VarTypeUnderlyingIs ["x"] "foo")`},
		{`m["x"].Type.ConvertibleTo("foo")`, `(VarTypeConvertibleTo ["x"] "foo")`},
		{`m["x"].Type.AssignableTo("foo")`, `(VarTypeAssignableTo ["x"] "foo")`},
		{`m["x"].Type.Implements("foo")`, `(VarTypeImplements ["x"] "foo")`},
		{`m["x"].Text.Matches("^foo")`, `(VarTextMatches ["x"] "^foo")`},

		// Matcher methods.
		{`m.File().Name.Matches("foo")`, `(FileNameMatches ["foo"])`},
		{`m.File().PkgPath.Matches("foo")`, `(FilePkgPathMatches ["foo"])`},
		{`m.File().Imports("foo")`, `(FileImports ["foo"])`},

		// Custom filters.
		{`m["x"].Filter(f)`, `(VarFilter ["x"] (FilterFuncRef ["f"]))`},
		{`m["x"].Filter(g)`, `(VarFilter ["x"] (FilterFuncRef ["g"]))`},

		// Operands order should not matter for the conversion.
		{`10 >= m["x"].Line`, `(GtEq 10 (VarLine ["x"]))`},
	}

	for _, test := range tests {
		src := fmt.Sprintf(`package gorules
import "github.com/quasilyte/go-ruleguard/dsl"

func f(ctx *dsl.VarFilterContext) bool { return false }
func g(ctx *dsl.VarFilterContext) bool { return false }

func test(m dsl.Matcher) {
	m.Match("example").Where(%s).Report("ok")
}`, test.expr)

		fset := token.NewFileSet()
		parserFlags := parser.ParseComments
		f, err := parser.ParseFile(fset, "test.go", src, parserFlags)
		if err != nil {
			t.Fatalf("parse %s file: %v", test.expr, err)
		}
		imp := importer.ForCompiler(fset, "source", nil)

		typechecker := types.Config{Importer: imp}
		types := &types.Info{
			Types: map[ast.Expr]types.TypeAndValue{},
			Uses:  map[*ast.Ident]types.Object{},
			Defs:  map[*ast.Ident]types.Object{},
		}
		pkg, err := typechecker.Check("gorules", fset, []*ast.File{f}, types)
		if err != nil {
			t.Fatalf("typecheck %s: %v", test.expr, err)
		}

		irconvCtx := &Context{
			Pkg:   pkg,
			Types: types,
			Fset:  fset,
			Src:   []byte(src),
		}
		irfile, err := ConvertFile(irconvCtx, f)
		if err != nil {
			t.Fatalf("irconv %s: %v", test.expr, err)
		}
		rule := irfile.RuleGroups[0].Rules[0]
		have := rule.WhereExpr.String()
		if have != test.want {
			t.Errorf("%s conversion:\nhave: %s\nwant: %s", test.expr, have, test.want)
			continue
		}
		if rule.WhereExpr.Src != test.expr {
			t.Errorf("%s src doesn't match (got %s)", test.expr, rule.WhereExpr.Src)
			continue
		}
	}
}
