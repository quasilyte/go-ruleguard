package analyzer

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"io/ioutil"
	"os"
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
	flagRules   string
	flagE       string
	flagEnable  string
	flagDisable string

	flagDebug              string
	flagDebugImports       bool
	flagDebugEnableDisable bool
)

func init() {
	Analyzer.Flags.StringVar(&flagRules, "rules", "", "comma-separated list of gorule file paths")
	Analyzer.Flags.StringVar(&flagE, "e", "", "execute a single rule from a given string")
	Analyzer.Flags.StringVar(&flagDebug, "debug-group", "", "enable debug for the specified function")
	Analyzer.Flags.StringVar(&flagEnable, "enable", "<all>", "comma-separated list of enabled groups or '<all>' to enable everything")
	Analyzer.Flags.StringVar(&flagDisable, "disable", "", "comma-separated list of groups to be disabled")
	Analyzer.Flags.BoolVar(&flagDebugImports, "debug-imports", false, "enable debug for rules compile-time package lookups")
	Analyzer.Flags.BoolVar(&flagDebugEnableDisable, "debug-enable-disable", false, "enable debug for -enable/-disable related info")
}

type parseRulesResult struct {
	rset      *ruleguard.GoRuleSet
	multiFile bool
}

func debugPrint(s string) {
	fmt.Fprintln(os.Stderr, s)
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
		Debug:      flagDebug,
		DebugPrint: debugPrint,
		Pkg:        pass.Pkg,
		Types:      pass.TypesInfo,
		Sizes:      pass.TypesSizes,
		Fset:       pass.Fset,
		Report: func(info ruleguard.GoRuleInfo, n ast.Node, msg string, s *ruleguard.Suggestion) {
			msg = info.Group + ": " + msg
			if multiFile {
				msg += fmt.Sprintf(" (%s:%d)", filepath.Base(info.Filename), info.Line)
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

	disabledGroups := make(map[string]bool)
	enabledGroups := make(map[string]bool)
	for _, g := range strings.Split(flagDisable, ",") {
		g = strings.TrimSpace(g)
		disabledGroups[g] = true
	}
	if flagEnable != "<all>" {
		for _, g := range strings.Split(flagEnable, ",") {
			g = strings.TrimSpace(g)
			enabledGroups[g] = true
		}
	}

	ctx := &ruleguard.ParseContext{
		Fset:         fset,
		DebugImports: flagDebugImports,
		DebugPrint:   debugPrint,
		GroupFilter: func(g string) bool {
			whyDisabled := ""
			enabled := flagEnable == "<all>" || enabledGroups[g]
			switch {
			case !enabled:
				whyDisabled = "not enabled by -enabled flag"
			case disabledGroups[g]:
				whyDisabled = "disabled by -disable flag"
			}
			if flagDebugEnableDisable {
				if whyDisabled != "" {
					debugPrint(fmt.Sprintf("(-) %s is %s", g, whyDisabled))
				} else {
					debugPrint(fmt.Sprintf("(+) %s is enabled", g))
				}
			}
			return whyDisabled == ""
		},
	}

	switch {
	case flagRules != "":
		filenames := strings.Split(flagRules, ",")
		multifile := len(filenames) > 1
		var ruleSets []*ruleguard.GoRuleSet
		for _, filename := range filenames {
			filename = strings.TrimSpace(filename)
			data, err := ioutil.ReadFile(filename)
			if err != nil {
				return nil, fmt.Errorf("read rules file: %v", err)
			}
			rset, err := ruleguard.ParseRules(ctx, filename, bytes.NewReader(data))
			if err != nil {
				return nil, fmt.Errorf("parse rules file: %v", err)
			}
			if len(rset.Imports) != 0 {
				multifile = true
			}
			ruleSets = append(ruleSets, rset)
		}
		rset, err := ruleguard.MergeRuleSets(ruleSets)
		if err != nil {
			return nil, fmt.Errorf("merge rule files: %v", err)
		}
		return &parseRulesResult{rset: rset, multiFile: multifile}, nil

	case flagE != "":
		ruleText := fmt.Sprintf(`
			package gorules
			import "github.com/quasilyte/go-ruleguard/dsl"
			func e(m dsl.Matcher) {
				%s.Report("$$")
			}`,
			flagE)
		r := strings.NewReader(ruleText)
		rset, err := ruleguard.ParseRules(ctx, flagRules, r)
		return &parseRulesResult{rset: rset}, err

	default:
		return nil, fmt.Errorf("both -e and -rules flags are empty")
	}
}
