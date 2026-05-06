package engine

import (
	"strings"

	"github.com/nickg/plint/internal/parser"
)

// Filter removes hits suppressed by per-document metadata.
// disable removes all hits for a rule by ID; allow removes hits whose matched
// text (case-insensitive) appears in the allow list.
func Filter(hits []Hit, meta parser.DocumentMeta, src []byte) []Hit {
	if len(meta.Disable) == 0 && len(meta.Allow) == 0 {
		return hits
	}

	disabled := make(map[string]bool, len(meta.Disable))
	for _, id := range meta.Disable {
		disabled[id] = true
	}

	allowed := make(map[string]bool, len(meta.Allow))
	for _, phrase := range meta.Allow {
		allowed[strings.ToLower(phrase)] = true
	}

	out := hits[:0]
	for _, h := range hits {
		if disabled[h.Rule.Name] {
			continue
		}
		if len(allowed) > 0 {
			match := strings.ToLower(string(src[h.Offset:h.EndOffset]))
			if allowed[match] {
				continue
			}
		}
		out = append(out, h)
	}
	return out
}
