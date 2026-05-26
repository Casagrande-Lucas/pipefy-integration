package valueobject

import (
	"testing"
)

func TestNewPriority(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		patrimonyValue float64
		want           Priority
	}{
		{
			name:           "exactly at threshold is high priority",
			patrimonyValue: 200_000.0,
			want:           PriorityHigh,
		},
		{
			name:           "above threshold is high priority",
			patrimonyValue: 200_000.01,
			want:           PriorityHigh,
		},
		{
			name:           "well above threshold is high priority",
			patrimonyValue: 1_000_000.0,
			want:           PriorityHigh,
		},
		{
			name:           "one cent below threshold is normal priority",
			patrimonyValue: 199_999.99,
			want:           PriorityNormal,
		},
		{
			name:           "zero is normal priority",
			patrimonyValue: 0,
			want:           PriorityNormal,
		},
		{
			name:           "negative value is normal priority",
			patrimonyValue: -1.0,
			want:           PriorityNormal,
		},
		{
			name:           "small positive value is normal priority",
			patrimonyValue: 100.0,
			want:           PriorityNormal,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := NewPriority(tc.patrimonyValue)
			if got != tc.want {
				t.Errorf("NewPriority(%v) = %q; want %q", tc.patrimonyValue, got, tc.want)
			}
		})
	}
}

func TestPriorityString(t *testing.T) {
	t.Parallel()

	if PriorityHigh.String() != "prioridade_alta" {
		t.Errorf("PriorityHigh.String() = %q; want %q", PriorityHigh.String(), "prioridade_alta")
	}
	if PriorityNormal.String() != "prioridade_normal" {
		t.Errorf("PriorityNormal.String() = %q; want %q", PriorityNormal.String(), "prioridade_normal")
	}
}
