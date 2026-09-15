package testutils

import (
	"bytes"
	"filippo.io/age"
	"os"
	"testing"
)

func CreateTempEncryptedFile(t *testing.T, passphrase string, plaintext string) string {
	t.Helper()

	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

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

	_, err = tmpfile.Write(buf.Bytes())
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	t.Cleanup(func() {
		os.Remove(tmpfile.Name())
	})

	return tmpfile.Name()
}

func CreateTempFile(t *testing.T, s string) string {
	f, err := os.CreateTemp("", "gossh_test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	f.WriteString(s)
	f.Close()

	t.Cleanup(func() {
		os.Remove(f.Name())
	})
	return f.Name()
}
