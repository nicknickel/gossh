package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nicknickel/gossh/internal/config"
	"github.com/nicknickel/gossh/internal/connection"
	"github.com/nicknickel/gossh/internal/encryption"
	"github.com/nicknickel/gossh/internal/log"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "enter" && m.list.FilterState() != list.Filtering {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
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

func RunConnection(i connection.Item) string {

	var connCmd *exec.Cmd

	sshPassPath, err := exec.LookPath("sshpass")
	if err != nil {
		log.Logger.Warn("sshpass not found")
	}

	// determine correct program to run
	if i.Conn.PassFile != "" && sshPassPath != "" {
		pw := encryption.GetEncryptedContents(i.Conn.PassFile)
		if pw == "" {
			connCmd = exec.Command("sshpass", "-f", i.Conn.PassFile, "ssh", "-o", "ServerAliveInterval=30", i.FinalAddr())
		} else {
			connCmd = exec.Command("sshpass", "-e", "ssh", "-o", "ServerAliveInterval=30", i.FinalAddr())
			connCmd.Env = append(connCmd.Environ(), "SSHPASS="+pw)
		}
	} else if i.Conn.IdentityFile != "" {
		tempIdFile := encryption.GetEncryptedIdentity(i.Conn.IdentityFile)
		if tempIdFile != "" {
			connCmd = exec.Command("ssh", "-o", "ServerAliveInterval=30", "-i", tempIdFile, i.FinalAddr())
			defer os.Remove(tempIdFile)
		} else {
			connCmd = exec.Command("ssh", "-o", "ServerAliveInterval=30", "-i", i.Conn.IdentityFile, i.FinalAddr())
		}
	} else {
		connCmd = exec.Command("ssh", "-o", "ServerAliveInterval=30", i.FinalAddr())
	}

	connCmd.Stdin = os.Stdin
	connCmd.Stdout = os.Stdout
	connCmd.Stderr = os.Stderr
	connCmd.Run()
	return fmt.Sprintf("\n%v\n", strings.Join(connCmd.Args, " "))
}

func FilterFunc(t string, items []string) []list.Rank {
	var results []list.Rank
	terms := strings.Split(t, " ")

	for i, item := range items {
		termsMatched := 0
		// want to make sure all space separated search words
		// are found in one of the fields
		for _, term := range terms {
			// Splitting the FilterValue as it is space separated by
			// i.Name + " " + i.Conn.Address + " " + i.Conn.User + " " + i.Conn.Description
			searchFields := strings.SplitN(item, " ", 4)
			for _, field := range searchFields {
				if index := strings.Index(strings.ToLower(field), strings.ToLower(term)); index > -1 {
					termsMatched++
					break // term exists in one of the fields so don't need to keep looking
				}
			}

		}

		if termsMatched == len(terms) {
			results = append(results, list.Rank{Index: i, MatchedIndexes: nil})
		}
	}

	return results
}

func OutputAuthentication(i connection.Item) string {
	var output string
	if i.Conn.PassFile != "" {
		pw := encryption.GetEncryptedContents(i.Conn.PassFile)
		if pw == "" {
			output = i.Conn.PassFile
		} else {
			output = pw
		}
	} else if i.Conn.IdentityFile != "" {
		tempIdFile := encryption.GetEncryptedIdentity(i.Conn.IdentityFile)
		if tempIdFile != "" {
			output = tempIdFile
		} else {
			output = i.Conn.IdentityFile
		}
	}

	return output
}

func main() {
	log.Init()

	items := config.ReadConnections()
	l := list.NewDefaultDelegate()
	l.Styles.SelectedTitle = l.Styles.SelectedTitle.
		BorderForeground(lipgloss.Color("#06bf18")).
		Foreground(lipgloss.Color("#06bf18"))
	l.Styles.SelectedDesc = l.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#06bf18")).
		BorderForeground(lipgloss.Color("#06bf18"))

	m := model{list: list.New(items, l, 0, 0)}
	m.list.Title = "Go SSH Connection Manager"
	m.list.Styles.Title = lipgloss.NewStyle().Background(lipgloss.Color("#045edb")).Padding(0, 1)
	m.list.FilterInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))
	m.list.Filter = FilterFunc

	p := tea.NewProgram(m, tea.WithAltScreen())

	fm, err := p.Run()
	if err != nil {
		log.Logger.Error("Error running program: ", err)
		os.Exit(1)
	}

	lm := fm.(model)
	c := lm.list.SelectedItem().(connection.Item)

	outputAuth := flag.Bool("o", false, "Decrypt and output authentication for selected connection")
	flag.Parse()

	var output string
	if *outputAuth {
		output = OutputAuthentication(c)
	} else {
		if err := HandleTmux(c.WindowName()); err != nil {
			fmt.Printf("\nCould not rename tmux window: %v\n", err)
		}

		output = RunConnection(c)

		if err := HandleTmux(""); err != nil {
			fmt.Printf("\nCould not reset tmux window: %v\n", err)
		}
	}

	fmt.Println(output)

}
