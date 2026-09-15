package encryption

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"filippo.io/age"
	"github.com/nicknickel/gossh/internal/log"
)

func GetPassphrase() string {
	var passphrase string

	pphrase_env := os.Getenv("GOSSH_PASSPHRASE")
	if pphrase_env != "" {
		passphrase = pphrase_env
	}

	return passphrase
}

// GetEncryptedContents returns the contents
// of an encrypted file encrypted with age. Returns ""
// if unable to decrypt.
func GetEncryptedContents(encFile string) string {
	if encFile == "" {
		return ""
	}

	p := GetPassphrase()
	if p == "" {
		return ""
	}

	identity, err := age.NewScryptIdentity(p)
	if err != nil {
		log.Logger.Error("Could not create a new scrypt identity", "err", err)
		return ""
	}

	f, err := os.Open(encFile)
	if err != nil {
		log.Logger.Error("Failed to open file", "file", encFile, "err", err)
		return ""
	}

	r, err := age.Decrypt(f, identity)
	if err != nil {
		log.Logger.Error("Failed to open encrypted file", "file", encFile, "err", err)
		return ""
	}
	out := &bytes.Buffer{}
	if _, err := io.Copy(out, r); err != nil {
		log.Logger.Error("Failed to read encrypted file", "file", encFile, "err", err)
		return ""
	}

	return out.String()
}

func GetEncryptedIdentity(encFile string) string {
	if encFile == "" {
		return ""
	}

	identityContents := GetEncryptedContents(encFile)
	if identityContents != "" {
		pattern := fmt.Sprintf("%v.pem.*", filepath.Base(encFile))
		f, err := os.CreateTemp("", pattern)
		if err != nil {
			log.Logger.Error("Failed to create temp identity file", "file", encFile, "err", err)
			return ""
		}

		defer f.Close()

		_, err = f.Write([]byte(identityContents))
		return f.Name()
	}
	return ""
}

// EncryptFile returns the path to a file
// that is an encrypted version of f.
// Returns "" if unable. Encryption is
// Passphrase-based using GOSSH_PASSPHRASE.
func EncryptFile(src string) string {
	if src == "" {
		return ""
	}

	p := GetPassphrase()
	if p == "" {
		return ""
	}

	srcContents, err := os.ReadFile(src)
	if err != nil || len(srcContents) == 0 {
		log.Logger.Error("Failed to read unencrypted file or file empty", "file", src, "err", err)
		return ""
	}

	pattern := fmt.Sprintf("%v.enc.*", filepath.Base(src))
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		log.Logger.Error("Failed to create temp encrypted file", "err", err)
		return ""
	}

	r, err := age.NewScryptRecipient(p)
	if err != nil {
		log.Logger.Error("Could not create a new scrypt recipient", "err", err)
		return ""
	}

	w, err := age.Encrypt(f, r)
	if err != nil {
		log.Logger.Error("Failed to create encrypted file", "err", err)
		return ""
	}

	if _, err := w.Write(srcContents); err != nil {
		log.Logger.Error("Failed to write temp encrypted file", "err", err)
		return ""
	}
	if err := w.Close(); err != nil {
		log.Logger.Error("Failed to close temp encrypted file", "err", err)
		return ""
	}

	return f.Name()
}
