package testanalyzer

import (
	"fmt"
	"go/token"
	"path/filepath"

	"github.com/quasilyte/go-ruleguard/ruleguard"
	"github.com/quasilyte/go-ruleguard/ruleguard/ir"
	"golang.org/x/tools/go/analysis"
)

func New(ruleSet *ir.File) *analysis.Analyzer {
	runAnalyzer := func(pass *analysis.Pass) (interface{}, error) {
		e := ruleguard.NewEngine()
		fset := token.NewFileSet()
		ctx := &ruleguard.LoadContext{
			Fset: fset,
		}
		if err := e.LoadFromIR(ctx, "rules.go", ruleSet); err != nil {
			return nil, err
		}
		runCtx := &ruleguard.RunContext{
			Pkg:   pass.Pkg,
			Types: pass.TypesInfo,
			Sizes: pass.TypesSizes,
			Fset:  pass.Fset,
			Report: func(data *ruleguard.ReportData) {
				info := data.RuleInfo
				msg := data.Message
				fullMessage := fmt.Sprintf("%s: %s (%s:%d)",
					info.Group.Name, msg, filepath.Base(info.Group.Filename), info.Line)
				pass.Report(analysis.Diagnostic{
					Pos:     data.Node.Pos(),
					Message: fullMessage,
				})
			},
		}
		for _, f := range pass.Files {
			if err := e.Run(runCtx, f); err != nil {
				return nil, err
			}
		}
		return nil, nil
	}

	var analyzer = &analysis.Analyzer{
		Name: "ruleguard",
		Run:  runAnalyzer,
		Doc:  "execute dynamic gogrep-based rules",
	}

	return analyzer
}
