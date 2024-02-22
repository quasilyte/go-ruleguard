package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/quasilyte/go-ruleguard/analyzer"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
