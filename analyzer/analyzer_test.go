package analyzer_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
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

type RulesHandler struct {
	testdata string
}

func (h *RulesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fpath := path.Join(h.testdata, "../../rules.go")
	http.ServeFile(w, r, fpath)
}

func TestAnalyzer_Rules_RemoteConfig(t *testing.T) {
	handler := RulesHandler{
		testdata: analysistest.TestData(),
	}
	server := httptest.NewServer(&handler)
	defer server.Close()

	err := analyzer.Analyzer.Flags.Set("rules", server.URL)
	if err != nil {
		t.Fatalf("set rules flag: %v", err)
	}
	analysistest.Run(t, handler.testdata, analyzer.Analyzer, "rules")
}

func TestAnalyzer_Rules_LocalConfig(t *testing.T) {
	test := "rules"
	filename := "../rules.go"
	testdata := analysistest.TestData()
	err := analyzer.Analyzer.Flags.Set("rules", filename)
	if err != nil {
		t.Fatalf("set rules flag: %v", err)
	}
	analysistest.Run(t, testdata, analyzer.Analyzer, test)
}
