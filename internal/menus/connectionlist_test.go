package menus

import "testing"

func TestFilterFunc(t *testing.T) {
	items := []string{
		"host1 addr1 user1 desc1",
		"host2 addr2 user2 desc2",
		"host3 addr3 user3 desc3",
	}

	tests := []struct {
		name     string
		term     string
		expected []int // indices of matched items
	}{
		{
			name:     "single term",
			term:     "host1",
			expected: []int{0},
		},
		{
			name:     "multiple terms",
			term:     "user2 desc2",
			expected: []int{1},
		},
		{
			name:     "no match",
			term:     "nonexistent",
			expected: []int{},
		},
		{
			name:     "all terms must match",
			term:     "host1 user3",
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ranks := FilterFunc(tt.term, items)
			if len(ranks) != len(tt.expected) {
				t.Errorf("FilterFunc() len = %d, want %d", len(ranks), len(tt.expected))
			}
			for i, r := range ranks {
				if r.Index != tt.expected[i] {
					t.Errorf("Rank %d index = %d, want %d", i, r.Index, tt.expected[i])
				}
			}
		})
	}
}
