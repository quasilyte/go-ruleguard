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
	"sync"

	"github.com/quasilyte/go-ruleguard/ruleguard"
	"golang.org/x/tools/go/analysis"
)

// Version contains extra version info.
// It's is initialized via ldflags -X when ruleguard is built with Make.
// Can contain a git hash (dev builds) or a version tag (release builds).
var Version string

func docString() string {
	doc := "execute dynamic gogrep-based rules"
	if Version == "" {
		return doc
	}
	return doc + " (" + Version + ")"
}

// Analyzer exports ruleguard as a analysis-compatible object.
var Analyzer = &analysis.Analyzer{
	Name: "ruleguard",
	Doc:  docString(),
	Run:  runAnalyzer,
}

// ForceNewEngine disables engine cache optimization.
// This should only be useful for analyzer testing.
var ForceNewEngine = false

var (
	globalEngineMu      sync.Mutex
	globalEngine        *ruleguard.Engine
	globalEngineErrored bool
)

var (
	flagRules   string
	flagE       string
	flagEnable  string
	flagDisable string

	flagDebug              string
	flagDebugFilter        string
	flagDebugImports       bool
	flagDebugEnableDisable bool
)

func init() {
	Analyzer.Flags.StringVar(&flagRules, "rules", "", "comma-separated list of gorule file paths")
	Analyzer.Flags.StringVar(&flagE, "e", "", "execute a single rule from a given string")
	Analyzer.Flags.StringVar(&flagDebug, "debug-group", "", "enable debug for the specified matcher function")
	Analyzer.Flags.StringVar(&flagDebugFilter, "debug-filter", "", "enable debug for the specified filter function")
	Analyzer.Flags.StringVar(&flagEnable, "enable", "<all>", "comma-separated list of enabled groups or '<all>' to enable everything")
	Analyzer.Flags.StringVar(&flagDisable, "disable", "", "comma-separated list of groups to be disabled")
	Analyzer.Flags.BoolVar(&flagDebugImports, "debug-imports", false, "enable debug for rules compile-time package lookups")
	Analyzer.Flags.BoolVar(&flagDebugEnableDisable, "debug-enable-disable", false, "enable debug for -enable/-disable related info")
}

func debugPrint(s string) {
	fmt.Fprintln(os.Stderr, s)
}

func runAnalyzer(pass *analysis.Pass) (interface{}, error) {
	engine, err := prepareEngine()
	if err != nil {
		return nil, fmt.Errorf("load rules: %v", err)
	}
	// This condition will trigger only if we failed to init
	// the engine. Return without an error as other analysis
	// pass probably reported init error by this moment.
	if engine == nil {
		return nil, nil
	}

	printRuleLocation := flagE == ""

	ctx := &ruleguard.RunContext{
		Debug:        flagDebug,
		DebugImports: flagDebugImports,
		DebugPrint:   debugPrint,
		Pkg:          pass.Pkg,
		Types:        pass.TypesInfo,
		Sizes:        pass.TypesSizes,
		Fset:         pass.Fset,
		Report: func(info ruleguard.GoRuleInfo, n ast.Node, msg string, s *ruleguard.Suggestion) {
			fullMessage := info.Group + ": " + msg
			if printRuleLocation {
				fullMessage += fmt.Sprintf(" (%s:%d)", filepath.Base(info.Filename), info.Line)
			}
			diag := analysis.Diagnostic{
				Pos:     n.Pos(),
				Message: fullMessage,
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
		if err := engine.Run(ctx, f); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func prepareEngine() (*ruleguard.Engine, error) {
	if ForceNewEngine {
		return newEngine()
	}

	globalEngineMu.Lock()
	defer globalEngineMu.Unlock()

	if globalEngine != nil {
		return globalEngine, nil
	}
	// If we already failed once, don't try again to avoid #167.
	if globalEngineErrored {
		return nil, nil
	}

	engine, err := newEngine()
	if err != nil {
		globalEngineErrored = true
		return nil, err
	}
	globalEngine = engine
	return engine, nil
}

func newEngine() (*ruleguard.Engine, error) {
	e := ruleguard.NewEngine()
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
		DebugFilter:  flagDebugFilter,
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
		for _, filename := range filenames {
			filename = strings.TrimSpace(filename)
			data, err := ioutil.ReadFile(filename)
			if err != nil {
				return nil, fmt.Errorf("read rules file: %v", err)
			}
			if err := e.Load(ctx, filename, bytes.NewReader(data)); err != nil {
				return nil, fmt.Errorf("parse rules file: %v", err)
			}
		}
		return e, nil

	case flagE != "":
		ruleText := fmt.Sprintf(`
			package gorules
			import "github.com/quasilyte/go-ruleguard/dsl"
			func e(m dsl.Matcher) {
				%s.Report("$$")
			}`,
			flagE)
		r := strings.NewReader(ruleText)
		err := e.Load(ctx, "e", r)
		if err != nil {
			return nil, err
		}
		return e, nil

	default:
		return nil, fmt.Errorf("both -e and -rules flags are empty")
	}
}
