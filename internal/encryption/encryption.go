package encryption

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"filippo.io/age"
)

func GetPassphrase() string {
	var passphrase string

	pphrase_file := os.Getenv("GOSSH_PASSPHRASE_FILE")
	if pphrase_file != "" {
		b, err := os.ReadFile(pphrase_file)
		if err != nil {
			fmt.Printf("Could not read %v\n", pphrase_file)
		}
		passphrase = string(b)
	}

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
		log.Fatal("Could not create a new scrypt identity")
	}

	f, err := os.Open(encFile)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}

	r, err := age.Decrypt(f, identity)
	if err != nil {
		log.Fatalf("Failed to open encrypted file: %v", err)
	}
	out := &bytes.Buffer{}
	if _, err := io.Copy(out, r); err != nil {
		log.Fatalf("Failed to read encrypted file: %v", err)
	}

	return out.String()
}
