package config

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/nicknickel/gossh/internal/connection"
	"github.com/nicknickel/gossh/internal/log"
	"gopkg.in/yaml.v3"
)

func ConfigFiles() []string {
	loc := []string{"./gossh.yml"}
	userHome, err := os.UserHomeDir()

	if err == nil {
		loc = append(loc, userHome+"/.config/gossh/gossh.yml")
		loc = append(loc, userHome+"/.gossh.yml")
		loc = append(loc, userHome+"/gossh.yml")
	}

	configDir := os.Getenv("GOSSH_CONFIGDIR")
	if configDir == "" {
		return loc
	}

	files, err := os.ReadDir(configDir)
	if err != nil {
		if os.IsExist(err) {
			log.Logger.Error("Cannot read config directory", "dir", configDir, "err", err)
		}
	} else {
		for _, file := range files {
			loc = append(loc, fmt.Sprintf("%v/%v", configDir, file.Name()))
		}
	}

	return loc

}

func NormalizeString(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(s), "-", ""), "_", "")
}

func SortConns(c map[string]connection.Connection) []list.Item {
	keys := make([]string, len(c))
	var conns []list.Item

	i := 0
	for k := range c {
		keys[i] = k
		i++
	}

	slices.SortStableFunc(keys, func(a, b string) int {
		return strings.Compare(NormalizeString(a), NormalizeString(b))
	})

	for _, key := range keys {
		conns = append(conns, connection.Item{Name: key, Conn: c[key]})
	}

	return conns
}

func ReadConnections() []list.Item {
	config := make(map[string]connection.Connection)

	for _, file := range ConfigFiles() {
		f, err := os.ReadFile(file)
		if err != nil {
			if os.IsExist(err) {
				log.Logger.Error("Error reading config file", "file", file, "err", err)
			}
			continue
		}

		err = yaml.Unmarshal(f, config)
		if err != nil {
			debug := os.Getenv("GOSSH_DEBUG")
			if debug != "" {
				log.Logger.Debug("Error unmarshalling config", "file", file, "err", err)
			}
		}
	}

	return SortConns(config)
}
