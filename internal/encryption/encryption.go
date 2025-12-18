package encryption

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"filippo.io/age"
)

func GetPassphrase() string {
	var passphrase string

	pphrase_env := os.Getenv("GOSSH_PASSPHRASE")
	if pphrase_env != "" {
		passphrase = pphrase_env
	}

	return passphrase
}

func GetEncryptedPassword(encFile string) string {
	p := GetPassphrase()
	if p == "" {
		return ""
	}

	identity, err := age.NewScryptIdentity(p)
	if err != nil {
		fmt.Println("Could not create a new scrypt identity")
	}

	f, err := os.Open(encFile)
	if err != nil {
		fmt.Printf("Failed to open file: %v", err)
	}

	r, err := age.Decrypt(f, identity)
	if err != nil {
		fmt.Printf("Failed to open encrypted file: %v", err)
	}
	out := &bytes.Buffer{}
	if _, err := io.Copy(out, r); err != nil {
		fmt.Printf("Failed to read encrypted file: %v", err)
	}

	return out.String()
}
