package rules

import "strings"

// Expand takes a phrase that may contain (a|b|c) alternation groups and
// returns all expanded combinations. Groups may be nested. Phrases without
// groups are returned as a single-element slice.
//
// Examples:
//
//	"only (marginally|slightly)"  → ["only marginally", "only slightly"]
//	"(deeply|truly) wrong"        → ["deeply wrong", "truly wrong"]
//	"plain phrase"                → ["plain phrase"]
func Expand(phrase string) []string {
	start := strings.IndexByte(phrase, '(')
	if start == -1 {
		return []string{phrase}
	}
	end := strings.IndexByte(phrase[start:], ')')
	if end == -1 {
		// malformed: no closing paren — treat as literal
		return []string{phrase}
	}
	end += start

	prefix := phrase[:start]
	alts := strings.Split(phrase[start+1:end], "|")
	suffix := phrase[end+1:]

	var out []string
	for _, alt := range alts {
		out = append(out, Expand(prefix+alt+suffix)...)
	}
	return out
}
