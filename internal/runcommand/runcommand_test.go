package runcommand

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"filippo.io/age"
	"github.com/nicknickel/gossh/internal/connection"
	"github.com/nicknickel/gossh/internal/log"
	"io"
)

func TestGetPasswordTemplate(t *testing.T) {
	log.Init()

	if _, err := exec.LookPath("sshpass"); err != nil {
		os.WriteFile("sshpass", []byte(""), os.FileMode(0777))
		d, _ := os.Getwd()
		os.Setenv("PATH", d)
		defer os.Unsetenv("PATH")
		defer os.Remove("sshpass")
	}

	type testDef struct {
		connItem        connection.Item
		expectedCommand []string
		expectedEnv     []string
		thrownError     error
		unsetEnv        bool
	}

	encTest := testDef{
		connItem: connection.Item{
			Name: "encrypted passfile",
			Conn: connection.Connection{
				PassFile: "./gossh_test_enc_passfile",
			},
		},
		expectedCommand: []string{"sshpass", "-e"},
		expectedEnv:     []string{"SSHPASS=test"},
		thrownError:     nil,
		unsetEnv:        false,
	}

	tests := []struct {
		connItem        connection.Item
		expectedCommand []string
		expectedEnv     []string
		thrownError     error
		unsetEnv        bool
	}{
		{
			connItem: connection.Item{
				Name: "no passfile",
			},
			expectedCommand: []string{},
			expectedEnv:     []string{},
			thrownError:     errors.New("passfile parameter not defined"),
			unsetEnv:        false,
		},
		{
			connItem: connection.Item{
				Name: "plain passfile",
				Conn: connection.Connection{
					PassFile: "./gossh_test_plain_passfile",
				},
			},
			expectedCommand: []string{"sshpass", "-f", "{{.Conn.PassFile}}"},
			expectedEnv:     nil,
			thrownError:     nil,
			unsetEnv:        false,
		},
		{
			connItem: connection.Item{
				Name: "no sshpass",
			},
			expectedCommand: []string{},
			expectedEnv:     []string{},
			thrownError:     errors.New("sshpass not found"),
			unsetEnv:        true,
		},
	}

	passphrase := "test"
	encFile := "./gossh_test_enc_passfile"
	identity, err := age.NewScryptRecipient(passphrase)
	f, err2 := os.Create(encFile)
	defer os.Remove(encFile)
	e, err3 := age.Encrypt(f, identity)
	_, err4 := io.WriteString(e, passphrase)
	err5 := e.Close()

	if err == nil && err2 == nil && err3 == nil && err4 == nil && err5 == nil {
		os.Setenv("GOSSH_PASSPHRASE", passphrase)
		defer os.Unsetenv("GOSSH_PASSPHRASE")
		tests = append(tests, encTest)
	} else {
		fmt.Println("cannot test encrypted file")
	}

	for _, tt := range tests {
		t.Run(tt.connItem.Name, func(t *testing.T) {

			if tt.unsetEnv {
				currPath := os.Getenv("PATH")
				os.Unsetenv("PATH")
				defer os.Setenv("PATH", currPath)
			}
			cmd, env, err := GetPasswordTemplate(&tt.connItem)
			if !reflect.DeepEqual(cmd, tt.expectedCommand) {
				t.Errorf("GetPasswordTemplate() want command of %v, got %v", tt.expectedCommand, cmd)
			}
			if !reflect.DeepEqual(env, tt.expectedEnv) {
				t.Errorf("GetPasswordTemplate() want env of %v, got %v", tt.expectedEnv, env)
			}
			if !reflect.DeepEqual(err, tt.thrownError) {
				t.Errorf("GetPasswordTemplate() want error of %v, got %v", tt.thrownError, err)
			}
		})
	}
}
