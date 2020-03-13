package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"os"
	"strings"

	"github.com/quasilyte/go-ruleguard/ruleguard"
	"golang.org/x/tools/go/analysis"
)

// Analyzer exports ruleguard as a analysis-compatible object.
var Analyzer = &analysis.Analyzer{
	Name: "ruleguard",
	Doc:  "execute dynamic gogrep-based rules",
	Run:  runAnalyzer,
}

var (
	flagRules string
	flagE     string
)

func init() {
	Analyzer.Flags.StringVar(&flagRules, "rules", "", "path to a gorules file")
	Analyzer.Flags.StringVar(&flagE, "e", "", "execute a single rule from a given string")
}

func runAnalyzer(pass *analysis.Pass) (interface{}, error) {
	// TODO(quasilyte): parse config under sync.Once and
	// create rule sets from it.

	rset, err := readRules()
	if err != nil {
		return nil, fmt.Errorf("load rules: %v", err)
	}

	ctx := &ruleguard.Context{
		Pkg:   pass.Pkg,
		Types: pass.TypesInfo,
		Sizes: pass.TypesSizes,
		Fset:  pass.Fset,
		Report: func(n ast.Node, msg string, s *ruleguard.Suggestion) {
			diag := analysis.Diagnostic{
				Pos:     n.Pos(),
				Message: msg,
			}
			if s != nil {
				diag.SuggestedFixes = []analysis.SuggestedFix{
					{
						Message: "suggested replacement",
						TextEdits: []analysis.TextEdit{
							{
								Pos:     s.From,
								End:     s.To,
								NewText: s.Replacement,
							},
						},
					},
				}
			}
			pass.Report(diag)
		},
	}

	for _, f := range pass.Files {
		if err := ruleguard.RunRules(ctx, f, rset); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func readRules() (*ruleguard.GoRuleSet, error) {
	var r io.Reader

	switch {
	case flagRules != "":
		if flagRules == "" {
			return nil, fmt.Errorf("-rules values is empty")
		}
		f, err := os.Open(flagRules)
		if err != nil {
			return nil, fmt.Errorf("open rules file: %v", err)
		}
		defer f.Close()
		r = f
	case flagE != "":
		ruleText := fmt.Sprintf(`
			package gorules
			import "github.com/quasilyte/go-ruleguard/dsl/fluent"
			func _(m fluent.Matcher) {
				%s.Report("$$")
			}`,
			flagE)
		r = strings.NewReader(ruleText)
	default:
		return nil, fmt.Errorf("both -e and -rules flags are empty")
	}

	fset := token.NewFileSet()
	return ruleguard.ParseRules(flagRules, fset, r)
}
