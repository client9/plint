package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/nickg/plint"
)

func main() {
	var rulesPath string
	flag.StringVar(&rulesPath, "rules", "", "path to a rules directory or a single .yaml rule file")
	flag.StringVar(&rulesPath, "config", "", "alias for -rules (vale-ls compatibility)")
	outputFlag := flag.String("output", "line", `output format: "line", "json", or a path to a Go template file`)
	quiet := flag.Bool("q", false, "suppress output; exit 0 (clean), 1 (findings), 2 (error)")
	ver := flag.Bool("v", false, "print version and exit")
	flag.String("filter", "", "JMESPath alert filter (vale-ls compatibility, unused)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: plint [flags] [file.md ...]\n")
		fmt.Fprintf(os.Stderr, "       cat file.md | plint [flags]\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	// Subcommand: fix <alert-file> — used by vale-ls for quick-fix code actions.
	if flag.NArg() >= 2 && flag.Arg(0) == "fix" {
		cmdFix(flag.Arg(1))
	}

	if *ver {
		v := "(devel)"
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
			v = info.Main.Version
		}
		fmt.Println("plint", v)
		os.Exit(0)
	}

	if rulesPath == "" {
		found, err := findPlintDir()
		if err != nil {
			fatal(fmt.Errorf("no rules path specified and no .plint directory found"))
		}
		rulesPath = found
	}

	linter, err := plint.New(rulesPath)
	if err != nil {
		fatal(err)
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
	allFindings := make(map[string][]plint.Finding)
	for _, src := range sources {
		findings, warnings, err := linter.LintReader(src.r, src.name)
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

	if !*quiet {
		switch *outputFlag {
		case "json", "JSON":
			err = plint.WriteJSON(os.Stdout, allFindings)
		case "line":
			err = plint.WriteLine(os.Stdout, allFindings)
		default:
			err = plint.WriteTemplate(os.Stdout, *outputFlag, allFindings)
		}
		if err != nil {
			fatal(err)
		}
	}

	if len(allFindings) == 0 {
		os.Exit(0)
	}
	os.Exit(1)
}

// cmdFix reads a ValeAlert JSON file and writes a ValeFix JSON response to
// stdout. Used by vale-ls to populate quick-fix code actions.
func cmdFix(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fatal(err)
	}
	var finding plint.Finding
	if err := json.Unmarshal(data, &finding); err != nil {
		fatal(err)
	}
	type valeFix struct {
		Suggestions []string `json:"suggestions"`
		Error       string   `json:"error"`
	}
	fix := valeFix{Suggestions: finding.Action.Params, Error: ""}
	if fix.Suggestions == nil {
		fix.Suggestions = []string{}
	}
	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(fix); err != nil {
		fatal(err)
	}
	os.Exit(0)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "plint: %v\n", err)
	os.Exit(2)
}

// findPlintDir walks up from cwd looking for a .plint directory.
func findPlintDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, ".plint")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not found")
		}
		dir = parent
	}
}
