package config

import (
	"os"
	"strings"
	"testing"

	"github.com/nicknickel/gossh/internal/connection"
)

func prepareConfigFiles(t *testing.T, expected []string) {
	t.Helper()

	home, _ := os.UserHomeDir()

	err := os.MkdirAll(home+"/.config/gossh", 0666)
	if err != nil {
		t.Errorf("Could not prepare %v for test", home+"/.config/gossh")
	}

	for _, f := range expected {
		_, err := os.ReadFile(f)
		if err != nil {
			if os.IsNotExist(err) {
				err := os.WriteFile(f, []byte(""), 0666)
				if err != nil {
					t.Errorf("Could not prepare %v for test", f)
				}
			}
		}
	}

	t.Cleanup(func() {
		// cleanup blank files
		for _, f := range expected {
			data, _ := os.ReadFile(f)
			if string(data) == "" {
				os.Remove(f)
			}
		}

		files, err := os.ReadDir(home + "/.config/gossh")
		if err == nil && len(files) == 0 {
			os.Remove(home + "/.config/gossh")
		}
	})

}

func TestNormalizeString(t *testing.T) {
	testCases := []string{
		"examplestring",
		"example_string",
		"example-string",
		"exampl_e-string",
	}

	for _, testCase := range testCases {
		t.Run(testCase, func(t *testing.T) {
			result := NormalizeString(testCase)
			if result != "examplestring" {
				t.Errorf("expected NormalizeString(%v) to be examplestring but got %v", testCase, result)
			}
		})
	}

	result := NormalizeString("")
	if result != "" {
		t.Errorf("expected NormalizeString(nil) to be nil but got %v", result)
	}
}

func TestConfigFiles(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get user home dir")
	}

	expected := []string{
		"./gossh.yml",
		home + "/.config/gossh/gossh.yml",
		home + "/.gossh.yml",
		home + "/gossh.yml",
	}
	prepareConfigFiles(t, expected)

	tests := []struct {
		name          string
		expectExact   bool
		expectAtLeast bool
		envVar        string
		addedDir      string
	}{
		{
			name:          "no env",
			expectExact:   true,
			expectAtLeast: false,
			envVar:        "",
			addedDir:      "",
		},
		{
			name:          "bad env",
			expectExact:   true,
			expectAtLeast: false,
			envVar:        "./badddir",
			addedDir:      "",
		},
		{
			name:          "good env",
			expectExact:   false,
			expectAtLeast: true,
			envVar:        "./",
			addedDir:      "./",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("GOSSH_CONFIGDIR", tt.envVar)
			defer os.Unsetenv("GOSSH_CONFIGDIR")

			got := ConfigFiles()

			if len(got) != len(expected) && tt.expectExact {
				t.Errorf("ConfigFiles() = %v, want exactly %v", got, expected)
			}
			if len(got) < len(expected) && tt.expectAtLeast {
				t.Errorf("ConfigFiles() = %v, want at least %v", got, expected)
			}
			for i, e := range expected {
				if got[i] != e {
					t.Errorf("ConfigFiles()[%d] = %v, want %v", i, got[i], e)
				}
			}
			if !strings.Contains(got[len(got)-1], tt.addedDir) && tt.addedDir != "" {
				t.Errorf("ConfigFiles() = %v, want %v", got[len(got)-1], "./")
			}
		})
	}
}

func TestSortConns(t *testing.T) {
	conns := map[string]connection.Connection{
		"B-host": {Address: "b"},
		"A_host": {Address: "a"},
		"cHost":  {Address: "c"},
	}

	got := SortConns(conns)
	if len(got) != 3 {
		t.Errorf("SortConns() len = %d, want 3", len(got))
	}
	if got[0].(connection.Item).Name != "A_host" {
		t.Errorf("First item = %s, want A_host", got[0].(connection.Item).Name)
	}
	if got[1].(connection.Item).Name != "B-host" {
		t.Errorf("Second item = %s, want B-host", got[1].(connection.Item).Name)
	}
	if got[2].(connection.Item).Name != "cHost" {
		t.Errorf("Third item = %s, want cHost", got[2].(connection.Item).Name)
	}
}

func TestKeys(t *testing.T) {
	maps := map[string]connection.Connection{
		"B-host": {Address: "b"},
		"A_host": {Address: "a"},
		"cHost":  {Address: "c"},
	}

	got := Keys(maps)
	if len(got) != 3 {
		t.Errorf("Keys() len = %d, want 3", len(got))
	}
	if got[0] != "B-host" {
		t.Errorf("First item = %s, want B-host", got[0])
	}
	if got[1] != "A_host" {
		t.Errorf("Second item = %s, want A_host", got[1])
	}
	if got[2] != "cHost" {
		t.Errorf("Third item = %s, want cHost", got[2])
	}
}

// Note: TestReadConnections would require mocking file system, which is more complex. Skipping for now or implement with test files.
