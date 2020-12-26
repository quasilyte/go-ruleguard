package ruleguard

import (
	"go/ast"
	"go/token"
	"go/types"
	"io"

	"github.com/quasilyte/go-ruleguard/ruleguard/typematch"
)

type Context struct {
	Debug      string
	DebugPrint func(string)

	Types  *types.Info
	Sizes  types.Sizes
	Fset   *token.FileSet
	Report func(rule GoRuleInfo, n ast.Node, msg string, s *Suggestion)
	Pkg    *types.Package
}

type Suggestion struct {
	From        token.Pos
	To          token.Pos
	Replacement []byte
}

func ParseRules(filename string, fset *token.FileSet, r io.Reader) (*GoRuleSet, error) {
	config := rulesParserConfig{
		itab:     typematch.NewImportsTab(stdlibPackages),
		importer: newGoImporter(fset),
	}
	p := newRulesParser(config)
	return p.ParseFile(filename, fset, r)
}

func RunRules(ctx *Context, f *ast.File, rules *GoRuleSet) error {
	return newRulesRunner(ctx, rules).run(f)
}

type GoRuleInfo struct {
	// Filename is a file that defined this rule.
	Filename string

	// Line is a line inside a file that defined this rule.
	Line int

	// Group is a function name that contained this rule.
	Group string
}

type GoRuleSet struct {
	universal *scopedGoRuleSet
	local     *scopedGoRuleSet

	groups map[string]token.Position // To handle redefinitions

	// Imports is a set of rule bundles that were imported.
	Imports map[string]struct{}
}

func MergeRuleSets(toMerge []*GoRuleSet) (*GoRuleSet, error) {
	return mergeRuleSets(toMerge)
}
