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

func GetEncryptedContents(encFile string) string {
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
