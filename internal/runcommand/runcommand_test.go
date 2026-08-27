package runcommand

import (
	"github.com/nicknickel/gossh/internal/connection"
	"reflect"
	"testing"
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

func TestGetPasswordTemplate(t *testing.T) {
	// if _, err := os.Open("sshpass"); err != nil {
	// 	os.WriteFile("sshpass", []byte(""), os.FileMode(0777))
	// 	defer os.Remove("sshpass")
	// }

	tests := []struct {
		connItem        connection.Item
		expectedCommand []string
		expectedEnv     []string
		thrownError     error
	}{
		// {
		// 	connItem: connection.Item{
		// 		Name: "no passfile",
		// 	},
		// 	expectedCommand: []string{},
		// 	expectedEnv:     []string{},
		// 	thrownError:     errors.New("sshpass not found"),
		// },
		{
			connItem: connection.Item{
				Name: "encrypted passfile",
				Conn: connection.Connection{
					PassFile: "./gossh_test_enc_passfile",
				},
			},
			expectedCommand: []string{"sshpass", "-e"},
			expectedEnv:     []string{"SSHPASS=./gossh_test_plan_passfile"},
			thrownError:     nil,
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.connItem.Name, func(t *testing.T) {
			cmd, env, err := GetPasswordTemplate(&tt.connItem)
			if reflect.DeepEqual(cmd, tt.expectedCommand) {
				t.Errorf("GetPasswordTemplate() want command of %v, got %v", tt.expectedCommand, cmd)
			}
			if reflect.DeepEqual(env, tt.expectedEnv) {
				t.Errorf("GetPasswordTemplate() want env of %v, got %v", tt.expectedEnv, env)
			}
			if reflect.DeepEqual(err, tt.thrownError) {
				t.Errorf("GetPasswordTemplate() want error of %v, got %v", tt.thrownError, err)
			}
		})
	}
}
