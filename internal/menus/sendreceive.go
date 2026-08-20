package menus

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type sendreceiveModel struct {
	focusedInput int
	srcInput     textinput.Model
	destInput    textinput.Model
	err          error
}

func (m sendreceiveModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m sendreceiveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// switch input focus if we don't have all the data
			if m.srcInput.Value() == "" {
				m.focusedInput = 1
				m.srcInput.Focus()
				m.destInput.Blur()
			} else if m.destInput.Value() == "" {
				m.focusedInput = 0
				m.destInput.Focus()
				m.srcInput.Blur()
			} else {
				return m, tea.Quit
			}
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "tab", "shift+tab", "down", "up":
			switch m.focusedInput {
			case 0:
				m.focusedInput = 1
				m.destInput.Focus()
				m.srcInput.Blur()
			case 1:
				m.focusedInput = 0
				m.srcInput.Focus()
				m.destInput.Blur()
			}
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.srcInput, cmd = m.srcInput.Update(msg)
	m.destInput, cmd = m.destInput.Update(msg)
	return m, cmd
}

func (m sendreceiveModel) View() string {
	title := StyleTitle("File Details")
	return globalStyle(fmt.Sprintf(
		"%v\n\n%s\n\n%s\n\n%s\n\n%s\n\n%s",
		title,
		"File(s) to copy:",
		m.srcInput.View(),
		"Place to copy them to:",
		m.destInput.View(),
		"(esc to quit)",
	) + "\n")
}

func newSendreceiveModel() sendreceiveModel {
	src := textinput.New()
	src.Placeholder = "/file/to/copy"
	src.CharLimit = 156
	src.Width = 20
	src.Focus()

	dest := textinput.New()
	dest.Placeholder = "/place/to/copy/to"
	dest.CharLimit = 156
	dest.Width = 20

	return sendreceiveModel{
		focusedInput: 0,
		srcInput:     src,
		destInput:    dest,
		err:          nil,
	}
}

func SendReceive() (string, string, error) {
	p := tea.NewProgram(newSendreceiveModel(), tea.WithAltScreen())
	if m, err := p.Run(); err != nil {
		return "", "", err
	} else {
		model := m.(sendreceiveModel)
		return model.srcInput.Value(), model.destInput.Value(), nil
	}
}
