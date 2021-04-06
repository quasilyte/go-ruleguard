package analyzer_test

import (
	"fmt"
	"testing"

	"github.com/quasilyte/go-ruleguard/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	tests := []string{
		"gocritic",
		"filtertest",
		"extra",
		"suggest",
		"namedtype/nested",
		"namedtype",
		"revive",
		"golint",
		"regression",
		"testvendored",
		"quasigo",
		"matching",
		"dgryski",
		"comments",
	}

	analyzer.ForceNewEngine = true
	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			testdata := analysistest.TestData()
			rulesFilename := fmt.Sprintf("./testdata/src/%s/rules.go", test)
			if err := analyzer.Analyzer.Flags.Set("rules", rulesFilename); err != nil {
				t.Fatalf("set rules flag: %v", err)
			}
			analysistest.Run(t, testdata, analyzer.Analyzer, test)
		})
	}
}
