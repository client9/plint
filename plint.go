// Package plint is a prose linting library for Markdown documents.
package plint

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nickg/plint/internal/engine"
	"github.com/nickg/plint/internal/parser"
	"github.com/nickg/plint/internal/rules"
)

// Finding is a single lint result.
type Finding struct {
	File     string `json:"File"`
	Line     int    `json:"Line"`
	Span     [2]int `json:"Span"` // [col_start, col_end], 1-based
	Check    string `json:"Check"`
	Message  string `json:"Message"`
	Severity string `json:"Severity"`
	Match    string `json:"Match"`
	Link     string `json:"Link"`
}

// Linter holds loaded rules and applies them to documents.
type Linter struct {
	trie       *engine.Trie
	defs       map[string]*rules.RuleDef
	knownIDs   map[string]bool
	spellRules []spellRuleState
}

// New creates a Linter from rulesPath, which may be a directory of *.yaml
// files or a single .yaml file.
func New(rulesPath string) (*Linter, error) {
	defs, err := loadRuleDefs(rulesPath)
	if err != nil {
		return nil, err
	}

	// Determine the directory that contains the rule files so that relative
	// paths in spell rules (wordlists, dictionaries) resolve correctly.
	rd := rulesPath
	if fi, statErr := os.Stat(rulesPath); statErr == nil && !fi.IsDir() {
		rd = filepath.Dir(rulesPath)
	}

	trie := &engine.Trie{}
	defsByID := make(map[string]*rules.RuleDef, len(defs))
	knownIDs := make(map[string]bool, len(defs))
	var spellRules []spellRuleState

	for _, r := range defs {
		defsByID[r.ID] = r
		knownIDs[r.ID] = true
		if r.Type == "spell" {
			state, err := buildSpellState(r, rd)
			if err != nil {
				return nil, err
			}
			spellRules = append(spellRules, *state)
		} else {
			rules.AddToTrie(trie, r)
		}
	}

	return &Linter{
		trie:       trie,
		defs:       defsByID,
		knownIDs:   knownIDs,
		spellRules: spellRules,
	}, nil
}

// Lint parses src as Markdown and returns findings and any configuration
// warnings (e.g. unknown rule names in front matter disable lists).
func (l *Linter) Lint(src []byte, filename string) ([]Finding, []string, error) {
	doc, err := parser.ParseMarkdown(src, filename)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", filename, err)
	}

	warnings := engine.ValidateMeta(doc.Meta, l.knownIDs)

	hits := engine.Lint(doc, l.trie, engine.DefaultScope)

	for i := range l.spellRules {
		sr := &l.spellRules[i]
		checker := sr.buildChecker(doc.Meta.Spelling)
		spellHits := engine.SpellCheck(doc, checker, engine.DefaultScope, sr.rule)
		spellHits = filterSpellingIgnore(spellHits, doc.Meta.Spelling.Ignore, src)
		hits = append(hits, spellHits...)
	}

	hits = engine.Filter(hits, doc.Meta, src)

	lm := engine.NewLineMap(src)
	findings := make([]Finding, len(hits))
	for i, h := range hits {
		findings[i] = buildFinding(h, src, lm, filename, l.defs)
	}
	return findings, warnings, nil
}

// LintReader reads from r and lints the content as Markdown.
func (l *Linter) LintReader(r io.Reader, filename string) ([]Finding, []string, error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", filename, err)
	}
	return l.Lint(src, filename)
}

func buildFinding(h engine.Hit, src []byte, lm engine.LineMap, filename string, defs map[string]*rules.RuleDef) Finding {
	line, col := lm.Position(h.Offset)
	_, endCol := lm.Position(h.EndOffset - 1)
	match := string(src[h.Offset:h.EndOffset])

	f := Finding{
		File:  filename,
		Line:  line,
		Span:  [2]int{col, endCol},
		Check: h.Rule.Name,
		Match: match,
	}
	if def, ok := defs[h.Rule.Name]; ok {
		f.Message = def.FormatMessage(match)
		f.Severity = def.Severity
		f.Link = def.Link
	}
	return f
}

func loadRuleDefs(path string) ([]*rules.RuleDef, error) {
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
