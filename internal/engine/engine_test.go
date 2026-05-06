package engine

import (
	"testing"

	"github.com/nickg/plint/internal/parser"
)

func TestLint_RoundTrip(t *testing.T) {
	src := []byte("This is clearly wrong.\n\nMake no mistake, it is deeply flawed.\n")

	doc, err := parser.ParseMarkdown(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}

	tr := &Trie{}
	tr.Add([]string{"clearly"}, Rule{Name: "adverbial-inflation"})
	tr.Add([]string{"deeply"}, Rule{Name: "adverbial-inflation"})
	tr.Add([]string{"make", "no", "mistake"}, Rule{Name: "throat-clearing"})

	hits := Lint(doc, tr, DefaultScope)

	if len(hits) != 3 {
		t.Fatalf("got %d hits, want 3: %+v", len(hits), hits)
	}

	// "clearly" is at offset 8 in the source
	if hits[0].Rule.Name != "adverbial-inflation" || hits[0].Offset != 8 {
		t.Errorf("hit[0] = %+v, want adverbial-inflation at offset 8", hits[0])
	}

	// "Make no mistake" starts at offset 24 (start of second paragraph)
	if hits[1].Rule.Name != "throat-clearing" || hits[1].Offset != 24 {
		t.Errorf("hit[1] = %+v, want throat-clearing at offset 24", hits[1])
	}

	// "deeply" starts at offset 47
	if hits[2].Rule.Name != "adverbial-inflation" || hits[2].Offset != 47 {
		t.Errorf("hit[2] = %+v, want adverbial-inflation at offset 47", hits[2])
	}
}

func TestLint_FrontMatterIgnored(t *testing.T) {
	src := []byte("---\ntitle: Test\n---\n\nClearly this matters.\n")

	doc, err := parser.ParseMarkdown(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}

	tr := &Trie{}
	tr.Add([]string{"clearly"}, Rule{Name: "adverbial-inflation"})

	hits := Lint(doc, tr, DefaultScope)

	if len(hits) != 1 {
		t.Fatalf("got %d hits, want 1: %+v", len(hits), hits)
	}
}

func TestLint_NoMatches(t *testing.T) {
	src := []byte("A clean sentence with no flagged words.\n")

	doc, err := parser.ParseMarkdown(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}

	tr := &Trie{}
	tr.Add([]string{"clearly"}, Rule{Name: "adverbial-inflation"})

	hits := Lint(doc, tr, DefaultScope)

	if len(hits) != 0 {
		t.Fatalf("got %d hits, want 0: %+v", len(hits), hits)
	}
}
