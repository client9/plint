package plint

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"text/template"
)

// WriteJSON writes findings as a Vale-compatible JSON object keyed by filename.
// The template data shape is map[string][]Finding.
func WriteJSON(w io.Writer, findings map[string][]Finding) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(findings)
}

// WriteLine writes one line per finding in file:line:col:rule:message format.
// Files are output in sorted order for deterministic output.
func WriteLine(w io.Writer, findings map[string][]Finding) error {
	files := make([]string, 0, len(findings))
	for f := range findings {
		files = append(files, f)
	}
	sort.Strings(files)

	for _, file := range files {
		for _, f := range findings[file] {
			if _, err := fmt.Fprintf(w, "%s:%d:%d:%s:%s\n", file, f.Line, f.Span[0], f.Check, f.Message); err != nil {
				return err
			}
		}
	}
	return nil
}

// WriteTemplate loads a Go text/template from templatePath and executes it
// with map[string][]Finding as the data.
func WriteTemplate(w io.Writer, templatePath string, findings map[string][]Finding) error {
	src, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}
	tmpl, err := template.New("output").Parse(string(src))
	if err != nil {
		return err
	}
	return tmpl.Execute(w, findings)
}
