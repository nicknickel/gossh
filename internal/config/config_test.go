package config

import (
	"os"
	"testing"

	"github.com/nicknickel/gossh/internal/connection"
)

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

	err = os.MkdirAll(home+"/.config/gossh", 0666)
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
	got := ConfigFiles()

	if len(got) < len(expected) {
		t.Errorf("ConfigFiles() = %v, want at least %v", got, expected)
	}
	for i, e := range expected {
		if got[i] != e {
			t.Errorf("ConfigFiles()[%d] = %v, want %v", i, got[i], e)
		}
	}

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

// Note: TestReadConnections would require mocking file system, which is more complex. Skipping for now or implement with test files.
