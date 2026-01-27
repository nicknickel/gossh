package main

import (
	"fmt"
	"os"
	"os/exec"

	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tmux "github.com/jubnzv/go-tmux"
	"github.com/nicknickel/gossh/internal/config"
	"github.com/nicknickel/gossh/internal/connection"
	"github.com/nicknickel/gossh/internal/encryption"
	"github.com/nicknickel/gossh/internal/log"
	"github.com/riywo/loginshell"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type sshFinishedMsg struct {
	err error
	msg string
}

type programCloseMsg struct{ err error }

type model struct {
	list         list.Model
	checkedCount int
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
			// c := []connection.Item{}
			// // If items are checked, ignore selected item
			// if m.checkedCount > 0 {
			// 	for _, listItem := range m.list.Items() {
			// 		if listItem.(connection.Item).Checked {
			// 			c = append(c, listItem.(connection.Item))
			// 		}
			// 	}
			// } else {
			// 	// get the selected item and assert type to item
			// 	c = append(c, m.list.SelectedItem().(connection.Item))
			// }
			// return m, runConnections(c)
		}
		if msg.String() == " " && m.list.FilterState() != list.Filtering {
			i := m.list.SelectedItem().(connection.Item)
			if i.Checked {
				i.Checked = false
				i.Name = strings.Replace(i.Name, "[✓] ", "", 1)
				m.checkedCount--
			} else {
				i.Checked = true
				i.Name = fmt.Sprintf("[✓] %v", i.Name)
				m.checkedCount++
			}
			m.list.SetItem(m.list.GlobalIndex(), i)
			return m, nil
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case sshFinishedMsg:
		return m, LogConn(msg)
	case programCloseMsg:
		if msg.err != nil {
			log.Logger.Error("logging connection exited with error", "err", msg.err)
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

func LogConn(msg sshFinishedMsg) tea.Cmd {
	errorMsg := ""
	if msg.err != nil {
		log.Logger.Error("ssh exited with error", "err", msg.err)
		errorMsg = fmt.Sprintf("ssh exited with error: %v\n", msg.err)
	} else {
		log.Logger.Info("ssh connection finished", "cmd", msg.msg)
	}

	c := exec.Command("echo", msg.msg, "\n", errorMsg)

	cmd := tea.ExecProcess(c, func(err error) tea.Msg {
		return programCloseMsg{err: err}
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

func GetCommand(i connection.Item, sshPass bool) *exec.Cmd {
	var connCmd *exec.Cmd

	// determine correct program to run
	if i.Conn.PassFile != "" && sshPass {
		pw := encryption.GetEncryptedPassword(i.Conn.PassFile)
		if pw == "" {
			connCmd = exec.Command("sshpass", "-f", i.Conn.PassFile, "ssh", "-o", "ServerAliveInterval=30", i.FinalAddr())
		} else {
			// connCmd = exec.Command("sshpass", "-e", "ssh", "-o", "ServerAliveInterval=30", i.FinalAddr())
			connCmd = exec.Command("export", "SSHPASS="+pw, "sshpass", "ssh", "-o", "ServerAliveInterval=30", i.FinalAddr())
			// connCmd.Env = append(connCmd.Environ(), "SSHPASS="+pw)
		}
	} else if i.Conn.IdentityFile != "" {
		connCmd = exec.Command("ssh", "-o", "ServerAliveInterval=30", "-i", i.Conn.IdentityFile, i.FinalAddr())
	} else {
		connCmd = exec.Command("ssh", "-o", "ServerAliveInterval=30", i.FinalAddr())
	}

	return connCmd
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
	if strings.ToLower(tmuxType) == "true" && tmux.IsInsideTmux() {
		err := c.Run()
		if err != nil {
			// log.Logger.Errorf("could not rename tmux window: %e", err)
			return fmt.Errorf("could not rename tmux window: %e", err)
		}
	}

	return nil
}

// func runConnection(c []connection.Item) tea.Cmd {
// 	t := runConnection(c[0])
// 	return t
// }

func GetPanes() ([]string, error) {
	listPanes := exec.Command("tmux", "list-panes", "-F", "#{pane_id}")
	panes, err := listPanes.Output()
	if err != nil {
		fmt.Println("unable to get tmux panes")
		return nil, err
	}
	p := strings.Split(string(panes), "\n")
	return p, nil
}
func runConnections(c []connection.Item) error {
	tmuxType := os.Getenv("GOSSH_TMUX")
	sshPass := false
	_, err := exec.LookPath("sshpass")
	if err == nil {
		sshPass = true
	}
	if strings.ToLower(tmuxType) == "true" && tmux.IsInsideTmux() {
		// p, _ := GetPanes()
		// if len(c) > len(p) {
		// 	diff := len(c) - len(p)
		// 	for range diff {
		// 		newPaneCmd := exec.Command("tmux", "split-window", "-d", strings.Join(sshCmd.Args, " "))
		// 		newPaneCmd.Run()
		// 	}
		// }

		// tileCmd := exec.Command("tmux", "select-layout", "tiled")
		// tileCmd.Run()

		// p, _ = GetPanes()
		// fmt.Println(len(p))

		for _, conn := range c {
			sshCmd := GetCommand(conn, sshPass)
			shell, err := loginshell.Shell()
			if err != nil {
				shell = "/bin/bash"
			}
			newPaneCmd := exec.Command("tmux", "split-window", "-d", "\"", strings.Join(sshCmd.Args, " "), ";", shell, "\"")
			fmt.Println(newPaneCmd.Args)
			newPaneCmd.Run()
			// cmd := exec.Command("tmux", "send-keys", "-t", fmt.Sprintf("%v", p[i]), strings.Join(sshCmd.Args, " "), "C-m")
			// stdout, _ := cmd.StdoutPipe()
			// cmd.Start()

		}
		tileCmd := exec.Command("tmux", "select-layout", "tiled")
		tileCmd.Run()
	}

	return nil
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

	m := model{list: list.New(items, l, 0, 0), checkedCount: 0}
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
	c := []connection.Item{}
	// If items are checked, ignore selected item
	if lm.checkedCount > 0 {
		for _, listItem := range lm.list.Items() {
			if listItem.(connection.Item).Checked {
				c = append(c, listItem.(connection.Item))
			}
		}
	} else {
		// get the selected item and assert type to item
		c = append(c, lm.list.SelectedItem().(connection.Item))
	}

	err = runConnections(c)
	if err != nil {
		fmt.Println("log something here")
	}
}
