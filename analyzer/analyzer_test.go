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
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			testdata := analysistest.TestData()
			rulesFilename := fmt.Sprintf("./testdata/src/%s/rules.go", test)
			analyzer.Analyzer.Flags.Set("rules", rulesFilename)
			analysistest.Run(t, testdata, analyzer.Analyzer, test)
		})
	}
}
