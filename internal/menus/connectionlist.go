package menus

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nicknickel/gossh/internal/config"
	"github.com/nicknickel/gossh/internal/connection"
)

type connectionlistModel struct {
	list         list.Model
	CheckedCount int
	Action       string
}

func (m connectionlistModel) Init() tea.Cmd {
	return nil
}

func (m connectionlistModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.list.FilterState() == list.Filtering {
			break
		}
		if key.Matches(msg, connectionListKeyBindings.ShowAuth) {
			m.Action = "ShowAuth"
			if m.CheckedCount == 0 {
				i := m.list.SelectedItem().(connection.Item)
				i.Checked = true
				m.CheckedCount++
				m.list.SetItem(m.list.GlobalIndex(), i)
			}
			return m, tea.Quit
		}
		if key.Matches(msg, connectionListKeyBindings.Choose) {
			m.Action = "Connect"
			if m.CheckedCount == 0 {
				i := m.list.SelectedItem().(connection.Item)
				i.Checked = true
				m.CheckedCount++
				m.list.SetItem(m.list.GlobalIndex(), i)
			}
			return m, tea.Quit
		}
		if key.Matches(msg, connectionListKeyBindings.Select) {
			i := m.list.SelectedItem().(connection.Item)
			if i.Checked {
				i.Checked = false
				m.CheckedCount--
			} else {
				i.Checked = true
				m.CheckedCount++
			}
			m.list.SetItem(m.list.GlobalIndex(), i)
			fv := m.list.FilterValue()
			if fv != "" {
				m.list.SetFilterText(fv)
			}
		}
		if key.Matches(msg, connectionListKeyBindings.SelectAll) {
			// (un)select all; honor filter
			var checkAll bool
			if m.CheckedCount == 0 {
				checkAll = true
			} else {
				checkAll = false
			}
			for _, val := range m.list.VisibleItems() {
				connItem := val.(connection.Item)
				if !connItem.Checked && checkAll {
					connItem.Checked = true
					m.CheckedCount++
				} else if connItem.Checked && !checkAll {
					connItem.Checked = false
					m.CheckedCount--
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
		if key.Matches(msg, connectionListKeyBindings.RunCommand) {
			m.Action = "RunCommand"
			if m.CheckedCount == 0 {
				i := m.list.SelectedItem().(connection.Item)
				i.Checked = true
				m.CheckedCount++
				m.list.SetItem(m.list.GlobalIndex(), i)
			}
			return m, tea.Quit
		}
		if key.Matches(msg, connectionListKeyBindings.SendFile) {
			m.Action = "SendFile"
			if m.CheckedCount == 0 {
				i := m.list.SelectedItem().(connection.Item)
				i.Checked = true
				m.CheckedCount++
				m.list.SetItem(m.list.GlobalIndex(), i)
			}
			return m, tea.Quit
		}
		if key.Matches(msg, connectionListKeyBindings.ReceiveFile) {
			m.Action = "ReceiveFile"
			if m.CheckedCount == 0 {
				i := m.list.SelectedItem().(connection.Item)
				i.Checked = true
				m.CheckedCount++
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

func (m connectionlistModel) View() string {
	return globalStyle(m.list.View())
}

func (m connectionlistModel) GetCheckedItems() []connection.Item {
	var connItems []connection.Item
	for _, val := range m.list.Items() {
		connItem := val.(connection.Item)
		if connItem.Checked {
			connItems = append(connItems, connItem)
		}
	}

	return connItems
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

type connectionListKeyMap struct {
	Choose      key.Binding
	Select      key.Binding
	SelectAll   key.Binding
	ShowAuth    key.Binding
	RunCommand  key.Binding
	SendFile    key.Binding
	ReceiveFile key.Binding
}

func (c *connectionListKeyMap) AdditionalKeys() []key.Binding {
	return []key.Binding{c.Choose, c.Select, c.SelectAll, c.ShowAuth, c.RunCommand, c.SendFile, c.ReceiveFile}
}

var connectionListKeyBindings = connectionListKeyMap{
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
	SendFile: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "send-file"),
	),
	ReceiveFile: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "receive-file"),
	),
}

func newConnectionlistModel(items []list.Item) connectionlistModel {
	l := list.NewDefaultDelegate()
	l.Styles.SelectedTitle = l.Styles.SelectedTitle.
		BorderForeground(lipgloss.Color("#06bf18")).
		Foreground(lipgloss.Color("#06bf18"))
	l.Styles.SelectedDesc = l.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#06bf18")).
		BorderForeground(lipgloss.Color("#06bf18"))

	m := connectionlistModel{
		list: list.New(items, l, 0, 0),
	}
	m.list.Title = "Go SSH Connection Manager"
	m.list.Styles.Title = lipgloss.NewStyle().Background(lipgloss.Color("#045edb")).Padding(0, 1)
	m.list.FilterInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))
	m.list.Filter = FilterFunc
	m.list.AdditionalShortHelpKeys = connectionListKeyBindings.AdditionalKeys
	m.list.AdditionalFullHelpKeys = connectionListKeyBindings.AdditionalKeys

	return m
}

func ConnectionList(initialFilter string) (*connectionlistModel, error) {
	items := config.ReadConnections()
	m := newConnectionlistModel(items)
	if initialFilter != "" {
		m.list.SetFilterText(initialFilter)
	}
	p := tea.NewProgram(m, tea.WithAltScreen())

	fm, err := p.Run()
	if err != nil {
		return nil, err
	}

	modelFm := fm.(connectionlistModel)

	return &modelFm, nil
}
