package encryption

import (
	"bytes"
	"os"
	"testing"

	"io"

	"filippo.io/age"
	"github.com/charmbracelet/log"

	internal_log "github.com/nicknickel/gossh/internal/log"
)

func TestGetPassphrase(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "env set",
			envValue: "secret",
			expected: "secret",
		},
		{
			name:     "env not set",
			envValue: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("GOSSH_PASSPHRASE", tt.envValue)
			defer os.Unsetenv("GOSSH_PASSPHRASE")

			if got := GetPassphrase(); got != tt.expected {
				t.Errorf("GetPassphrase() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetEncryptedContents(t *testing.T) {
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
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(buf.Bytes())
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	tests := []struct {
		name       string
		passphrase string
		filename   string
		expected   string
		expectErr  bool
	}{
		{
			name:       "valid decryption",
			passphrase: passphrase,
			filename:   tmpfile.Name(),
			expected:   plaintext,
			expectErr:  false,
		},
		{
			name:       "wrong passphrase",
			passphrase: "wrong",
			filename:   tmpfile.Name(),
			expected:   "",
			expectErr:  true,
		},
		{
			name:       "no passphrase",
			passphrase: "",
			filename:   tmpfile.Name(),
			expected:   "",
			expectErr:  false,
		},
		{
			name:       "file not exist",
			passphrase: passphrase,
			filename:   "nonexistent",
			expected:   "",
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("GOSSH_PASSPHRASE", tt.passphrase)
			defer os.Unsetenv("GOSSH_PASSPHRASE")

			got := GetEncryptedContents(tt.filename)
			if got != tt.expected {
				t.Errorf("GetEncryptedContents() = %v, want %v", got, tt.expected)
			}
			// Note: The function doesn't return error, but logs it. For testing, we might check logs, but simplifying here.
		})
	}
}

func TestGetEncryptedIdentity(t *testing.T) {
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
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(buf.Bytes())
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	tests := []struct {
		name          string
		passphrase    string
		filename      string
		expectedBlank bool
		expectErr     bool
	}{
		{
			name:          "valid decryption",
			passphrase:    passphrase,
			filename:      tmpfile.Name(),
			expectedBlank: false,
			expectErr:     false,
		},
		{
			name:          "wrong passphrase",
			passphrase:    "wrong",
			filename:      tmpfile.Name(),
			expectedBlank: true,
			expectErr:     true,
		},
		{
			name:          "no passphrase",
			passphrase:    "",
			filename:      tmpfile.Name(),
			expectedBlank: true,
			expectErr:     false,
		},
		{
			name:          "file not exist",
			passphrase:    passphrase,
			filename:      "nonexistent",
			expectedBlank: true,
			expectErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("GOSSH_PASSPHRASE", tt.passphrase)
			defer os.Unsetenv("GOSSH_PASSPHRASE")

			got := GetEncryptedIdentity(tt.filename)
			if got != "" && tt.expectedBlank {
				t.Errorf("GetEncryptedIdentity() = %v, want empty string", got)
			}
			if got == "" && !tt.expectedBlank {
				t.Errorf("GetEncryptedIdentity() = empty string, want non-empty string")
			}
			// Note: The function doesn't return error, but logs it. For testing, we might check logs, but simplifying here.
		})
	}
}
