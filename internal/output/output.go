package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"text/template"

	"github.com/nickg/plint/internal/engine"
	"github.com/nickg/plint/internal/rules"
)

// Finding is the per-match output object, shaped after Vale's JSON format.
type Finding struct {
	Line     int    `json:"Line"`
	Span     [2]int `json:"Span"` // [col_start, col_end], 1-based
	Check    string `json:"Check"`
	Message  string `json:"Message"`
	Severity string `json:"Severity"`
	Match    string `json:"Match"`
	Link     string `json:"Link"`
}

// Build converts a Hit into a Finding using the source bytes, line map, and rule definitions.
func Build(h engine.Hit, src []byte, lm engine.LineMap, defs map[string]*rules.RuleDef) Finding {
	line, col := lm.Position(h.Offset)
	_, endCol := lm.Position(h.EndOffset - 1)

	match := string(src[h.Offset:h.EndOffset])

	f := Finding{
		Line:  line,
		Span:  [2]int{col, endCol},
		Check: h.Rule.Name,
		Match: match,
	}

	if def, ok := defs[h.Rule.Name]; ok {
		f.Message = fmt.Sprintf(def.Message, match)
		f.Severity = def.Severity
		f.Link = def.Link
	}

	return f
}

// WriteJSON writes all findings as a Vale-compatible JSON object keyed by filename.
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
			_, err := fmt.Fprintf(w, "%s:%d:%d:%s:%s\n", file, f.Line, f.Span[0], f.Check, f.Message)
			if err != nil {
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
