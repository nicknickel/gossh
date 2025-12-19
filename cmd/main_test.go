package main

import (
	"os"
	"testing"
)

func TestHandleTmux(t *testing.T) {
	tests := []struct {
		name       string
		windowName string
		envTMUX    string
		envGOSSH   string
		expectErr  bool
	}{
		{
			name:       "no tmux",
			windowName: "test",
			envTMUX:    "",
			envGOSSH:   "",
			expectErr:  false,
		},
		{
			name:       "with tmux but no GOSSH_TMUX",
			windowName: "test",
			envTMUX:    "/tmp/tmux",
			envGOSSH:   "",
			expectErr:  false,
		},
		{
			name:       "with tmux and GOSSH_TMUX",
			windowName: "test",
			envTMUX:    "/tmp/tmux",
			envGOSSH:   "1",
			expectErr:  false, // Assuming tmux command succeeds
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldTMUX := os.Getenv("TMUX")
			oldGOSSH := os.Getenv("GOSSH_TMUX")
			os.Setenv("TMUX", tt.envTMUX)
			os.Setenv("GOSSH_TMUX", tt.envGOSSH)
			defer os.Setenv("TMUX", oldTMUX)
			defer os.Setenv("GOSSH_TMUX", oldGOSSH)

			err := HandleTmux(tt.windowName)
			if (err != nil) != tt.expectErr {
				t.Errorf("HandleTmux() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

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

// Other functions like runConnection involve exec, harder to test without mocks.
