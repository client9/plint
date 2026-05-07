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
	ID string `json:"id"`
	// Type is "phrase" (default) or "spell". Phrase rules use Tokens; spell
	// rules use Dictionaries, Wordlists, and Words.
	Type         string   `json:"type"`
	Message      string   `json:"message"` // template source, e.g. `"{{.Match}}" may be misspelled`
	Severity     string   `json:"severity"`
	Link         string   `json:"link"`
	Tokens       []string `json:"tokens"`       // phrase rules only
	Dictionaries []string `json:"dictionaries"` // spell rules: base dict first, supplements follow
	Wordlists    []string `json:"wordlists"`    // spell rules: paths to word list files
	Words        []string `json:"words"`        // spell rules: inline allowed words

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
	switch rule.Type {
	case "", "phrase":
		rule.Type = "phrase"
	case "spell":
		if len(rule.Dictionaries) == 0 {
			return nil, fmt.Errorf("%s: spell rule must list at least one dictionary", path)
		}
	default:
		return nil, fmt.Errorf("%s: unknown rule type %q", path, rule.Type)
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
// alternation groups before inserting. Spell rules are skipped; they are
// handled separately via SpellCheck.
func AddToTrie(t *engine.Trie, rule *RuleDef) {
	if rule.Type == "spell" {
		return
	}
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
