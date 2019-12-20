package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"

	"github.com/quasilyte/go-ruleguard/ruleguard"
	"golang.org/x/tools/go/analysis"
)

// Analyzer exports ruleguard as a analysis-compatible object.
var Analyzer = &analysis.Analyzer{
	Name: "ruleguard",
	Doc:  "execute dynamic gogrep-based rules",
	Run:  runAnalyzer,
}

var flagRules string

func init() {
	Analyzer.Flags.StringVar(&flagRules, "rules", "", "path to a gorules file")
}

func runAnalyzer(pass *analysis.Pass) (interface{}, error) {
	// TODO(quasilyte): parse config under sync.Once and
	// create rule sets from it.

	fset := token.NewFileSet()
	if flagRules == "" {
		return nil, fmt.Errorf("-rules values is empty")
	}
	f, err := os.Open(flagRules)
	if err != nil {
		return nil, fmt.Errorf("open rules file: %v", err)
	}
	defer f.Close()
	rset, err := ruleguard.ParseRules(flagRules, fset, f)
	if err != nil {
		return nil, fmt.Errorf("parse rules file: %v", err)
	}

	ctx := &ruleguard.Context{
		Types: pass.TypesInfo,
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
		ruleguard.RunRules(ctx, f, rset)
	}

	return nil, nil
}
