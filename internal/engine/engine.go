package engine

import "github.com/nickg/plint/internal/parser"

// Lint runs the trie against every text node in doc that the scope allows.
// Hit offsets are absolute byte offsets into the original source.
func Lint(doc *parser.Document, trie *Trie, scope Scope) []Hit {
	var hits []Hit
	for _, node := range doc.Nodes {
		if !scope.allows(node.Type) {
			continue
		}
		tokens := Tokenize(node.Text)
		for _, h := range trie.Scan(tokens) {
			h.Offset += node.Offset
			h.EndOffset += node.Offset
			hits = append(hits, h)
		}
	}
	return hits
}
