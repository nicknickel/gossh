package main

import (
	"fmt"
	"os"
	"os/exec"
	"slices"

	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nicknickel/gossh/internal/config"
	"github.com/nicknickel/gossh/internal/connection"
	"github.com/nicknickel/gossh/internal/encryption"
	"github.com/nicknickel/gossh/internal/log"
	"github.com/nicknickel/gossh/internal/runcommand"
	"golang.org/x/term"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type model struct {
	list         list.Model
	checkedCount int
	action       string
}

type customKeyMap struct {
	Choose     key.Binding
	Select     key.Binding
	SelectAll  key.Binding
	ShowAuth   key.Binding
	RunCommand key.Binding
}

func (c *customKeyMap) AdditionalKeys() []key.Binding {
	return []key.Binding{c.Choose, c.Select, c.SelectAll, c.ShowAuth, c.RunCommand}
}

var customKeyBindings = customKeyMap{
	Choose: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "choose"),
	),
	Select: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("<space>", "select"),
	),
	SelectAll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "select-all"),
	),
	ShowAuth: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "show-auth"),
	),
	RunCommand: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "run-command"),
	),
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
		if key.Matches(msg, customKeyBindings.ShowAuth) && m.list.FilterState() != list.Filtering {
			m.action = "ShowAuth"
			if m.checkedCount == 0 {
				i := m.list.SelectedItem().(connection.Item)
				i.Checked = true
				m.checkedCount++
				m.list.SetItem(m.list.GlobalIndex(), i)
			}
			return m, tea.Quit
		}
		if key.Matches(msg, customKeyBindings.Choose) && m.list.FilterState() != list.Filtering {
			m.action = "Connect"
			if m.checkedCount == 0 {
				i := m.list.SelectedItem().(connection.Item)
				i.Checked = true
				m.checkedCount++
				m.list.SetItem(m.list.GlobalIndex(), i)
			}
			return m, tea.Quit
		}
		if key.Matches(msg, customKeyBindings.Select) && m.list.FilterState() != list.Filtering {
			i := m.list.SelectedItem().(connection.Item)
			if i.Checked {
				i.Checked = false
				m.checkedCount--
			} else {
				i.Checked = true
				m.checkedCount++
			}
			m.list.SetItem(m.list.GlobalIndex(), i)
		}
		if key.Matches(msg, customKeyBindings.SelectAll) && m.list.FilterState() != list.Filtering {
			// (un)select all; honor filter
			var checkAll bool
			if m.checkedCount == 0 {
				checkAll = true
			} else {
				checkAll = false
			}
			for _, val := range m.list.VisibleItems() {
				connItem := val.(connection.Item)
				if !connItem.Checked && checkAll {
					connItem.Checked = true
					m.checkedCount++
				} else if connItem.Checked && !checkAll {
					connItem.Checked = false
					m.checkedCount--
				}
				m.list.SetItem(connItem.Index, connItem)
			}
			// the below is required for some reason
			// without it the list does re-render to show the x marks
			// but only when filtering
			fv := m.list.FilterValue()
			if fv != "" {
				m.list.SetFilterText(fv)
			}
		}
		if key.Matches(msg, customKeyBindings.RunCommand) && m.list.FilterState() != list.Filtering {
			m.action = "RunCommand"
			if m.checkedCount == 0 {
				i := m.list.SelectedItem().(connection.Item)
				i.Checked = true
				m.checkedCount++
				m.list.SetItem(m.list.GlobalIndex(), i)
			}
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

func RunCommand(i connection.Item, c string) string {
	var env []string
	command := []string{"ssh", "-o", "ServerAliveInterval=30"}

	sshPassPath, err := exec.LookPath("sshpass")
	if err != nil {
		log.Logger.Warn("sshpass not found")
	}

	if i.Conn.PassFile != "" && sshPassPath != "" {
		command = slices.Insert(command, 0, "sshpass")

		pw := encryption.GetEncryptedContents(i.Conn.PassFile)
		if pw == "" {
			command = slices.Insert(command, 1, []string{"-f", i.Conn.PassFile}...)
		} else {
			command = slices.Insert(command, 1, "-e")
			env = append(env, "SSHPASS="+pw)
		}
	} else if i.Conn.IdentityFile != "" {
		command = append(command, "-i")
		tempIdFile := encryption.GetEncryptedIdentity(i.Conn.IdentityFile)
		if tempIdFile != "" {
			command = append(command, tempIdFile)
			defer os.Remove(tempIdFile)
		} else {
			command = append(command, i.Conn.IdentityFile)
		}
	}

	command = append(command, i.FinalAddr())
	if c != "" {
		command = append(command, strings.Split(c, " ")...)
	}

	cmd := exec.Command(command[0], command[1:]...)
	for _, val := range env {
		cmd.Env = append(cmd.Env, val)
	}

	if c == "" {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		return fmt.Sprintf("\n%v\n", strings.Join(cmd.Args, " "))
	} else {
		out, err := cmd.Output()
		if err != nil {
			return fmt.Sprintf("%s: %s", err.Error(), out)
		}
		return string(out)
	}
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

func GetCheckedItems(m model) []connection.Item {
	var connItems []connection.Item
	for _, val := range m.list.Items() {
		connItem := val.(connection.Item)
		if connItem.Checked {
			connItems = append(connItems, connItem)
		}
	}

	return connItems
}

func initModel(items []list.Item) model {
	l := list.NewDefaultDelegate()
	l.Styles.SelectedTitle = l.Styles.SelectedTitle.
		BorderForeground(lipgloss.Color("#06bf18")).
		Foreground(lipgloss.Color("#06bf18"))
	l.Styles.SelectedDesc = l.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#06bf18")).
		BorderForeground(lipgloss.Color("#06bf18"))

	m := model{
		list: list.New(items, l, 0, 0),
	}
	m.list.Title = "Go SSH Connection Manager"
	m.list.Styles.Title = lipgloss.NewStyle().Background(lipgloss.Color("#045edb")).Padding(0, 1)
	m.list.FilterInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))
	m.list.Filter = FilterFunc
	m.list.AdditionalShortHelpKeys = customKeyBindings.AdditionalKeys
	m.list.AdditionalFullHelpKeys = customKeyBindings.AdditionalKeys

	return m
}

func main() {
	log.Init()

	items := config.ReadConnections()
	m := initModel(items)
	p := tea.NewProgram(m, tea.WithAltScreen())

	fm, err := p.Run()
	if err != nil {
		log.Logger.Error("Error running program: ", err)
		os.Exit(1)
	}

	lm := fm.(model)

	if lm.checkedCount > 0 {
		var connItems []connection.Item = GetCheckedItems(lm)

		switch lm.action {
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

			fmt.Println(RunCommand(c, ""))

			if err := HandleTmux(""); err != nil {
				fmt.Printf("\nCould not reset tmux window: %v\n", err)
			}
		case "RunCommand":
			// get command to run
			cmdToRun, err := runcommand.Get()

			if cmdToRun != "" && err == nil {
				fd := int(os.Stdout.Fd())
				width, _, err := term.GetSize(fd)
				if err != nil {
					width = 100
				}
				style := lipgloss.NewStyle().
					BorderStyle(lipgloss.NormalBorder()).
					Padding(1, 1, 1, 1).
					BorderForeground(lipgloss.Color("228")).
					Width(width - 10)
				for _, val := range connItems {
					title := fmt.Sprintf("Running '%v' on %v", cmdToRun, val.WindowName())
					fmt.Println(title)
					out := RunCommand(val, cmdToRun)
					fmt.Println(style.Render(out))
				}
			}
		}

	}
}
