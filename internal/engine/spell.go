package engine

import "github.com/nickg/plint/internal/parser"

// SpellChecker checks whether a word is correctly spelled.
// gospell.Checker satisfies this interface.
type SpellChecker interface {
	Spell(word string) bool
}

// Suggester is an optional interface that SpellCheckers may implement.
// SpellCheck uses it via type assertion to populate Hit.Suggestions.
// Returning []string (not gospell.Suggestion) keeps this package free of
// gospell types.
type Suggester interface {
	Suggest(word string, limit int) []string
}

// suggestionLimit is the maximum number of suggestions to attach to a hit.
const suggestionLimit = 5

// SpellCheck returns a Hit for each token in doc that the checker rejects.
// Only nodes allowed by scope are checked. Tokens are passed to the checker
// in their original case (not lowercased), so the checker can apply
// hunspell's case-folding rules correctly.
// If checker also implements Suggester, up to suggestionLimit suggestions
// are attached to each hit.
func SpellCheck(doc *parser.Document, checker SpellChecker, scope Scope, rule Rule) []Hit {
	suggester, canSuggest := checker.(Suggester)

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
			h := Hit{
				Rule:      rule,
				Offset:    node.Offset + tok.Offset,
				EndOffset: node.Offset + tok.Offset + tok.Len,
				Len:       1,
			}
			if canSuggest {
				h.Suggestions = suggester.Suggest(word, suggestionLimit)
			}
			hits = append(hits, h)
		}
	}
	return hits
}
