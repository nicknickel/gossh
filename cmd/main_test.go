package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"filippo.io/age"

	"github.com/charmbracelet/log"
	"github.com/nicknickel/gossh/internal/connection"
	internal_log "github.com/nicknickel/gossh/internal/log"
)

// type Connection struct {
// 	Address      string `yaml:"address,omitempty"`
// 	User         string `yaml:"user,omitempty"`
// 	Description  string `yaml:"comment,omitempty"`
// 	IdentityFile string `yaml:"identity,omitempty"`
// 	PassFile     string `yaml:"passfile,omitempty"`
// 	SshProgram   string `yaml:"sshprogram,omitempty"`
// }
//
// type Item struct {
// 	Name    string
// 	Conn    Connection
// 	Checked bool
// 	Index   int
// }

func TestGetAuthentication(t *testing.T) {
	internal_log.Logger = log.New(io.Discard)
	// Create a temporary encrypted file
	passphrase := "testpass"
	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	plaintext := "secretpassword"
	buf := new(bytes.Buffer)
	w, err := age.Encrypt(buf, recipient)
	if err != nil {
		t.Fatalf("Failed to create encrypt writer: %v", err)
	}
	_, err = w.Write([]byte(plaintext))
	if err != nil {
		t.Fatalf("Failed to write plaintext: %v", err)
	}
	err = w.Close()
	if err != nil {
		t.Fatalf("Failed to close encrypt writer: %v", err)
	}

	tmpfile, err := os.CreateTemp("", "encrypted")
	if err != nil {
		t.Fatalf("Failed to create encrypted temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(buf.Bytes())
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	// create plain text temp file
	plaintmpfile, err := os.CreateTemp("", "plain")
	if err != nil {
		t.Fatalf("Failed to create plain text temp file: %v", err)
	}
	defer os.Remove(plaintmpfile.Name())

	plaintmpfile.Close()

	tests := []struct {
		item       connection.Item
		expected   string
		exactMatch bool
		passphrase string
	}{
		{
			expected:   "Password is " + plaintext,
			exactMatch: true,
			passphrase: passphrase,
			item: connection.Item{
				Name:    "encrypted passfile",
				Checked: false,
				Index:   1,
				Conn: connection.Connection{
					PassFile: tmpfile.Name(),
				},
			},
		},
		{
			expected:   "Password can be found in " + plaintmpfile.Name(),
			exactMatch: true,
			passphrase: passphrase,
			item: connection.Item{
				Name:    "unencrypted passfile",
				Checked: false,
				Index:   2,
				Conn: connection.Connection{
					PassFile: plaintmpfile.Name(),
				},
			},
		},
		{
			expected:   "Identity file is " + plaintmpfile.Name(),
			exactMatch: true,
			passphrase: passphrase,
			item: connection.Item{
				Name:    "unencrypted identity file",
				Checked: false,
				Index:   3,
				Conn: connection.Connection{
					IdentityFile: plaintmpfile.Name(),
				},
			},
		},
		{
			expected:   "Password can be found in " + plaintmpfile.Name(),
			exactMatch: true,
			passphrase: passphrase,
			item: connection.Item{
				Name:    "unencrypted passfile and identity file defined",
				Checked: false,
				Index:   4,
				Conn: connection.Connection{
					PassFile:     plaintmpfile.Name(),
					IdentityFile: tmpfile.Name(),
				},
			},
		},
		{
			expected:   "Temporary identity file is ",
			passphrase: passphrase,
			exactMatch: false,
			item: connection.Item{
				Name:    "encrypted identity file",
				Checked: false,
				Index:   5,
				Conn: connection.Connection{
					IdentityFile: tmpfile.Name(),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.item.Name, func(t *testing.T) {
			os.Setenv("GOSSH_PASSPHRASE", tt.passphrase)
			defer os.Unsetenv("GOSSH_PASSPHRASE")

			got := GetAuthentication(tt.item)
			if (got != tt.expected && tt.exactMatch) || (!strings.Contains(got, tt.expected) && !tt.exactMatch) {
				t.Errorf("GetAuthentication() = %v, want %v", got, tt.expected)
			}
		})
	}

}

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
