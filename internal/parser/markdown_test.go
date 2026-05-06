package parser

import (
	"testing"
)

func TestParseMarkdown_Paragraphs(t *testing.T) {
	src := []byte("First paragraph.\n\nSecond paragraph.\n")
	doc, err := ParseMarkdown(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Nodes) != 2 {
		t.Fatalf("got %d nodes, want 2", len(doc.Nodes))
	}
	for _, n := range doc.Nodes {
		if n.Type != NodeParagraph {
			t.Errorf("expected paragraph, got %q", n.Type)
		}
	}
}

func TestParseMarkdown_Heading(t *testing.T) {
	src := []byte("# My Heading\n\nA paragraph.\n")
	doc, err := ParseMarkdown(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	var heading, para *TextNode
	for i := range doc.Nodes {
		switch doc.Nodes[i].Type {
		case NodeHeading:
			heading = &doc.Nodes[i]
		case NodeParagraph:
			para = &doc.Nodes[i]
		}
	}
	if heading == nil {
		t.Fatal("no heading node found")
	}
	if para == nil {
		t.Fatal("no paragraph node found")
	}
	if heading.Text != "My Heading" {
		t.Errorf("heading text = %q, want %q", heading.Text, "My Heading")
	}
}

func TestParseMarkdown_CodeBlock(t *testing.T) {
	src := []byte("A paragraph.\n\n```\nsome code\n```\n")
	doc, err := ParseMarkdown(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	var para, code *TextNode
	for i := range doc.Nodes {
		switch doc.Nodes[i].Type {
		case NodeParagraph:
			para = &doc.Nodes[i]
		case NodeCode:
			code = &doc.Nodes[i]
		}
	}
	if para == nil {
		t.Error("expected paragraph node")
	}
	if code == nil {
		t.Error("expected code node — parser extracts all types; scope system filters")
	}
}

func TestParseMarkdown_FrontMatter(t *testing.T) {
	src := []byte("---\ntitle: My Doc\n---\n\nA paragraph.\n")
	doc, err := ParseMarkdown(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if doc.FrontMatter.Raw == nil {
		t.Fatal("expected front matter, got nil")
	}
	if len(doc.Nodes) != 1 || doc.Nodes[0].Type != NodeParagraph {
		t.Errorf("expected 1 paragraph node, got %+v", doc.Nodes)
	}
}

func TestParseMarkdown_NoFrontMatter(t *testing.T) {
	src := []byte("Just a paragraph.\n")
	doc, err := ParseMarkdown(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if doc.FrontMatter.Raw != nil {
		t.Error("expected nil front matter")
	}
}

func TestParseMarkdown_Offset(t *testing.T) {
	src := []byte("First.\n\nSecond.\n")
	doc, err := ParseMarkdown(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Nodes) != 2 {
		t.Fatalf("got %d nodes, want 2", len(doc.Nodes))
	}
	// "First." starts at offset 0, "Second." starts at offset 8
	if doc.Nodes[0].Offset != 0 {
		t.Errorf("node[0].Offset = %d, want 0", doc.Nodes[0].Offset)
	}
	if doc.Nodes[1].Offset != 8 {
		t.Errorf("node[1].Offset = %d, want 8", doc.Nodes[1].Offset)
	}
}
