package config

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/nicknickel/gossh/connection"
	"gopkg.in/yaml.v3"
)

func ConfigFiles() []string {
	loc := []string{"./gossh.yml",
		"./nccm.yml",
	}
	userHome, err := os.UserHomeDir()

	if err != nil {
		loc = append(loc, userHome+"/.config/nccm/nccm.yml")
		loc = append(loc, userHome+"/.nccm.yml")
		loc = append(loc, userHome+"/nccm.yml")
	}

	configDir := "/etc/nccm.d/"
	files, err := os.ReadDir(configDir)
	if err != nil {
		if os.IsExist(err) {
			fmt.Printf("Cannot read %v due to: %v\n", configDir, err)
		}
		return loc
	}

	for _, file := range files {
		loc = append(loc, configDir+file.Name())
	}

	return loc

}

func NormalizeString(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(s), "-", ""), "_", "")
}

func ReadConnections() []list.Item {
	config := make(map[string]connection.Connection)
	var conns []list.Item

	for _, file := range ConfigFiles() {
		f, err := os.ReadFile(file)
		if err != nil {
			if os.IsExist(err) {
				fmt.Println(err)
			}
			continue
		}

		err = yaml.Unmarshal(f, config)
		if err != nil {
			debug := os.Getenv("GOSSH_DEBUG")
			if debug != "" {
				fmt.Println(err)
			}
		}
	}

	// Source - https://stackoverflow.com/a
	// Posted by Vinay Pai, modified by community. See post 'Timeline' for change history
	// Retrieved 2025-11-06, License - CC BY-SA 3.0

	keys := make([]string, len(config))

	i := 0
	for k, _ := range config {
		keys[i] = k
		i++
	}

	slices.SortStableFunc(keys, func(a, b string) int {
		return strings.Compare(NormalizeString(a), NormalizeString(b))
	})

	for _, key := range keys {
		conns = append(conns, connection.Item{Name: key, Conn: config[key]})
	}

	return conns

}
