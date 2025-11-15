package main

import (
	"fmt"
	"os"
	"os/exec"

	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nicknickel/gossh/internal/config"
	"github.com/nicknickel/gossh/internal/connection"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type sshFinishedMsg struct {
	err error
	msg string
}

type programCloseMsg struct{ err error }

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
			// get the selected item and assert type to item
			i := m.list.SelectedItem().(connection.Item)
			return m, runConnection(i)
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case sshFinishedMsg:
		return m, LogConn(msg)
	case programCloseMsg:
		if msg.err != nil {
			fmt.Printf("logging connection exited with error: %v\n", msg.err)
		}
		return m, tea.Quit
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
			return fmt.Errorf("could not rename tmux window: %e", err)
		}
	}

	return nil
}

func LogConn(msg sshFinishedMsg) tea.Cmd {
	errorMsg := ""
	if msg.err != nil {
		errorMsg = fmt.Sprintf("ssh exited with error: %v\n", msg.err)
	}

	c := exec.Command("echo", msg.msg, "\n", errorMsg)

	cmd := tea.ExecProcess(c, func(err error) tea.Msg {
		return programCloseMsg{err: err}
	})
	return cmd
}

func runConnection(i connection.Item) tea.Cmd {

	var connCmd *exec.Cmd
	windowName := i.Name

	if i.Name != i.Conn.Address {
		windowName = windowName + " (" + i.Conn.Address + ")"
	}
	if i.Conn.User != "" {
		windowName = i.Conn.User + "@" + windowName
	}

	HandleTmux(windowName)

	sshPassPath, err := exec.LookPath("sshpass")
	if err != nil {
		fmt.Printf("sshpass not found")
	}

	// determine correct program to run
	if i.Conn.PassFile != "" && sshPassPath != "" {
		connCmd = exec.Command("sshpass", "-f", i.Conn.PassFile, "ssh", "-o", "ServerAliveInterval=30", i.FinalAddr())
	} else if i.Conn.IdentityFile != "" {
		connCmd = exec.Command("ssh", "-o", "ServerAliveInterval=30", "-i", i.Conn.IdentityFile, i.FinalAddr())
	} else {
		connCmd = exec.Command("ssh", "-o", "ServerAliveInterval=30", i.FinalAddr())
	}

	cmd := tea.ExecProcess(connCmd, func(err error) tea.Msg {
		HandleTmux("")
		return sshFinishedMsg{err: err, msg: strings.Join(connCmd.Args, " ")}
	})
	return cmd

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

func main() {

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

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

}
