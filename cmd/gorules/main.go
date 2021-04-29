package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/token"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/cespare/subcmd"
	"github.com/quasilyte/go-ruleguard/ruleguard"
)

func main() {
	cmds := []subcmd.Command{
		{
			Name:        "doc",
			Description: "query rules documentation",
			Do:          docMain,
		},
	}

	subcmd.Run(cmds)
}

func docMain(args []string) {
	if err := docCommand(args); err != nil {
		log.Fatal(err)
	}
}

func docCommand(args []string) error {
	type JsonListEntry struct {
		Name       string
		Filename   string
		Line       int
		DocSummary string
		DocBefore  string
		DocAfter   string
		DocTags    []string
	}
	type JsonList struct {
		List []JsonListEntry
	}

	fs := flag.NewFlagSet("gorules doc", flag.ExitOnError)
	flagRules := fs.String("rules", "", `comma-separated list of ruleguard file paths`)
	flagJson := fs.Bool("json", false, `format the output as JSON`)
	fs.Parse(args)

	var groupName string
	extraArgs := fs.Args()
	if len(extraArgs) != 0 {
		groupName = extraArgs[0]
	}

	e := ruleguard.NewEngine()
	fset := token.NewFileSet()
	ctx := &ruleguard.ParseContext{
		Fset: fset,
	}
	filenames := strings.Split(*flagRules, ",")
	for _, filename := range filenames {
		filename = strings.TrimSpace(filename)
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("read rules file: %v", err)
		}
		if err := e.Load(ctx, filename, bytes.NewReader(data)); err != nil {
			return fmt.Errorf("parse rules file: %v", err)
		}
	}

	if *flagJson {
		var result JsonList
		for _, g := range e.LoadedGroups() {
			result.List = append(result.List, JsonListEntry{
				Name:       g.Name,
				Line:       g.Line,
				Filename:   g.Filename,
				DocSummary: g.DocSummary,
				DocBefore:  g.DocBefore,
				DocAfter:   g.DocAfter,
				DocTags:    g.DocTags,
			})
		}
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(result); err != nil {
			return err
		}
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, buf.Bytes(), "", "  "); err != nil {
			return err
		}
		fmt.Print(pretty.String())
		return nil
	}
	if groupName != "" {
		var g *ruleguard.GoRuleGroup
		groups := e.LoadedGroups()
		for i := range groups {
			if groups[i].Name == groupName {
				g = &groups[i]
				break
			}
		}
		if g == nil {
			return fmt.Errorf("the requested %s group was not loaded", groupName)
		}
		fmt.Printf("%s:%d: %s\n\n", filepath.Base(g.Filename), g.Line, g.Name)
		fmt.Printf("Full path: %s\n\n", g.Filename)
		if g.DocSummary != "" {
			fmt.Printf("Summary: %s\n\n", g.DocSummary)
		}
		if len(g.DocTags) != 0 {
			fmt.Printf("Tags: %v\n\n", g.DocTags)
		}
		if g.DocBefore != "" && g.DocAfter != "" {
			fmt.Printf("Before:\n")
			fmt.Printf("\t%s\n", g.DocBefore)
			fmt.Printf("After:\n")
			fmt.Printf("\t%s\n", g.DocAfter)
		}
	} else {
		for _, g := range e.LoadedGroups() {
			fmt.Printf("%s:%d: %s\n", filepath.Base(g.Filename), g.Line, g.Name)
		}
	}

	return nil
}
