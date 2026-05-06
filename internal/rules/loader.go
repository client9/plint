package rules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/client9/tojson"
	"github.com/nickg/plint/internal/engine"
)

// RuleDef is a parsed rule file.
type RuleDef struct {
	ID       string   `json:"id"`
	Message  string   `json:"message"`
	Severity string   `json:"severity"`
	Link     string   `json:"link"`
	Tokens   []string `json:"tokens"`
}

// Load reads a single YAML rule file and returns the parsed RuleDef.
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

// AddToTrie adds all token patterns from rule into t.
func AddToTrie(t *engine.Trie, rule *RuleDef) {
	for _, phrase := range rule.Tokens {
		tokens := engine.Tokenize(phrase)
		words := make([]string, len(tokens))
		for i, tok := range tokens {
			words[i] = tok.Text
		}
		t.Add(words, engine.Rule{Name: rule.ID})
	}
}
