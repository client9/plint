package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/nickg/plint"
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

	linter, err := plint.New(*rulesFlag)
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

	if len(allFindings) == 0 {
		os.Exit(0)
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

	os.Exit(1)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "plint: %v\n", err)
	os.Exit(2)
}
