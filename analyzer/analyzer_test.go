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
	}

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

func TestAnalyzer_Rules_RemoteConfig(t *testing.T) {
	test := "rules"
	url := "https://raw.githubusercontent.com/quasilyte/go-ruleguard/master/rules.go"
	testdata := analysistest.TestData()
	err := analyzer.Analyzer.Flags.Set("rules", url)
	if err != nil {
		t.Fatalf("set rules flag: %v", err)
	}
	analysistest.Run(t, testdata, analyzer.Analyzer, test)
}

func TestAnalyzer_Rules_LocalConfig(t *testing.T) {
	test := "rules"
	url := "../rules.go"
	testdata := analysistest.TestData()
	err := analyzer.Analyzer.Flags.Set("rules", url)
	if err != nil {
		t.Fatalf("set rules flag: %v", err)
	}
	analysistest.Run(t, testdata, analyzer.Analyzer, test)
}
