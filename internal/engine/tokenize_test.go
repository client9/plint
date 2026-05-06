package engine

import "testing"

func TestTokenize_Basic(t *testing.T) {
	tokens := Tokenize("Hello, World!")
	want := []Token{
		{Text: "hello", Offset: 0, Len: 5},
		{Text: "world", Offset: 7, Len: 5},
	}
	assertTokens(t, tokens, want)
}

func TestTokenize_Phrase(t *testing.T) {
	tokens := Tokenize("at the end of the day.")
	want := []Token{
		{Text: "at", Offset: 0, Len: 2},
		{Text: "the", Offset: 3, Len: 3},
		{Text: "end", Offset: 7, Len: 3},
		{Text: "of", Offset: 11, Len: 2},
		{Text: "the", Offset: 14, Len: 3},
		{Text: "day", Offset: 18, Len: 3},
	}
	assertTokens(t, tokens, want)
}

func TestTokenize_Multiline(t *testing.T) {
	tokens := Tokenize("foo\nbar baz")
	want := []Token{
		{Text: "foo", Offset: 0, Len: 3},
		{Text: "bar", Offset: 4, Len: 3},
		{Text: "baz", Offset: 8, Len: 3},
	}
	assertTokens(t, tokens, want)
}

func TestTokenize_Empty(t *testing.T) {
	if Tokenize("") != nil {
		t.Fatal("expected nil for empty string")
	}
}

func TestTokenize_Unicode(t *testing.T) {
	// "naïve" — ï is U+00EF, encoded as 2 bytes (0xC3 0xAF) in UTF-8
	tokens := Tokenize("naïve café")
	want := []Token{
		{Text: "naïve", Offset: 0, Len: 6}, // n(1)+a(1)+ï(2)+v(1)+e(1) = 6
		{Text: "café", Offset: 7, Len: 5},  // c(1)+a(1)+f(1)+é(2) = 5
	}
	assertTokens(t, tokens, want)
}

func assertTokens(t *testing.T, got, want []Token) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len %d, want %d: got %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("token[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}
