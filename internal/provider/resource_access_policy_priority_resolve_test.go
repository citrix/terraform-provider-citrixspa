package provider

import "testing"

func ptr(i int) *int { return &i }

func TestResolveRulePriorities(t *testing.T) {
	tests := []struct {
		name  string
		known []*int
		want  []int
	}{
		{"all omitted", []*int{nil, nil, nil}, []int{1, 2, 3}},
		{"all explicit", []*int{ptr(10), ptr(5), ptr(7)}, []int{10, 5, 7}},
		{"explicit then omitted", []*int{ptr(2), nil}, []int{2, 3}},
		{"omitted around explicit", []*int{nil, ptr(1), nil}, []int{2, 1, 3}},
		{"omitted collides with index", []*int{nil, ptr(1)}, []int{2, 1}},
		{"empty", []*int{}, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveRulePriorities(tt.known)
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d (%v)", len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("got %v, want %v", got, tt.want)
				}
			}
			// No omitted priority may duplicate an explicit one.
			explicit := map[int]bool{}
			for _, k := range tt.known {
				if k != nil {
					explicit[*k] = true
				}
			}
			for i, k := range tt.known {
				if k == nil && explicit[got[i]] {
					t.Fatalf("filled priority %d at index %d duplicates an explicit priority (%v)", got[i], i, got)
				}
			}
		})
	}
}
