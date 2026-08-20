package menus

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type commandtorunModel struct {
	textInput textinput.Model
	err       error
}

func (m commandtorunModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m commandtorunModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m commandtorunModel) View() string {
	title := StyleTitle("What command should be executed?")
	return globalStyle(fmt.Sprintf(
		"%v\n\n%s\n\n%s",
		title,
		m.textInput.View(),
		"(esc to quit)",
	) + "\n")
}

func newCommandtorunModel() commandtorunModel {
	ti := textinput.New()
	ti.Placeholder = "command to run"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return commandtorunModel{
		textInput: ti,
		err:       nil,
	}
}

func CommandToRun() (string, error) {
	p := tea.NewProgram(newCommandtorunModel(), tea.WithAltScreen())
	if m, err := p.Run(); err != nil {
		return "", err
	} else {
		model := m.(commandtorunModel)
		return model.textInput.Value(), nil
	}
}
