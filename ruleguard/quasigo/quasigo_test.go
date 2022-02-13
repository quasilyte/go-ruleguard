package quasigo_test

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"

	"github.com/quasilyte/go-ruleguard/ruleguard/quasigo"
)

const testPackage = "testpkg"

type parsedTestFile struct {
	ast   *ast.File
	pkg   *types.Package
	types *types.Info
	fset  *token.FileSet
}

func parseGoFile(pkgPath, src string) (*parsedTestFile, error) {
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
	pkg, err := typechecker.Check(pkgPath, fset, []*ast.File{file}, info)
	result := &parsedTestFile{
		ast:   file,
		pkg:   pkg,
		types: info,
		fset:  fset,
	}
	return result, err
}

func compileTestFile(env *quasigo.Env, targetFunc, pkgPath string, parsed *parsedTestFile) (*quasigo.Func, error) {
	var resultFunc *quasigo.Func
	for _, decl := range parsed.ast.Decls {
		decl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if decl.Body == nil {
			continue
		}
		ctx := &quasigo.CompileContext{
			Env:     env,
			Package: parsed.pkg,
			Types:   parsed.types,
			Fset:    parsed.fset,
		}
		fn, err := quasigo.Compile(ctx, decl)
		if err != nil {
			return nil, fmt.Errorf("compile %s func: %v", decl.Name, err)
		}
		if decl.Name.String() == targetFunc {
			resultFunc = fn
		} else {
			env.AddFunc(pkgPath, decl.Name.String(), fn)
		}
	}
	return resultFunc, nil
}

func compileTestFunc(env *quasigo.Env, fn string, parsed *parsedTestFile) (*quasigo.Func, error) {
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

	ctx := &quasigo.CompileContext{
		Env:     env,
		Package: parsed.pkg,
		Types:   parsed.types,
		Fset:    parsed.fset,
	}
	return quasigo.Compile(ctx, target)
}
