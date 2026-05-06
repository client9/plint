package engine

import (
	"fmt"

	"github.com/nickg/plint/internal/parser"
)

// ValidateMeta checks per-document metadata against the loaded rule set and
// returns a warning message for each disable entry that names an unknown rule.
func ValidateMeta(meta parser.DocumentMeta, knownIDs map[string]bool) []string {
	var warnings []string
	for _, id := range meta.Disable {
		if !knownIDs[id] {
			warnings = append(warnings, fmt.Sprintf("disable: unknown rule %q", id))
		}
	}
	return warnings
}
