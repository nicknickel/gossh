package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/nicknickel/gossh/internal/connection"
	"github.com/nicknickel/gossh/internal/encryption"
	"github.com/nicknickel/gossh/internal/log"
	"github.com/nicknickel/gossh/internal/menus"
	"github.com/nicknickel/gossh/internal/runcommand"
)

var updateVersion bool
var version string = "dev"
var initialFilter string

func init() {
	log.Init()
	flag.BoolVar(&updateVersion, "update", false, "Pass this flag to update the gossh version to latest github release and exit")
	flag.StringVar(&initialFilter, "filter", "", "Pass this flag to filter the initial list of connections")
	flag.StringVar(&initialFilter, "f", "", "shorthand for filter")
}

func updateExecutable() error {

	if version == "dev" {
		return errors.New("dev version not upgradable")
	}

	latest, found, err := selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug("nicknickel/gossh"))
	if err != nil {
		return fmt.Errorf("error occurred while detecting version: %w", err)
	}
	if !found {
		return fmt.Errorf("latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH)
	}

	if latest.LessOrEqual(version) {
		fmt.Printf("Current version (%s) is the latest\n", version)
		return nil
	}

	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		return errors.New("could not locate executable path")
	}
	if err := selfupdate.UpdateTo(context.Background(), latest.AssetURL, latest.AssetName, exe); err != nil {
		return fmt.Errorf("error occurred while updating binary: %w", err)
	}
	fmt.Printf("Successfully updated to version %s\n", latest.Version())
	return nil
}

func HandleTmux(name string) error {
	var c *exec.Cmd

	if name != "" {
		c = exec.Command("tmux", "-2u", "rename-window", name)
	} else {
		c = exec.Command("tmux", "-2u", "set-window-option", "automatic-rename", "on")
	}

	// adjust tmux settings, if indicated
	tmuxType := os.Getenv("GOSSH_TMUX")
	isTmux := os.Getenv("TMUX")
	if tmuxType != "" && isTmux != "" {
		err := c.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func GetAuthentication(i connection.Item) string {
	var output string
	if i.Conn.PassFile != "" {
		pw := encryption.GetEncryptedContents(i.Conn.PassFile)
		if pw == "" {
			output = fmt.Sprintf("Password can be found in %v", i.Conn.PassFile)
		} else {
			output = fmt.Sprintf("Password is %v", strings.TrimSpace(pw))
		}
	} else if i.Conn.IdentityFile != "" {
		tempIdFile := encryption.GetEncryptedIdentity(i.Conn.IdentityFile)
		if tempIdFile != "" {
			output = fmt.Sprintf("Temporary identity file is %v (remove when done)", tempIdFile)
		} else {
			output = fmt.Sprintf("Identity file is %v", i.Conn.IdentityFile)
		}
	}

	return output
}

func main() {
	flag.Parse()

	if updateVersion {
		if err := updateExecutable(); err != nil {
			fmt.Printf("Could not update to latest version: %v\n", err)
			os.Exit(2)
		}
		os.Exit(0)
	}

	lm, err := menus.ConnectionList(initialFilter)
	if err != nil {
		log.Logger.Error("Error running program: ", err)
		os.Exit(1)
	}

	if lm.CheckedCount == 0 {
		fmt.Println("no items checked...nothing to do!")
		os.Exit(0)
	}

	connItems := lm.GetCheckedItems()

	switch lm.Action {
	case "ShowAuth":
		for _, val := range connItems {
			fmt.Printf("%v: %v\n", val.WindowName(), GetAuthentication(val))
		}
	case "Connect":
		c := connItems[0]
		if len(connItems) > 1 {
			fmt.Printf("Can only handle one connection but multiple selected.\n\t Connecting to %v...\n", c.WindowName())
		}

		if err := HandleTmux(c.WindowName()); err != nil {
			fmt.Printf("\nCould not rename tmux window: %v\n", err)
		}

		osCommand := []string{"ssh", "{{.FinalAddr}}"}
		out := runcommand.RunCommand(&c, osCommand, true)
		fmt.Println(out)

		if err := HandleTmux(""); err != nil {
			fmt.Printf("\nCould not reset tmux window: %v\n", err)
		}
	case "ReceiveFile":
		remoteSrc, dest, err := menus.SendReceive()

		if remoteSrc == "" || dest == "" || err != nil {
			break
		}

		destName := path.Clean(path.Join(dest, path.Base(remoteSrc)))
		osCommand := []string{"scp", "-rp", "{{.FinalAddr}}:" + remoteSrc, destName + "_{{.CleanTitle}}"}
		title := fmt.Sprintf("Copying %v on {{.WindowName}} to %v_{{.CleanTitle}}", remoteSrc, destName)
		runcommand.RunConcurrentCommandWithOutput(connItems, title, osCommand)

	case "SendFile":
		src, remoteDest, err := menus.SendReceive()

		if src == "" || remoteDest == "" || err != nil {
			break
		}

		osCommand := []string{"scp", "-rp", src, "{{.FinalAddr}}:" + remoteDest}
		title := fmt.Sprintf("Copying %v to %v on {{.WindowName}}", src, remoteDest)
		runcommand.RunConcurrentCommandWithOutput(connItems, title, osCommand)

	case "RunCommand":
		// get command to run
		cmdToRun, err := menus.CommandToRun()

		if cmdToRun == "" || err != nil {
			break
		}

		osCommand := []string{"ssh", "{{.FinalAddr}}"}
		osCommand = append(osCommand, strings.Split(cmdToRun, " ")...)
		title := fmt.Sprintf("running %v on {{.WindowName}}", cmdToRun)
		runcommand.RunConcurrentCommandWithOutput(connItems, title, osCommand)

	}

}
