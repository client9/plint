package rules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/client9/tojson"
	"github.com/nickg/plint/internal/engine"
)

// RuleDef is a parsed rule file.
type RuleDef struct {
	ID       string   `json:"id"`
	Message  string   `json:"message"` // template source, e.g. `"{{.Match}}" is wordy`
	Severity string   `json:"severity"`
	Link     string   `json:"link"`
	Tokens   []string `json:"tokens"`

	msgTmpl *template.Template // compiled at load time
}

// FormatMessage executes the message template with the matched text.
// Falls back to the raw Message string if execution fails.
func (r *RuleDef) FormatMessage(match string) string {
	if r.msgTmpl == nil {
		return r.Message
	}
	var buf bytes.Buffer
	if err := r.msgTmpl.Execute(&buf, struct{ Match string }{match}); err != nil {
		return r.Message
	}
	return buf.String()
}

// Load reads a single YAML rule file and returns the parsed RuleDef.
// Fails fast if the message template is invalid.
func Load(path string) (*RuleDef, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	jsonBytes, err := tojson.FromYAML(src)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	var rule RuleDef
	if err := json.Unmarshal(jsonBytes, &rule); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if rule.ID == "" {
		return nil, fmt.Errorf("%s: rule missing id", path)
	}
	tmpl, err := template.New(rule.ID).Option("missingkey=error").Parse(rule.Message)
	if err != nil {
		return nil, fmt.Errorf("%s: message template: %w", path, err)
	}
	rule.msgTmpl = tmpl
	return &rule, nil
}

// LoadDir reads all *.yaml files in dir and returns the parsed rules.
func LoadDir(dir string) ([]*RuleDef, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	var rules []*RuleDef
	for _, p := range paths {
		r, err := Load(p)
		if err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, nil
}

// AddToTrie adds all token patterns from rule into t, expanding any (a|b)
// alternation groups before inserting.
func AddToTrie(t *engine.Trie, rule *RuleDef) {
	for _, phrase := range rule.Tokens {
		for _, expanded := range Expand(phrase) {
			tokens := engine.Tokenize(expanded)
			words := make([]string, len(tokens))
			for i, tok := range tokens {
				words[i] = tok.Text
			}
			t.Add(words, engine.Rule{Name: rule.ID})
		}
	}
}
