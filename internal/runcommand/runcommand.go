package runcommand

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nicknickel/gossh/internal/connection"
	"github.com/nicknickel/gossh/internal/encryption"
	"github.com/nicknickel/gossh/internal/log"
	"golang.org/x/term"
	"sync"
	"text/template"
)

// GetPasswordTemplate returns a slice of strings containing the sshpass
// command template, another slice of strings containing the environment
// variables necessary for the command to work, and an error
func GetPasswordTemplate(i *connection.Item) ([]string, []string, error) {
	sshPassPath, err := exec.LookPath("sshpass")
	if err != nil || sshPassPath == "" {
		log.Logger.Warn("sshpass not found")
		return []string{}, []string{}, errors.New("sshpass not found")
	}

	if i.Conn.PassFile == "" {
		return []string{}, []string{}, errors.New("passfile parameter not defined")
	}

	pw := encryption.GetEncryptedContents(i.Conn.PassFile)
	if pw == "" {
		return []string{"sshpass", "-f", "{{.Conn.PassFile}}"}, nil, nil
	} else {
		return []string{"sshpass", "-e"}, []string{"SSHPASS=" + pw}, nil
	}
}

func GetIdentityTemplate(i *connection.Item) ([]string, bool, error) {
	if i.Conn.IdentityFile == "" {
		return []string{}, false, errors.New("No identify file indicated")
	}

	tempIdFile := encryption.GetEncryptedIdentity(i.Conn.IdentityFile)
	if tempIdFile != "" {
		return []string{"-i", tempIdFile}, true, nil
	}

	return []string{"-i", i.Conn.IdentityFile}, false, nil
}

func RenderTemplateSlice(s *[]string, i connection.Item) []string {
	joined := strings.Join(*s, " ")

	t1 := template.New("render")
	t1, _ = t1.Parse(joined)
	// Subjectively don't need to handle error due to
	// template being defined in code

	var buf bytes.Buffer
	t1.Execute(&buf, i)

	joined = buf.String()
	return strings.Split(joined, " ")
}

func CreateCommand(c *[]string, e *[]string, i connection.Item) *exec.Cmd {
	cText := RenderTemplateSlice(c, i)
	eText := RenderTemplateSlice(e, i)

	cmd := exec.Command(cText[0], cText[1:]...)
	for _, val := range eText {
		cmd.Env = append(cmd.Env, val)
	}
	return cmd
}

func GetEnv() []string {
	env := []string{"TERM=" + os.Getenv("TERM")}
	return env
}

func RunAttachedCommand(cmd *exec.Cmd) string {
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
	return fmt.Sprintf("\n%v\n", strings.Join(cmd.Args, " "))
}

func RunCommandWithOutput(cmd *exec.Cmd) string {
	outerr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("%s: %s", err.Error(), outerr)
	}
	if string(outerr) == "" {
		return "Success!"
	}
	return string(outerr)
}

func RunCommand(i *connection.Item, c []string, a bool) string {
	env := GetEnv()

	passTemplate, passEnv, err := GetPasswordTemplate(i)
	if err == nil {
		c = slices.Insert(c, 0, passTemplate...)
		if passEnv != nil {
			env = append(env, passEnv...)
		}
	} else {
		idTemplate, cleanup, err := GetIdentityTemplate(i)
		if err == nil {
			c = slices.Insert(c, 1, idTemplate...)
			if cleanup {
				defer os.Remove(idTemplate[1])
			}
		}
	}
	cmd := CreateCommand(&c, &env, *i)

	var out string
	if a {
		out = RunAttachedCommand(cmd)
	} else {
		out = RunCommandWithOutput(cmd)
	}
	return out
}

func RunConcurrentCommandWithOutput(items []connection.Item, title string, c []string) {
	width := GetTermWidth()

	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		Padding(0, 1).
		BorderForeground(lipgloss.Color("228")).
		Width(width - 10)

	var wg sync.WaitGroup
	maxConcurrent := 5
	concurrentEnv := os.Getenv("GOSSH_CONCURRENCY")
	if concurrentEnv != "" {
		maxConcurrent, _ = strconv.Atoi(concurrentEnv)
	}
	limiter := make(chan int, maxConcurrent)

	t1 := template.New("title")
	t1, _ = t1.Parse(title)
	// Subjectively unecessary to handle error due to
	// template being defined in code

	for _, item := range items {
		wg.Go(func() {
			limiter <- 1
			out := RunCommand(&item, c, false)
			output := fmt.Sprintf("\n%v\n", style.Render(out))
			t1.Execute(os.Stdout, item)
			fmt.Println(output)
			<-limiter
		})
	}
	wg.Wait()
}

func GetTermWidth() int {
	fd := int(os.Stdout.Fd())
	width, _, err := term.GetSize(fd)
	if err != nil {
		return 100
	}
	return width
}
