package engine

import "github.com/nickg/plint/internal/parser"

// Scope is the set of node types that will be linted.
type Scope map[parser.NodeType]bool

// DefaultScope lints paragraphs, headings, and list items; skips code and blockquotes.
var DefaultScope = Scope{
	parser.NodeParagraph: true,
	parser.NodeHeading:   true,
	parser.NodeListItem:  true,
}

func (s Scope) allows(t parser.NodeType) bool {
	return s[t]
}
