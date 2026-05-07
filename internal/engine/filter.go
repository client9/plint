package engine

import (
	"strings"

	"github.com/nickg/plint/internal/parser"
)

// Filter removes hits suppressed by per-document metadata.
// rules.disable removes all hits for a rule by ID; phrases.ignore removes hits
// whose matched text (case-insensitive) appears in the ignore list.
func Filter(hits []Hit, meta parser.DocumentMeta, src []byte) []Hit {
	if len(meta.Rules.Disable) == 0 && len(meta.Phrases.Ignore) == 0 {
		return hits
	}

	disabled := make(map[string]bool, len(meta.Rules.Disable))
	for _, id := range meta.Rules.Disable {
		disabled[id] = true
	}

	ignored := make(map[string]bool, len(meta.Phrases.Ignore))
	for _, phrase := range meta.Phrases.Ignore {
		ignored[strings.ToLower(phrase)] = true
	}

	out := hits[:0]
	for _, h := range hits {
		if disabled[h.Rule.Name] {
			continue
		}
		if len(ignored) > 0 {
			match := strings.ToLower(string(src[h.Offset:h.EndOffset]))
			if ignored[match] {
				continue
			}
		}
		out = append(out, h)
	}
	return out
}
