package log

import (
	"os"
	"path/filepath"

	"strconv"

	"github.com/charmbracelet/log"
)

var Logger *log.Logger

type RotatingFile struct {
	file        *os.File
	path        string
	currentSize int64
	maxSize     int64
}

func (r *RotatingFile) Write(p []byte) (n int, err error) {
	if r.currentSize+int64(len(p)) > r.maxSize {
		err = r.rotate()
		if err != nil {
			return
		}
	}
	n, err = r.file.Write(p)
	r.currentSize += int64(n)
	return
}

func (r *RotatingFile) rotate() error {
	r.file.Close()
	oldPath := r.path + ".old"
	os.Remove(oldPath)
	err := os.Rename(r.path, oldPath)
	if err != nil {
		return err
	}
	r.file, err = os.OpenFile(r.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	fi, err := os.Stat(r.path)
	if err == nil {
		r.currentSize = fi.Size()
	} else {
		r.currentSize = 0
	}
	return nil
}

func Init() {
	debug := os.Getenv("GOSSH_DEBUG")
	level := log.InfoLevel
	if debug != "" {
		level = log.DebugLevel
	}

	home, err := os.UserHomeDir()
	if err != nil {
		Logger = log.NewWithOptions(os.Stderr, log.Options{Level: level})
		return
	}

	logPath := filepath.Join(home, ".gossh.log")

	rolloverStr := os.Getenv("GOSSH_LOG_ROLLOVER")
	maxSize := int64(1024 * 1024) // default 1MB
	if rolloverStr != "" {
		var parsed int64
		parsed, err = strconv.ParseInt(rolloverStr, 10, 64)
		if err == nil && parsed > 0 {
			maxSize = parsed
		}
	}

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		Logger = log.NewWithOptions(os.Stderr, log.Options{Level: level})
		return
	}

	var currentSize int64
	fi, err := os.Stat(logPath)
	if err == nil {
		currentSize = fi.Size()
	} else {
		currentSize = 0
	}

	rot := &RotatingFile{
		file:        f,
		path:        logPath,
		currentSize: currentSize,
		maxSize:     maxSize,
	}

	Logger = log.NewWithOptions(rot, log.Options{
		Level:  level,
		Prefix: "gossh",
	})
}
