package engine

import (
	"regexp"
	"strings"
)

var wordRe = regexp.MustCompile(`\pL+`)

type Token struct {
	Text   string
	Offset int // byte offset into the source string
	Len    int // byte length in source (before lowercasing; may differ for Unicode)
}

// Tokenize splits s into lowercase word tokens with their byte offsets.
func Tokenize(s string) []Token {
	indices := wordRe.FindAllStringIndex(s, -1)
	if indices == nil {
		return nil
	}
	tokens := make([]Token, len(indices))
	for i, idx := range indices {
		tokens[i] = Token{
			Text:   strings.ToLower(s[idx[0]:idx[1]]),
			Offset: idx[0],
			Len:    idx[1] - idx[0],
		}
	}
	return tokens
}
