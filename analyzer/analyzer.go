package analyzer

import (
	"bytes"
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/tools/go/analysis"

	"github.com/quasilyte/go-ruleguard/ruleguard"
)

// Version contains extra version info.
// It's initialized via ldflags -X when ruleguard is built with Make.
// Can contain a git hash (dev builds) or a version tag (release builds).
var Version string

func docString() string {
	doc := "execute dynamic gogrep-based rules"
	if Version == "" {
		return doc
	}
	return doc + " (" + Version + ")"
}

// Analyzer exports ruleguard as an analysis-compatible object.
var Analyzer = &analysis.Analyzer{
	Name: "ruleguard",
	Doc:  docString(),
	Run:  runAnalyzer,
}

// ForceNewEngine disables engine cache optimization.
// This should only be useful for analyzer testing.
var ForceNewEngine = false

var runnerStatePool sync.Pool

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

	flagGoVersion string

	flagDebug              string
	flagDebugFunc          string
	flagDebugImports       bool
	flagDebugEnableDisable bool
)

func init() {
	Analyzer.Flags.StringVar(&flagDebugFunc, "debug-func", "", "[experimental!] enable debug for the specified bytecode function")
	Analyzer.Flags.StringVar(&flagDebug, "debug-group", "", "[experimental!] enable debug for the specified matcher function")
	Analyzer.Flags.BoolVar(&flagDebugImports, "debug-imports", false, "[experimental!] enable debug for rules compile-time package lookups")
	Analyzer.Flags.BoolVar(&flagDebugEnableDisable, "debug-enable-disable", false, "[experimental!] enable debug for -enable/-disable related info")

	Analyzer.Flags.StringVar(&flagGoVersion, "go", "", "select the Go version to target; leave as string for the latest")

	Analyzer.Flags.StringVar(&flagRules, "rules", "", "comma-separated list of ruleguard file paths")
	Analyzer.Flags.StringVar(&flagE, "e", "", "execute a single rule from a given string")
	Analyzer.Flags.StringVar(&flagEnable, "enable", "<all>", "comma-separated list of enabled groups or '<all>' to enable everything")
	Analyzer.Flags.StringVar(&flagDisable, "disable", "", "comma-separated list of groups to be disabled")
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

	goVersion, err := ruleguard.ParseGoVersion(flagGoVersion)
	if err != nil {
		return nil, fmt.Errorf("parse Go version: %w", err)
	}

	ctx := &ruleguard.RunContext{
		Debug:        flagDebug,
		DebugImports: flagDebugImports,
		DebugPrint:   debugPrint,
		Pkg:          pass.Pkg,
		Types:        pass.TypesInfo,
		Sizes:        pass.TypesSizes,
		Fset:         pass.Fset,
		GoVersion:    goVersion,
		Report: func(data *ruleguard.ReportData) {
			fullMessage := data.Message
			info := data.RuleInfo
			if printRuleLocation {
				fullMessage = fmt.Sprintf("%s: %s (%s:%d)",
					info.Group.Name, data.Message, filepath.Base(info.Group.Filename), info.Line)
			}
			diag := analysis.Diagnostic{
				Pos:     data.Node.Pos(),
				Message: fullMessage,
			}
			if data.Suggestion != nil {
				s := data.Suggestion
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

	if runnerStatePool.New != nil {
		state := runnerStatePool.Get().(*ruleguard.RunnerState)
		ctx.State = state
		defer func() {
			runnerStatePool.Put(state)
		}()
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
	runnerStatePool = sync.Pool{
		New: func() interface{} {
			return ruleguard.NewRunnerState(globalEngine)
		},
	}
	return engine, nil
}

func newEngine() (*ruleguard.Engine, error) {
	e := ruleguard.NewEngine()
	e.InferBuildContext()
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

	ctx := &ruleguard.LoadContext{
		Fset:         fset,
		DebugFunc:    flagDebugFunc,
		DebugImports: flagDebugImports,
		DebugPrint:   debugPrint,
		GroupFilter: func(g *ruleguard.GoRuleGroup) bool {
			whyDisabled := ""
			enabled := flagEnable == "<all>" || enabledGroups[g.Name]
			switch {
			case !enabled:
				whyDisabled = "not enabled by -enabled flag"
			case disabledGroups[g.Name]:
				whyDisabled = "disabled by -disable flag"
			}
			if flagDebugEnableDisable {
				if whyDisabled != "" {
					debugPrint(fmt.Sprintf("(-) %s is %s", g.Name, whyDisabled))
				} else {
					debugPrint(fmt.Sprintf("(+) %s is enabled", g.Name))
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
			data, err := os.ReadFile(filename)
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
