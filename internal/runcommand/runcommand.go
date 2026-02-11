package runcommand

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func Get() (string, error) {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if m, err := p.Run(); err != nil {
		return "", err
	} else {
		model := m.(model)
		return model.textInput.Value(), nil
	}
}

type (
	errMsg error
)

type model struct {
	textInput textinput.Model
	err       error
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "command to run"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return model{
		textInput: ti,
		err:       nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func StyleTitle(t string) string {
	s := lipgloss.NewStyle().Background(lipgloss.Color("#045edb")).Padding(0, 1)
	return s.Render(t)
}

func (m model) View() string {
	title := StyleTitle("What command should be executed?")
	return fmt.Sprintf(
		"%v\n\n%s\n\n%s",
		title,
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}
