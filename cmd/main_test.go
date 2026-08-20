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
	}

	tmuxIsInstalledTest := struct {
		name       string
		windowName string
		envTMUX    string
		envGOSSH   string
		expectErr  bool
	}{
		name:       "with tmux and GOSSH_TMUX",
		windowName: "test",
		envTMUX:    "/tmp/tmux",
		envGOSSH:   "1",
		expectErr:  false, // Assuming tmux command succeeds
	}

	// Can only test tmux if tmux is currently running
	isTmux := os.Getenv("TMUX")
	if isTmux != "" {
		tests = append(tests, tmuxIsInstalledTest)
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

	// cleanup tmux window
	HandleTmux("")
}
