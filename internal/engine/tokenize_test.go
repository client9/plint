package engine

import "testing"

func TestTokenize_Basic(t *testing.T) {
	tokens := Tokenize("Hello, World!")
	want := []Token{
		{Text: "hello", Offset: 0},
		{Text: "world", Offset: 7},
	}
	assertTokens(t, tokens, want)
}

func TestTokenize_Phrase(t *testing.T) {
	tokens := Tokenize("at the end of the day.")
	want := []Token{
		{Text: "at", Offset: 0},
		{Text: "the", Offset: 3},
		{Text: "end", Offset: 7},
		{Text: "of", Offset: 11},
		{Text: "the", Offset: 14},
		{Text: "day", Offset: 18},
	}
	assertTokens(t, tokens, want)
}

func TestTokenize_Multiline(t *testing.T) {
	tokens := Tokenize("foo\nbar baz")
	want := []Token{
		{Text: "foo", Offset: 0},
		{Text: "bar", Offset: 4},
		{Text: "baz", Offset: 8},
	}
	assertTokens(t, tokens, want)
}

func TestTokenize_Empty(t *testing.T) {
	if Tokenize("") != nil {
		t.Fatal("expected nil for empty string")
	}
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
