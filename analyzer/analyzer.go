package analyzer

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"io/ioutil"
	"path/filepath"
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
	Analyzer.Flags.StringVar(&flagRules, "rules", "", "comma-separated list of gorule file paths")
	Analyzer.Flags.StringVar(&flagE, "e", "", "execute a single rule from a given string")
}

type parseRulesResult struct {
	rset      *ruleguard.GoRuleSet
	multiFile bool
}

func runAnalyzer(pass *analysis.Pass) (interface{}, error) {
	// TODO(quasilyte): parse config under sync.Once and
	// create rule sets from it.

	parseResult, err := readRules()
	if err != nil {
		return nil, fmt.Errorf("load rules: %v", err)
	}
	rset := parseResult.rset
	multiFile := parseResult.multiFile

	ctx := &ruleguard.Context{
		Pkg:   pass.Pkg,
		Types: pass.TypesInfo,
		Sizes: pass.TypesSizes,
		Fset:  pass.Fset,
		Report: func(info ruleguard.GoRuleInfo, n ast.Node, msg string, s *ruleguard.Suggestion) {
			if multiFile {
				msg += fmt.Sprintf(" (%s)", filepath.Base(info.Filename))
			}
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

func readRules() (*parseRulesResult, error) {
	fset := token.NewFileSet()

	switch {
	case flagRules != "":
		if flagRules == "" {
			return nil, fmt.Errorf("-rules values is empty")
		}
		filenames := strings.Split(flagRules, ",")
		var ruleSets []*ruleguard.GoRuleSet
		for _, filename := range filenames {
			filename = strings.TrimSpace(filename)
			data, err := ioutil.ReadFile(filename)
			if err != nil {
				return nil, fmt.Errorf("read rules file: %v", err)
			}
			rset, err := ruleguard.ParseRules(filename, fset, bytes.NewReader(data))
			if err != nil {
				return nil, fmt.Errorf("parse rules file: %v", err)
			}
			ruleSets = append(ruleSets, rset)
		}
		rset := ruleguard.MergeRuleSets(ruleSets)
		return &parseRulesResult{rset: rset, multiFile: len(filenames) > 1}, nil

	case flagE != "":
		ruleText := fmt.Sprintf(`
			package gorules
			import "github.com/quasilyte/go-ruleguard/dsl/fluent"
			func _(m fluent.Matcher) {
				%s.Report("$$")
			}`,
			flagE)
		r := strings.NewReader(ruleText)
		rset, err := ruleguard.ParseRules(flagRules, fset, r)
		return &parseRulesResult{rset: rset}, err

	default:
		return nil, fmt.Errorf("both -e and -rules flags are empty")
	}
}
