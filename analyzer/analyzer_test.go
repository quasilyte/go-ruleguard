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

var tests = []struct {
	name  string
	flags map[string]string
}{
	{name: "gocritic"},
	{name: "filtertest"},
	{name: "extra"},
	{name: "suggest"},
	{name: "namedtype/nested"},
	{name: "namedtype"},
	{name: "revive"},
	{name: "golint"},
	{name: "regression"},
	{name: "testvendored"},
	{name: "quasigo"},
	{name: "matching"},
	{name: "dgryski"},
	{name: "comments"},
	{name: "stdlib"},
	{name: "uber"},
	{name: "goversion", flags: map[string]string{"go": "1.16"}},
}

func TestAnalyzer(t *testing.T) {
	analyzer.ForceNewEngine = true
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			testdata := analysistest.TestData()
			rulesFilename := fmt.Sprintf("./testdata/src/%s/rules.go", test.name)
			if err := analyzer.Analyzer.Flags.Set("rules", rulesFilename); err != nil {
				t.Fatalf("set rules flag: %v", err)
			}
			for key, val := range test.flags {
				if err := analyzer.Analyzer.Flags.Set(key, val); err != nil {
					t.Fatalf("set rules flag: %v", err)
				}
			}
			analysistest.Run(t, testdata, analyzer.Analyzer, test.name)
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

	for i := range tests {
		test := tests[i]
		if test.flags != nil {
			continue // Run only trivial tests for now
		}
		t.Run(test.name, func(t *testing.T) {
			rulesFilename := filepath.Join(wd, "testdata", "src", test.name, "rules.go")
			data, err := ioutil.ReadFile(rulesFilename)
			if err != nil {
				t.Fatalf("%s: %v", test.name, err)
			}
			fset := token.NewFileSet()
			f, err := goutil.LoadGoFile(goutil.LoadConfig{
				Fset:     fset,
				Filename: rulesFilename,
				Data:     data,
			})
			if err != nil {
				t.Fatalf("%s: %v", test.name, err)
			}
			ctx := &irconv.Context{
				Pkg:   f.Pkg,
				Types: f.Types,
				Fset:  fset,
				Src:   data,
			}
			irfile, err := irconv.ConvertFile(ctx, f.Syntax)
			if err != nil {
				t.Fatalf("%s: irconv: %v", test.name, err)
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
			srcRulesCmd.Dir = filepath.Join(wd, "testdata", "src", test.name)
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
			irRulesCmd.Dir = filepath.Join(wd, "testdata", "src", test.name)
			irOut, _ := irRulesCmd.CombinedOutput()

			if diff := cmp.Diff(string(irOut), string(srcOut)); diff != "" {
				t.Errorf("%s output mismatches: %s", test.name, diff)
			}

		})
	}
}
