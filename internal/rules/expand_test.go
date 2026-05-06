package rules

import (
	"reflect"
	"testing"
)

func TestExpand(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{
			"plain phrase",
			[]string{"plain phrase"},
		},
		{
			"only (marginally|slightly)",
			[]string{"only marginally", "only slightly"},
		},
		{
			"(deeply|truly) wrong",
			[]string{"deeply wrong", "truly wrong"},
		},
		{
			"(a|b|c)",
			[]string{"a", "b", "c"},
		},
		{
			// nested groups
			"(only|just) (marginally|slightly) better",
			[]string{
				"only marginally better",
				"only slightly better",
				"just marginally better",
				"just slightly better",
			},
		},
		{
			// malformed: no closing paren — returned as-is
			"only (marginally",
			[]string{"only (marginally"},
		},
		{
			// no group
			"make no mistake",
			[]string{"make no mistake"},
		},
		{
			// suffix outside the group
			"only (marginal|slight|bare)ly",
			[]string{"only marginally", "only slightly", "only barely"},
		},
	}

	for _, tt := range tests {
		got := Expand(tt.in)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Expand(%q)\n  got  %v\n  want %v", tt.in, got, tt.want)
		}
	}
}
