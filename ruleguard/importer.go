package ruleguard

import (
	"go/importer"
	"go/token"
	"go/types"
)

type goImporter struct {
	// TODO(quasilyte): share importers with gogrep?

	gcImporter  types.Importer
	srcImporter types.Importer
}

func newGoImporter(fset *token.FileSet) *goImporter {
	return &goImporter{
		gcImporter:  importer.Default(),
		srcImporter: importer.ForCompiler(fset, "source", nil),
	}
}

func (imp *goImporter) Import(path string) (*types.Package, error) {
	pkg, err := imp.srcImporter.Import(path)
	if err == nil {
		return pkg, nil
	}
	return imp.gcImporter.Import(path)
}

func (imp *goImporter) ImportFrom(path, dir string, mode types.ImportMode) (*types.Package, error) {
	if srcImporter, ok := imp.srcImporter.(types.ImporterFrom); ok {
		pkg, err := srcImporter.ImportFrom(path, dir, mode)
		if err == nil {
			return pkg, nil
		}
	}
	if gcImporter, ok := imp.gcImporter.(types.ImporterFrom); ok {
		return gcImporter.ImportFrom(path, dir, mode)
	}
	return imp.Import(path)
}
