package runcommand

import (
	"errors"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nicknickel/gossh/internal/connection"
	"github.com/nicknickel/gossh/internal/log"
	"github.com/nicknickel/gossh/internal/testutils"
	"github.com/nicknickel/gossh/internal/utils"
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
	passphrase := "test"
	encFile := testutils.CreateTempEncryptedFile(t, passphrase, passphrase)
	os.Setenv("GOSSH_PASSPHRASE", passphrase)
	defer os.Unsetenv("GOSSH_PASSPHRASE")

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
		{
			connItem: connection.Item{
				Name: "encrypted passfile",
				Conn: connection.Connection{
					PassFile: encFile,
				},
			},
			expectedCommand: []string{"sshpass", "-e"},
			expectedEnv:     []string{"SSHPASS=test"},
			thrownError:     nil,
			unsetEnv:        false,
		},
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

func TestGetIdentityTemplate(t *testing.T) {
	passphrase := "test"
	encFile := testutils.CreateTempEncryptedFile(t, passphrase, passphrase)
	os.Setenv("GOSSH_PASSPHRASE", passphrase)
	defer os.Unsetenv("GOSSH_PASSPHRASE")

	tests := []struct {
		connItem        connection.Item
		expectedCommand []string
		expectedCleanup bool
		thrownError     error
	}{
		{
			connItem: connection.Item{
				Name: "no identity file",
			},
			expectedCommand: []string{},
			expectedCleanup: false,
			thrownError:     errors.New("No identity file indicated"),
		},
		{
			connItem: connection.Item{
				Name: "unencrypted file",
				Conn: connection.Connection{
					IdentityFile: "gossh_test_identityfile",
				},
			},
			expectedCommand: []string{"-i", "gossh_test_identityfile"},
			expectedCleanup: false,
			thrownError:     nil,
		},
		{
			connItem: connection.Item{
				Name: "encrypted identity file",
				Conn: connection.Connection{
					IdentityFile: encFile,
				},
			},
			expectedCommand: []string{"-i", encFile + ".pem."},
			expectedCleanup: true,
			thrownError:     nil,
		},
	}

	opt := cmp.Comparer(func(x string, y string) bool {
		if x == y {
			return true
		}
		if strings.Contains(x, y) || len(y) > len(x) {
			return true
		}
		return false
	})

	for _, tt := range tests {
		t.Run(tt.connItem.Name, func(t *testing.T) {
			cmd, cleanup, err := GetIdentityTemplate(&tt.connItem)

			if !cmp.Equal(cmd, tt.expectedCommand, opt) {
				t.Errorf("GetIdentityTemplate() want command %v, got %v", tt.expectedCommand, cmd)
			}
			if tt.expectedCleanup != cleanup {
				t.Errorf("GetIdentityTemplate() want cleanup %v, got %v", tt.expectedCleanup, cleanup)
			}
			if !reflect.DeepEqual(err, tt.thrownError) {
				t.Errorf("GetIdentityTemplate() want error %v, got %v", tt.thrownError, err)
			}
			if cleanup {
				os.Remove(cmd[1])
			}
		})
	}
}

func TestRenderTemplateSlice(t *testing.T) {
	tests := []struct {
		connItem       connection.Item
		template       []string
		expectedOutput []string
	}{
		{
			connItem: connection.Item{
				Name: "connect_to_host",
			},
			template:       utils.NewConnectTemplate(),
			expectedOutput: []string{"ssh", "connect_to_host"},
		},
		{
			connItem: connection.Item{
				Name: "command_to_run",
			},
			template:       utils.NewCommandTemplate("echo test"),
			expectedOutput: []string{"ssh", "command_to_run", "echo", "test"},
		},
		{
			connItem: connection.Item{
				Name: "receive_file",
			},
			template:       utils.NewReceiveTemplate("/remote/file", "/tmp/file"),
			expectedOutput: []string{"scp", "-rp", "receive_file:/remote/file", "/tmp/file_receive_file"},
		},
		{
			connItem: connection.Item{
				Name: "send_file",
			},
			template:       utils.NewSendTemplate("/local/file", "/remote/dest"),
			expectedOutput: []string{"scp", "-rp", "/local/file", "send_file:/remote/dest"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.connItem.Name, func(t *testing.T) {
			s := RenderTemplateSlice(&tt.template, tt.connItem)

			if !reflect.DeepEqual(s, tt.expectedOutput) {
				t.Errorf("RenderTemplateSlice() want  %v, got %v", tt.expectedOutput, s)
			}
		})
	}
}
