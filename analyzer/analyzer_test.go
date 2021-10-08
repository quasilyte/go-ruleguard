package analyzer_test

import (
	"bytes"
	"fmt"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quasilyte/go-ruleguard/analyzer"
	"github.com/quasilyte/go-ruleguard/ruleguard/goutil"
	"github.com/quasilyte/go-ruleguard/ruleguard/irconv"
	"github.com/quasilyte/go-ruleguard/ruleguard/irprint"
	"golang.org/x/tools/go/analysis/analysistest"
)

var tests = []string{
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

func TestAnalyzer(t *testing.T) {
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

func TestPrintIR(t *testing.T) {
	analyzerTemplate := `
package main

import (
	"github.com/quasilyte/go-ruleguard/analyzer/testanalyzer"
	"github.com/quasilyte/go-ruleguard/ruleguard/ir"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	analyzer := testanalyzer.New(&rulesFile)
	singlechecker.Main(analyzer)
}

var rulesFile = %s
`

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	{
		args := []string{
			"build",
			"-o", "test-ruleguard",
			filepath.Join(wd, "..", "cmd", "ruleguard"),
		}
		out, err := exec.Command("go", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("build go-ruleguard: %v: %s", err, out)
		}
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			rulesFilename := filepath.Join(wd, "testdata", "src", test, "rules.go")
			data, err := ioutil.ReadFile(rulesFilename)
			if err != nil {
				t.Fatalf("%s: %v", test, err)
			}
			fset := token.NewFileSet()
			f, err := goutil.LoadGoFile(goutil.LoadConfig{
				Fset:     fset,
				Filename: rulesFilename,
				Data:     data,
			})
			if err != nil {
				t.Fatalf("%s: %v", test, err)
			}
			ctx := &irconv.Context{
				Pkg:   f.Pkg,
				Types: f.Types,
				Fset:  fset,
				Src:   data,
			}
			irfile, err := irconv.ConvertFile(ctx, f.Syntax)
			if err != nil {
				t.Fatalf("%s: irconv: %v", test, err)
			}
			var irfileBuf bytes.Buffer
			irprint.File(&irfileBuf, irfile)
			mainFile, err := ioutil.TempFile("", "ruleguard-test*.go")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(mainFile.Name())
			_, err = mainFile.WriteString(fmt.Sprintf(analyzerTemplate, irfileBuf.String()))
			if err != nil {
				t.Fatal(err)
			}

			srcRulesCmd := exec.Command(filepath.Join(wd, "test-ruleguard"), "-rules", rulesFilename, "./...") // nolint:gosec
			srcRulesCmd.Dir = filepath.Join(wd, "testdata", "src", test)
			srcOut, _ := srcRulesCmd.CombinedOutput()

			{
				args := []string{
					"build",
					"-o", "test-ruleguard-ir",
					mainFile.Name(),
				}
				out, err := exec.Command("go", args...).CombinedOutput()
				if err != nil {
					t.Fatalf("build go-ruleguard IR: %v: %s", err, out)
				}
			}

			irRulesCmd := exec.Command(filepath.Join(wd, "test-ruleguard-ir"), "./...") // nolint:gosec
			irRulesCmd.Dir = filepath.Join(wd, "testdata", "src", test)
			irOut, _ := irRulesCmd.CombinedOutput()

			if diff := cmp.Diff(string(irOut), string(srcOut)); diff != "" {
				t.Errorf("%s output mismatches: %s", test, diff)
			}

		})
	}
}
