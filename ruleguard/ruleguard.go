package ruleguard

import (
	"go/ast"
	"go/token"
	"go/types"
	"io"

	"github.com/quasilyte/go-ruleguard/ruleguard/typematch"
)

type ParseContext struct {
	DebugImports bool
	DebugPrint   func(string)

	// GroupFilter is called for every rule group being parsed.
	// If function returns false, that group will not be included
	// in the resulting rules set.
	// Nil filter accepts all rule groups.
	GroupFilter func(string) bool

	Fset *token.FileSet
}

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

func ParseRules(ctx *ParseContext, filename string, r io.Reader) (*GoRuleSet, error) {
	config := rulesParserConfig{
		ctx:      ctx,
		itab:     typematch.NewImportsTab(stdlibPackages),
		importer: newGoImporter(ctx),
	}
	p := newRulesParser(config)
	return p.ParseFile(filename, r)
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
