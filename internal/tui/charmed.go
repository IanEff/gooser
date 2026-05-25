package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/ianeff/gooser/internal/gooser"
)

type model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

func initialModel(apps ...string) model {
	return model{
		choices:  apps,
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	// From the docs: 'Just return `nil`, which means "no I/O right now, please"'
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", "space":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	s := "Select an app to GOOSE\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}
	s += "\nPress q to quit.\n"

	return tea.NewView(s)
}

func Run(apps []gooser.Application) (string, error) {
	names := make([]string, len(apps))
	for i, a := range apps {
		names[i] = a.Name
	}
	m := initialModel(names...)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	result := finalModel.(model)

	for idx := range result.selected {
		return result.choices[idx], nil
	}
	return "", fmt.Errorf("no app selected")
}
