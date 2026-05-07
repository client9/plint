package engine

import (
	"strings"
	"testing"

	"github.com/nickg/plint/internal/parser"
)

// badWordChecker rejects words in its bad set (case-insensitive).
type badWordChecker struct {
	bad map[string]bool
}

func (c *badWordChecker) Spell(word string) bool {
	return !c.bad[strings.ToLower(word)]
}

// recordingChecker records every word it receives and accepts all.
type recordingChecker struct {
	received []string
}

func (c *recordingChecker) Spell(word string) bool {
	c.received = append(c.received, word)
	return true
}

func TestSpellCheck_SingleMiss(t *testing.T) {
	src := []byte("The quikc brown fox.\n")
	doc, _ := parser.ParseMarkdown(src, "test.md")

	checker := &badWordChecker{bad: map[string]bool{"quikc": true}}
	hits := SpellCheck(doc, checker, DefaultScope, Rule{Name: "spelling"})

	if len(hits) != 1 {
		t.Fatalf("got %d hits, want 1", len(hits))
	}
	if hits[0].Rule.Name != "spelling" {
		t.Errorf("rule name = %q, want spelling", hits[0].Rule.Name)
	}
	got := string(src[hits[0].Offset:hits[0].EndOffset])
	if got != "quikc" {
		t.Errorf("matched %q, want %q", got, "quikc")
	}
}

func TestSpellCheck_OriginalCase(t *testing.T) {
	// Checker must receive original-case words so hunspell case rules apply.
	src := []byte("Hello World.\n")
	doc, _ := parser.ParseMarkdown(src, "test.md")

	checker := &recordingChecker{}
	SpellCheck(doc, checker, DefaultScope, Rule{Name: "spelling"})

	want := []string{"Hello", "World"}
	if len(checker.received) != len(want) {
		t.Fatalf("received %v, want %v", checker.received, want)
	}
	for i, w := range want {
		if checker.received[i] != w {
			t.Errorf("word[%d] = %q, want %q", i, checker.received[i], w)
		}
	}
}

func TestSpellCheck_SkipsCodeBlocks(t *testing.T) {
	src := []byte("```\nmisspeled code\n```\n")
	doc, _ := parser.ParseMarkdown(src, "test.md")

	checker := &badWordChecker{bad: map[string]bool{"misspeled": true}}
	hits := SpellCheck(doc, checker, DefaultScope, Rule{Name: "spelling"})
	if len(hits) != 0 {
		t.Errorf("code blocks should be outside DefaultScope, got %d hits", len(hits))
	}
}

func TestSpellCheck_AbsoluteOffsets(t *testing.T) {
	// "wrod" is in the second paragraph; verify offset is absolute into src.
	src := []byte("First paragraph.\n\nBad wrod here.\n")
	doc, _ := parser.ParseMarkdown(src, "test.md")

	checker := &badWordChecker{bad: map[string]bool{"wrod": true}}
	hits := SpellCheck(doc, checker, DefaultScope, Rule{Name: "spelling"})

	if len(hits) != 1 {
		t.Fatalf("got %d hits, want 1", len(hits))
	}
	got := string(src[hits[0].Offset:hits[0].EndOffset])
	if got != "wrod" {
		t.Errorf("src[%d:%d] = %q, want %q", hits[0].Offset, hits[0].EndOffset, got, "wrod")
	}
}

func TestSpellCheck_MultipleMisses(t *testing.T) {
	src := []byte("Ths iz rong.\n")
	doc, _ := parser.ParseMarkdown(src, "test.md")

	checker := &badWordChecker{bad: map[string]bool{"ths": true, "iz": true, "rong": true}}
	hits := SpellCheck(doc, checker, DefaultScope, Rule{Name: "spelling"})
	if len(hits) != 3 {
		t.Errorf("got %d hits, want 3", len(hits))
	}
}

func TestSpellCheck_NoHitsWhenAllCorrect(t *testing.T) {
	src := []byte("Everything looks fine here.\n")
	doc, _ := parser.ParseMarkdown(src, "test.md")

	checker := &badWordChecker{bad: map[string]bool{}}
	hits := SpellCheck(doc, checker, DefaultScope, Rule{Name: "spelling"})
	if len(hits) != 0 {
		t.Errorf("got %d hits, want 0", len(hits))
	}
}
