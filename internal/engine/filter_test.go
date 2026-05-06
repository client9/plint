package engine

import (
	"testing"

	"github.com/nickg/plint/internal/parser"
)

func TestFilter_Disable(t *testing.T) {
	src := []byte("Make no mistake, this is clearly wrong.\n")
	doc, _ := parser.ParseMarkdown(src, "test.md")

	tr := &Trie{}
	tr.Add([]string{"make", "no", "mistake"}, Rule{Name: "throat-clearing"})
	tr.Add([]string{"clearly"}, Rule{Name: "adverbial-inflation"})

	hits := Lint(doc, tr, DefaultScope)
	if len(hits) != 2 {
		t.Fatalf("pre-filter: got %d hits, want 2", len(hits))
	}

	hits = Filter(hits, parser.DocumentMeta{Disable: []string{"throat-clearing"}}, src)
	if len(hits) != 1 || hits[0].Rule.Name != "adverbial-inflation" {
		t.Errorf("post-filter: got %+v, want only adverbial-inflation", hits)
	}
}

func TestFilter_Allow(t *testing.T) {
	src := []byte("In fact, this is clearly wrong.\n")
	doc, _ := parser.ParseMarkdown(src, "test.md")

	tr := &Trie{}
	tr.Add([]string{"in", "fact"}, Rule{Name: "throat-clearing"})
	tr.Add([]string{"clearly"}, Rule{Name: "adverbial-inflation"})

	hits := Lint(doc, tr, DefaultScope)
	if len(hits) != 2 {
		t.Fatalf("pre-filter: got %d hits, want 2", len(hits))
	}

	hits = Filter(hits, parser.DocumentMeta{Allow: []string{"in fact"}}, src)
	if len(hits) != 1 || hits[0].Rule.Name != "adverbial-inflation" {
		t.Errorf("post-filter: got %+v, want only adverbial-inflation", hits)
	}
}

func TestFilter_Empty(t *testing.T) {
	src := []byte("Clearly wrong.\n")
	doc, _ := parser.ParseMarkdown(src, "test.md")

	tr := &Trie{}
	tr.Add([]string{"clearly"}, Rule{Name: "adverbial-inflation"})

	hits := Lint(doc, tr, DefaultScope)
	hits = Filter(hits, parser.DocumentMeta{}, src)
	if len(hits) != 1 {
		t.Errorf("empty meta should not filter anything, got %d hits", len(hits))
	}
}
