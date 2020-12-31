package gorules_test

import (
	"testing"

	"github.com/quasilyte/go-ruleguard/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestRules(t *testing.T) {
	testdata := analysistest.TestData()
	rules := "diag.go,refactor.go,style.go"
	if err := analyzer.Analyzer.Flags.Set("rules", rules); err != nil {
		t.Fatalf("set rules flag: %v", err)
	}
	analysistest.Run(t, testdata, analyzer.Analyzer, "./...")
}
