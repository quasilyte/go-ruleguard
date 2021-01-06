package quasigo

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
)

const testPackage = "testpkg"

type parsedTestFile struct {
	ast   *ast.File
	types *types.Info
	fset  *token.FileSet
}

func parseGoFile(src string) (*parsedTestFile, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		return nil, err
	}
	typechecker := &types.Config{
		Importer: importer.ForCompiler(fset, "source", nil),
	}
	info := &types.Info{
		Types: map[ast.Expr]types.TypeAndValue{},
		Uses:  map[*ast.Ident]types.Object{},
		Defs:  map[*ast.Ident]types.Object{},
	}
	_, err = typechecker.Check(testPackage, fset, []*ast.File{file}, info)
	result := &parsedTestFile{
		ast:   file,
		types: info,
		fset:  fset,
	}
	return result, err
}

func compileTestFunc(env *Env, fn string, parsed *parsedTestFile) (*Func, error) {
	var target *ast.FuncDecl
	for _, decl := range parsed.ast.Decls {
		decl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if decl.Name.String() == fn {
			target = decl
			break
		}
	}
	if target == nil {
		return nil, fmt.Errorf("test function %s not found", fn)
	}

	ctx := &CompileContext{
		Env:   env,
		Types: parsed.types,
		Fset:  parsed.fset,
	}
	return Compile(ctx, target)
}
