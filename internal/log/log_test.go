package log

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/log"
)

func TestRotatingFile_WriteAndRotate(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	r := &RotatingFile{
		path:    logPath,
		maxSize: 10,
	}

	var err error
	r.file, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer r.file.Close()

	// Write small amount
	n, err := r.Write([]byte("12345"))
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}
	if n != 5 {
		t.Errorf("Wrote %d bytes, want 5", n)
	}
	if r.currentSize != 5 {
		t.Errorf("currentSize = %d, want 5", r.currentSize)
	}

	// Write to trigger rotate
	n, err = r.Write([]byte("6789012345"))
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}
	if n != 10 {
		t.Errorf("Wrote %d bytes, want 10", n)
	}
	if r.currentSize != 10 {
		t.Errorf("currentSize = %d, want 10 after rotate", r.currentSize)
	}

	// Check old file
	oldContent, err := os.ReadFile(logPath + ".old")
	if err != nil {
		t.Errorf("Failed to read old file: %v", err)
	}
	if string(oldContent) != "12345" {
		t.Errorf("old file content = %s, want 12345", oldContent)
	}

	// Check new file
	newContent, err := os.ReadFile(logPath)
	if err != nil {
		t.Errorf("Failed to read new file: %v", err)
	}
	if string(newContent) != "6789012345" {
		t.Errorf("new file content = %s, want 6789012345", newContent)
	}
}

func TestInit(t *testing.T) {
	home := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", home)
	defer os.Setenv("HOME", oldHome)

	tests := []struct {
		name          string
		debug         string
		rollover      string
		expectedLevel log.Level
		expectedPath  string
	}{
		{
			name:          "default",
			debug:         "",
			rollover:      "",
			expectedLevel: log.InfoLevel,
			expectedPath:  filepath.Join(home, ".gossh.log"),
		},
		{
			name:          "debug",
			debug:         "1",
			rollover:      "",
			expectedLevel: log.DebugLevel,
			expectedPath:  filepath.Join(home, ".gossh.log"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("GOSSH_DEBUG", tt.debug)
			os.Setenv("GOSSH_LOG_ROLLOVER", tt.rollover)
			defer os.Unsetenv("GOSSH_DEBUG")
			defer os.Unsetenv("GOSSH_LOG_ROLLOVER")

			Init()

			if Logger.GetLevel() != tt.expectedLevel {
				t.Errorf("Logger level = %v, want %v", Logger.GetLevel(), tt.expectedLevel)
			}
			// Check if file exists
			if _, err := os.Stat(tt.expectedPath); err != nil {
				t.Errorf("Log file not created: %v", err)
			}
		})
	}
}
