package plint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gospell "github.com/client9/gospell"

	"github.com/nickg/plint/internal/engine"
	"github.com/nickg/plint/internal/parser"
	"github.com/nickg/plint/internal/rules"
)

// spellRuleState holds the pre-loaded dictionary state for one spell rule.
// A new gospell.Checker is built per-document so per-document word lists
// can be added without mutating shared state.
type spellRuleState struct {
	rule            engine.Rule
	def             *rules.RuleDef
	base            *gospell.GoSpell
	globalWordLists []*gospell.WordList
}

// buildChecker creates a fresh Checker for a single document, layering
// the global word lists and any per-document words on top of the base dict.
// The returned checker also implements engine.Suggester when the base
// dictionary has a suggester configured.
func (s *spellRuleState) buildChecker(meta parser.SpellingMeta) engine.SpellChecker {
	c := gospell.NewChecker(s.base)
	for _, wl := range s.globalWordLists {
		c.AddWordList(wl)
	}
	if len(meta.Words) > 0 {
		docWL := &gospell.WordList{}
		for _, w := range meta.Words {
			docWL.Add(w)
		}
		c.AddWordList(docWL)
	}
	return &spellCheckerWithSuggestions{c}
}

// spellCheckerWithSuggestions wraps a gospell.Checker and implements
// engine.Suggester by converting gospell.Suggestion to plain strings.
// This keeps the engine package free of gospell types.
type spellCheckerWithSuggestions struct {
	*gospell.Checker
}

func (s *spellCheckerWithSuggestions) Suggest(word string, limit int) []string {
	suggestions, err := s.Checker.Suggest(word, limit)
	if err != nil || len(suggestions) == 0 {
		return nil
	}
	words := make([]string, len(suggestions))
	for i, sg := range suggestions {
		words[i] = sg.Word
	}
	return words
}

// buildSpellState loads the dictionary and word list files declared in a spell
// rule and returns the pre-loaded state shared across all documents.
func buildSpellState(def *rules.RuleDef, ruleDir string) (*spellRuleState, error) {
	paths := plintSearchPaths()

	// Load base dictionary (first entry; must have .aff + .dic).
	base, err := gospell.Open(resolveDict(def.Dictionaries[0], ruleDir), paths)
	if err != nil {
		return nil, fmt.Errorf("spell rule %q: %w", def.ID, err)
	}
	if err := base.SetSuggester(gospell.NewLevenshteinSuggester(gospell.LevenshteinOptions{})); err != nil {
		return nil, fmt.Errorf("spell rule %q: configure suggester: %w", def.ID, err)
	}

	var globalWordLists []*gospell.WordList

	// Supplemental dictionaries (index 1..n): try full base dict first, then
	// .dic-only supplement.
	for _, name := range def.Dictionaries[1:] {
		resolved := resolveDict(name, ruleDir)
		if extra, err := gospell.Open(resolved, paths); err == nil {
			wl := &gospell.WordList{}
			extra.ForEachWord(func(w string) bool {
				wl.Add(w)
				return true
			})
			globalWordLists = append(globalWordLists, wl)
		} else if wl, err := gospell.OpenSupplement(resolved, paths); err == nil {
			globalWordLists = append(globalWordLists, wl)
		} else {
			return nil, fmt.Errorf("spell rule %q: dictionary %q: not found as base or supplement", def.ID, name)
		}
	}

	// Word list files.
	for _, wlPath := range def.Wordlists {
		resolved := resolveFilePath(wlPath, ruleDir)
		wl, err := gospell.NewWordListFile(resolved)
		if err != nil {
			return nil, fmt.Errorf("spell rule %q: wordlist %q: %w", def.ID, wlPath, err)
		}
		globalWordLists = append(globalWordLists, wl)
	}

	// Inline words.
	if len(def.Words) > 0 {
		inlineWL := &gospell.WordList{}
		for _, w := range def.Words {
			inlineWL.Add(w)
		}
		globalWordLists = append(globalWordLists, inlineWL)
	}

	return &spellRuleState{
		rule:            engine.Rule{Name: def.ID},
		def:             def,
		base:            base,
		globalWordLists: globalWordLists,
	}, nil
}

// filterSpellingIgnore removes spell hits whose matched word appears in the
// per-document spelling.ignore list (case-insensitive).
func filterSpellingIgnore(hits []engine.Hit, ignore []string, src []byte) []engine.Hit {
	if len(ignore) == 0 {
		return hits
	}
	set := make(map[string]bool, len(ignore))
	for _, w := range ignore {
		set[strings.ToLower(w)] = true
	}
	out := hits[:0]
	for _, h := range hits {
		if !set[strings.ToLower(string(src[h.Offset:h.EndOffset]))] {
			out = append(out, h)
		}
	}
	return out
}

// plintSearchPaths returns the ordered dictionary search path:
// PLINT_DICT_PATH (plint-specific) first, then gospell.SearchPaths()
// (DICPATH + system defaults).
func plintSearchPaths() []string {
	var paths []string
	if p := os.Getenv("PLINT_DICT_PATH"); p != "" {
		paths = append(paths, filepath.SplitList(p)...)
	}
	return append(paths, gospell.SearchPaths()...)
}

// resolveDict resolves a dictionary name relative to ruleDir when the name
// starts with "./" or "../". Bare names and absolute paths are returned
// unchanged for gospell.Open / gospell.OpenSupplement to handle.
func resolveDict(name, ruleDir string) string {
	if strings.HasPrefix(name, "./") || strings.HasPrefix(name, "../") {
		return filepath.Join(ruleDir, name)
	}
	return name
}

// resolveFilePath resolves a word list file path relative to ruleDir when it
// is not absolute.
func resolveFilePath(path, ruleDir string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(ruleDir, path)
}
