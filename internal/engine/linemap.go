package engine

// LineMap maps byte offsets to 1-based line and column numbers.
// Build it once from the source, then query it for each hit.
type LineMap []int // byte offset of the start of each line

// NewLineMap scans src and records the start offset of every line.
func NewLineMap(src []byte) LineMap {
	m := LineMap{0}
	for i, b := range src {
		if b == '\n' {
			m = append(m, i+1)
		}
	}
	return m
}

// Position returns the 1-based line and column for a byte offset.
func (m LineMap) Position(offset int) (line, col int) {
	lo, hi := 0, len(m)-1
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if m[mid] <= offset {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return lo + 1, offset - m[lo] + 1
}
