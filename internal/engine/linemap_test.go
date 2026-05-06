package engine

import "testing"

func TestLineMap(t *testing.T) {
	src := []byte("foo\nbar\nbaz")
	m := NewLineMap(src)

	tests := []struct {
		offset, line, col int
	}{
		{0, 1, 1}, // 'f'
		{2, 1, 3}, // 'o'
		{4, 2, 1}, // 'b' of "bar"
		{7, 2, 4}, // '\n' after "bar" — still line 2
		{8, 3, 1}, // 'b' of "baz"
	}
	for _, tt := range tests {
		line, col := m.Position(tt.offset)
		if line != tt.line || col != tt.col {
			t.Errorf("Position(%d) = (%d,%d), want (%d,%d)", tt.offset, line, col, tt.line, tt.col)
		}
	}
}
