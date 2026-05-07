package engine

import "github.com/nickg/plint/internal/parser"

// SpellChecker checks whether a word is correctly spelled.
// gospell.Checker satisfies this interface.
type SpellChecker interface {
	Spell(word string) bool
}

// SpellCheck returns a Hit for each token in doc that the checker rejects.
// Only nodes allowed by scope are checked. Tokens are passed to the checker
// in their original case (not lowercased), so the checker can apply
// hunspell's case-folding rules correctly.
func SpellCheck(doc *parser.Document, checker SpellChecker, scope Scope, rule Rule) []Hit {
	var hits []Hit
	for _, node := range doc.Nodes {
		if !scope.allows(node.Type) {
			continue
		}
		tokens := Tokenize(node.Text)
		for _, tok := range tokens {
			word := node.Text[tok.Offset : tok.Offset+tok.Len]
			if checker.Spell(word) {
				continue
			}
			hits = append(hits, Hit{
				Rule:      rule,
				Offset:    node.Offset + tok.Offset,
				EndOffset: node.Offset + tok.Offset + tok.Len,
				Len:       1,
			})
		}
	}
	return hits
}
