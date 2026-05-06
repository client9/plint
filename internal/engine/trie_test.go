package engine

import "testing"

// toks builds a token slice from words, using index as a synthetic offset.
func toks(words ...string) []Token {
	t := make([]Token, len(words))
	for i, w := range words {
		t[i] = Token{Text: w, Offset: i}
	}
	return t
}

func TestTrie_WordMatch(t *testing.T) {
	tr := &Trie{}
	tr.Add([]string{"deeply"}, Rule{Name: "adverbial-inflation"})

	rule, n, ok := tr.Match(toks("deeply"), 0)
	if !ok || rule.Name != "adverbial-inflation" || n != 1 {
		t.Fatalf("got (%v, %d, %v)", rule, n, ok)
	}
}

func TestTrie_PhraseMatch(t *testing.T) {
	tr := &Trie{}
	tr.Add([]string{"at", "the", "end", "of", "the", "day"}, Rule{Name: "throat-clearing"})

	tokens := toks("so", "at", "the", "end", "of", "the", "day", "it", "matters")
	rule, n, ok := tr.Match(tokens, 1)
	if !ok || rule.Name != "throat-clearing" || n != 6 {
		t.Fatalf("got (%v, %d, %v)", rule, n, ok)
	}
}

func TestTrie_NoMatch(t *testing.T) {
	tr := &Trie{}
	tr.Add([]string{"clearly"}, Rule{Name: "adverbial-inflation"})

	_, _, ok := tr.Match(toks("obviously"), 0)
	if ok {
		t.Fatal("expected no match")
	}
}

func TestTrie_PrefixNotMatch(t *testing.T) {
	tr := &Trie{}
	tr.Add([]string{"at", "the", "end"}, Rule{Name: "r1"})

	_, _, ok := tr.Match(toks("at", "the"), 0)
	if ok {
		t.Fatal("prefix should not match")
	}
}

func TestTrie_LongestMatch(t *testing.T) {
	tr := &Trie{}
	tr.Add([]string{"in", "fact"}, Rule{Name: "short"})
	tr.Add([]string{"in", "fact", "wrong"}, Rule{Name: "long"})

	rule, n, ok := tr.Match(toks("in", "fact", "wrong"), 0)
	if !ok || rule.Name != "long" || n != 3 {
		t.Fatalf("got (%v, %d, %v)", rule, n, ok)
	}
}

func TestTrie_LongestMatchFallsBackToShorter(t *testing.T) {
	tr := &Trie{}
	tr.Add([]string{"in", "fact"}, Rule{Name: "short"})
	tr.Add([]string{"in", "fact", "wrong"}, Rule{Name: "long"})

	rule, n, ok := tr.Match(toks("in", "fact", "right"), 0)
	if !ok || rule.Name != "short" || n != 2 {
		t.Fatalf("got (%v, %d, %v)", rule, n, ok)
	}
}

func TestTrie_MultipleRules(t *testing.T) {
	tr := &Trie{}
	tr.Add([]string{"truly"}, Rule{Name: "r1"})
	tr.Add([]string{"make", "no", "mistake"}, Rule{Name: "r2"})

	r, n, ok := tr.Match(toks("make", "no", "mistake"), 0)
	if !ok || r.Name != "r2" || n != 3 {
		t.Fatalf("got (%v, %d, %v)", r, n, ok)
	}
}

func TestTrie_Scan(t *testing.T) {
	tr := &Trie{}
	tr.Add([]string{"deeply"}, Rule{Name: "adverbial-inflation"})
	tr.Add([]string{"make", "no", "mistake"}, Rule{Name: "throat-clearing"})

	// "make no mistake" starts at offset 0, "deeply" starts at offset 16
	tokens := Tokenize("make no mistake this is deeply wrong")
	hits := tr.Scan(tokens)

	if len(hits) != 2 {
		t.Fatalf("got %d hits, want 2: %+v", len(hits), hits)
	}
	if hits[0].Rule.Name != "throat-clearing" || hits[0].Len != 3 {
		t.Errorf("hit[0] = %+v", hits[0])
	}
	if hits[1].Rule.Name != "adverbial-inflation" || hits[1].Len != 1 {
		t.Errorf("hit[1] = %+v", hits[1])
	}
}

func TestTrie_ScanOffset(t *testing.T) {
	tr := &Trie{}
	tr.Add([]string{"clearly"}, Rule{Name: "r1"})

	// "clearly" starts at byte offset 8 in "this is clearly wrong"
	tokens := Tokenize("this is clearly wrong")
	hits := tr.Scan(tokens)

	if len(hits) != 1 {
		t.Fatalf("got %d hits, want 1", len(hits))
	}
	if hits[0].Offset != 8 {
		t.Errorf("got offset %d, want 8", hits[0].Offset)
	}
}
