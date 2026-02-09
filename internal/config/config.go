package config

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/nicknickel/gossh/internal/connection"
	"github.com/nicknickel/gossh/internal/log"
	"gopkg.in/yaml.v3"
)

// Helper function to collect all keys in a map
// maps.Keys returns an iter type which doesn't work with SortStableFunc
func Keys(m map[string]connection.Connection) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}

	return keys
}

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
	var conns []list.Item
	keys := Keys(c)

	slices.SortStableFunc(keys, func(a, b string) int {
		return strings.Compare(NormalizeString(a), NormalizeString(b))
	})

	for _, key := range keys {
		conns = append(conns, connection.Item{Name: key, Conn: c[key], Checked: false})
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

		fc := make(map[string]connection.Connection)
		err = yaml.Unmarshal(f, fc)
		if err != nil {
			debug := os.Getenv("GOSSH_DEBUG")
			if debug != "" {
				log.Logger.Debug("Error unmarshalling config", "file", file, "err", err)
			}
		} else {

			// Need to look at each identity and passfile and resolve any relative paths
			// which allows for program to be called from any directory
			keys := Keys(fc)
			for _, key := range keys {
				if fc[key].IdentityFile != "" && !filepath.IsAbs(fc[key].IdentityFile) {
					p := filepath.Join(filepath.Dir(file), fc[key].IdentityFile)
					tConn := fc[key]
					tConn.IdentityFile = filepath.Clean(p)
					fc[key] = tConn
				}
				if fc[key].PassFile != "" && !filepath.IsAbs(fc[key].PassFile) {
					p := filepath.Join(filepath.Dir(file), fc[key].PassFile)
					tConn := fc[key]
					tConn.PassFile = filepath.Clean(p)
					fc[key] = tConn
				}
			}

			maps.Copy(config, fc)
		}
	}

	return SortConns(config)
}
