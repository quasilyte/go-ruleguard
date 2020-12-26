package ruleguard

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"

	"github.com/quasilyte/go-ruleguard/internal/golist"
)

// goImporter is a `types.Importer` that tries to load a package no matter what.
// It iterates through multiple import strategies and accepts whatever succeeds first.
type goImporter struct {
	// TODO(quasilyte): share importers with gogrep?

	// cache contains all imported packages, from any importer.
	// Both default and source importers have their own caches,
	// but since we use several importers, it's better to
	// have our own, unified cache.
	cache map[string]*types.Package

	fset *token.FileSet

	defaultImporter types.Importer
	srcImporter     types.Importer
}

func newGoImporter(fset *token.FileSet) *goImporter {
	return &goImporter{
		cache:           make(map[string]*types.Package),
		fset:            fset,
		defaultImporter: importer.Default(),
		srcImporter:     importer.ForCompiler(fset, "source", nil),
	}
}

func (imp *goImporter) Import(path string) (*types.Package, error) {
	if pkg := imp.cache[path]; pkg != nil {
		return pkg, nil
	}

	pkg, err1 := imp.defaultImporter.Import(path)
	if err1 == nil {
		imp.cache[path] = pkg
		return pkg, nil
	}

	pkg, err2 := imp.srcImporter.Import(path)
	if err2 == nil {
		imp.cache[path] = pkg
		return pkg, nil
	}

	// Fallback to `go list` as a last resort.
	pkg, err3 := imp.golistImport(path)
	if err3 == nil {
		imp.cache[path] = pkg
		return pkg, nil
	}

	return nil, err1
}

func (imp *goImporter) golistImport(path string) (*types.Package, error) {
	golistPkg, err := golist.JSON(path)
	if err != nil {
		return nil, err
	}

	files := make([]*ast.File, 0, len(golistPkg.GoFiles))
	for _, filename := range golistPkg.GoFiles {
		fullname := filepath.Join(golistPkg.Dir, filename)
		f, err := parser.ParseFile(imp.fset, fullname, nil, 0)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	// TODO: do we want to assign imp as importer for this nested typecherker?
	// Otherwise it won't be able to resolve imports.
	var typecheker types.Config
	var info types.Info
	return typecheker.Check(path, imp.fset, files, &info)
}
