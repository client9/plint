package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/nickg/plint/internal/engine"
	"github.com/nickg/plint/internal/output"
	"github.com/nickg/plint/internal/parser"
	"github.com/nickg/plint/internal/rules"
)

func main() {
	rulesFlag := flag.String("rules", "rules", "path to a rules directory or a single .yaml rule file")
	outputFlag := flag.String("output", "line", `output format: "line", "json", or a path to a Go template file`)
	quiet := flag.Bool("q", false, "suppress output; exit 0 (clean), 1 (findings), 2 (error)")
	ver := flag.Bool("v", false, "print version and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: plint [flags] [file.md ...]\n")
		fmt.Fprintf(os.Stderr, "       cat file.md | plint [flags]\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *ver {
		v := "(devel)"
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
			v = info.Main.Version
		}
		fmt.Println("plint", v)
		os.Exit(0)
	}

	// Load rules.
	defs, err := loadRules(*rulesFlag)
	if err != nil {
		fatal(err)
	}
	defsByID := make(map[string]*rules.RuleDef, len(defs))
	knownIDs := make(map[string]bool, len(defs))
	for _, r := range defs {
		defsByID[r.ID] = r
		knownIDs[r.ID] = true
	}

	// Build trie.
	trie := &engine.Trie{}
	for _, r := range defs {
		rules.AddToTrie(trie, r)
	}

	// Collect input sources: files from args, or stdin if no args given.
	type source struct {
		name string
		r    io.Reader
	}
	var sources []source
	if flag.NArg() == 0 {
		sources = []source{{name: "<stdin>", r: os.Stdin}}
	} else {
		for _, path := range flag.Args() {
			if path == "-" {
				sources = append(sources, source{name: "<stdin>", r: os.Stdin})
			} else {
				f, err := os.Open(path)
				if err != nil {
					fatal(err)
				}
				defer f.Close()
				sources = append(sources, source{name: path, r: f})
			}
		}
	}

	// Lint each source and collect findings.
	allFindings := make(map[string][]output.Finding)
	for _, src := range sources {
		findings, warnings, err := lintSource(src.name, src.r, trie, defsByID, knownIDs)
		if err != nil {
			fatal(err)
		}
		for _, w := range warnings {
			fmt.Fprintf(os.Stderr, "plint: %s: %s\n", src.name, w)
		}
		if len(warnings) > 0 {
			os.Exit(2)
		}
		if len(findings) > 0 {
			allFindings[src.name] = findings
		}
	}

	if len(allFindings) == 0 {
		os.Exit(0)
	}

	if !*quiet {
		switch *outputFlag {
		case "json":
			err = output.WriteJSON(os.Stdout, allFindings)
		case "line":
			err = output.WriteLine(os.Stdout, allFindings)
		default:
			err = output.WriteTemplate(os.Stdout, *outputFlag, allFindings)
		}
		if err != nil {
			fatal(err)
		}
	}

	os.Exit(1)
}

func lintSource(name string, r io.Reader, trie *engine.Trie, defsByID map[string]*rules.RuleDef, knownIDs map[string]bool) ([]output.Finding, []string, error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", name, err)
	}
	doc, err := parser.ParseMarkdown(src, name)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", name, err)
	}

	warnings := engine.ValidateMeta(doc.Meta, knownIDs)

	hits := engine.Lint(doc, trie, engine.DefaultScope)
	hits = engine.Filter(hits, doc.Meta, src)

	lm := engine.NewLineMap(src)
	findings := make([]output.Finding, len(hits))
	for i, h := range hits {
		findings[i] = output.Build(h, src, lm, defsByID)
	}
	return findings, warnings, nil
}

func loadRules(path string) ([]*rules.RuleDef, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return rules.LoadDir(path)
	}
	r, err := rules.Load(path)
	if err != nil {
		return nil, err
	}
	return []*rules.RuleDef{r}, nil
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "plint: %v\n", err)
	os.Exit(2)
}
