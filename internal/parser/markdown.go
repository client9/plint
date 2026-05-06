package parser

import (
	"encoding/json"
	"fmt"

	"github.com/client9/tojson"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// NodeType is the canonical node type vocabulary shared across parsers.
type NodeType string

const (
	NodeParagraph  NodeType = "paragraph"
	NodeHeading    NodeType = "heading"
	NodeBlockquote NodeType = "blockquote"
	NodeCode       NodeType = "code"
	NodeListItem   NodeType = "list-item"
)

// TextNode is a span of prose text with its position and type.
type TextNode struct {
	Text   string
	Offset int // byte offset into the original source
	Type   NodeType
}

// FrontMatter holds raw parsed front matter. Fields are populated as needed.
type FrontMatter struct {
	Raw []byte // JSON-encoded front matter, nil if absent
}

// Document is the result of parsing a source file.
type Document struct {
	Nodes       []TextNode
	FrontMatter FrontMatter
	Source      string // filename, for output
}

// ParseMarkdown parses src into a Document. source is used as the filename in output.
func ParseMarkdown(src []byte, source string) (*Document, error) {
	meta, body, err := tojson.FromFrontMatter(src)
	if err != nil {
		return nil, fmt.Errorf("front matter: %w", err)
	}

	// Validate meta is well-formed JSON if present.
	if meta != nil {
		if !json.Valid(meta) {
			return nil, fmt.Errorf("front matter: invalid JSON after conversion")
		}
	}

	nodes, err := parseBody(body)
	if err != nil {
		return nil, err
	}

	return &Document{
		Nodes:       nodes,
		FrontMatter: FrontMatter{Raw: meta},
		Source:      source,
	}, nil
}

func parseBody(src []byte) ([]TextNode, error) {
	reader := text.NewReader(src)
	doc := goldmark.DefaultParser().Parse(reader)

	var nodes []TextNode
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		nt, ok := nodeType(n)
		if !ok {
			return ast.WalkContinue, nil
		}

		text := extractText(n, src)
		if text == "" {
			return ast.WalkContinue, nil
		}

		offset := 0
		if lines := n.Lines(); lines.Len() > 0 {
			offset = lines.At(0).Start
		}

		nodes = append(nodes, TextNode{
			Text:   text,
			Offset: offset,
			Type:   nt,
		})

		return ast.WalkSkipChildren, nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking AST: %w", err)
	}

	return nodes, nil
}

// nodeType maps goldmark AST node kinds to canonical NodeTypes.
// Returns false for node types we don't extract text from.
func nodeType(n ast.Node) (NodeType, bool) {
	switch n.Kind() {
	case ast.KindParagraph, ast.KindTextBlock:
		return NodeParagraph, true
	case ast.KindHeading:
		return NodeHeading, true
	case ast.KindBlockquote:
		return NodeBlockquote, true
	case ast.KindFencedCodeBlock, ast.KindCodeBlock:
		return NodeCode, true
	case ast.KindListItem:
		return NodeListItem, true
	default:
		return "", false
	}
}

// extractText concatenates all text from a node.
// Block nodes (paragraphs, code) carry line segments; inline container nodes
// (headings) carry child Text nodes.
func extractText(n ast.Node, src []byte) string {
	lines := n.Lines()
	if lines.Len() > 0 {
		var buf []byte
		for i := 0; i < lines.Len(); i++ {
			seg := lines.At(i)
			buf = append(buf, seg.Value(src)...)
		}
		return string(buf)
	}
	// Walk children collecting Text node segments (used by headings).
	var buf []byte
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if t, ok := child.(*ast.Text); ok {
			seg := t.Segment
			buf = append(buf, seg.Value(src)...)
		}
	}
	return string(buf)
}
