package engine

import (
	"testing"

	"github.com/nickg/plint/internal/parser"
)

func TestValidateMeta(t *testing.T) {
	known := map[string]bool{"adverbial-inflation": true, "throat-clearing": true}

	t.Run("no warnings for known rules", func(t *testing.T) {
		meta := parser.DocumentMeta{Rules: parser.RulesMeta{Disable: []string{"adverbial-inflation"}}}
		if w := ValidateMeta(meta, known); len(w) != 0 {
			t.Errorf("expected no warnings, got %v", w)
		}
	})

	t.Run("warning for unknown rule", func(t *testing.T) {
		meta := parser.DocumentMeta{Rules: parser.RulesMeta{Disable: []string{"throte-clearing"}}}
		w := ValidateMeta(meta, known)
		if len(w) != 1 {
			t.Fatalf("expected 1 warning, got %v", w)
		}
		if w[0] != `rules.disable: unknown rule "throte-clearing"` {
			t.Errorf("unexpected warning: %s", w[0])
		}
	})

	t.Run("empty meta produces no warnings", func(t *testing.T) {
		if w := ValidateMeta(parser.DocumentMeta{}, known); len(w) != 0 {
			t.Errorf("expected no warnings, got %v", w)
		}
	})
}
