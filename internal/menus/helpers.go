package menus

import "github.com/charmbracelet/lipgloss"

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func StyleTitle(t string) string {
	s := lipgloss.NewStyle().Background(lipgloss.Color("#045edb")).Padding(0, 1)
	return s.Render(t)
}

type (
	errMsg error
)

func globalStyle(s string) string {
	return docStyle.Render(s)
}
