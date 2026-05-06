package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nickg/plint/internal/engine"
	"github.com/nickg/plint/internal/output"
	"github.com/nickg/plint/internal/parser"
	"github.com/nickg/plint/internal/rules"
)

func main() {
	rulesFlag := flag.String("rules", "rules", "path to a rules directory or a single .yaml rule file")
	outputFlag := flag.String("output", "line", `output format: "line", "json", or a path to a Go template file`)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: plint [flags] <file.md>\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	mdPath := flag.Arg(0)

	// Load rules.
	var defs []*rules.RuleDef
	info, err := os.Stat(*rulesFlag)
	if err != nil {
		fatal(err)
	}
	var loadErr error
	if info.IsDir() {
		defs, loadErr = rules.LoadDir(*rulesFlag)
	} else {
		var r *rules.RuleDef
		r, loadErr = rules.Load(*rulesFlag)
		defs = []*rules.RuleDef{r}
	}
	if loadErr != nil {
		fatal(loadErr)
	}

	defsByID := make(map[string]*rules.RuleDef, len(defs))
	for _, r := range defs {
		defsByID[r.ID] = r
	}

	// Build trie.
	trie := &engine.Trie{}
	for _, r := range defs {
		rules.AddToTrie(trie, r)
	}

	// Parse markdown.
	src, err := os.ReadFile(mdPath)
	if err != nil {
		fatal(err)
	}
	doc, err := parser.ParseMarkdown(src, mdPath)
	if err != nil {
		fatal(err)
	}

	// Lint.
	hits := engine.Lint(doc, trie, engine.DefaultScope)
	if len(hits) == 0 {
		return
	}

	lm := engine.NewLineMap(src)
	findings := make([]output.Finding, len(hits))
	for i, h := range hits {
		findings[i] = output.Build(h, src, lm, defsByID)
	}

	// Output.
	switch *outputFlag {
	case "json":
		err = output.WriteJSON(os.Stdout, mdPath, findings)
	case "line":
		err = output.WriteLine(os.Stdout, mdPath, findings)
	default:
		err = output.WriteTemplate(os.Stdout, *outputFlag, mdPath, findings)
	}
	if err != nil {
		fatal(err)
	}
	os.Exit(1)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "plint: %v\n", err)
	os.Exit(2)
}
