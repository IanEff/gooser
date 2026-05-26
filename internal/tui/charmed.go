// Package tui provides a superslick charmbracelet-based TUI for the gooser's pleasure.
package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/ianeff/gooser/internal/gooser"
)

// Action represents the action to perform on an Application.
type Action int

// These Actions are the verbs available to the user.
const (
	ActionGoose Action = iota
	ActionTwiddleOn
	ActionTwiddleOff
)

// Result holds the result of a TUI selection.
type Result struct {
	Application string
	Action      Action
}

type model struct {
	apps   []gooser.Application
	cursor int
	result Result
}

func initialModel(apps []gooser.Application) model {
	return model{
		apps: apps,
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
			if m.cursor < len(m.apps)-1 {
				m.cursor++
			}
		case "enter", "g":
			m.result.Application = m.apps[m.cursor].Name
			m.result.Action = ActionGoose
			return m, tea.Quit

		case "t":
			m.result.Application = m.apps[m.cursor].Name
			m.result.Action = ActionTwiddleOn
			return m, tea.Quit
		case "o":
			m.result.Application = m.apps[m.cursor].Name
			m.result.Action = ActionTwiddleOff
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	s := "🪿 Gooser — ArgoCD applications\n\n"

	for i, app := range m.apps {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %-40s sync=%-12s health=%s\n", cursor, app.Name, app.Sync, app.Health)
	}
	s += "\nSelect an Application to [g]oose, [t]widdle syncPolicy on, twiddle the syncPolicy [o]ff for maintenance mode, or [q]uit.\n"

	return tea.NewView(s)
}

// Run starts the TUI and returns the Result of the user's selection.
func Run(apps []gooser.Application) (Result, error) {
	m := initialModel(apps)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return Result{}, err
	}

	final := finalModel.(model)
	return Result{
		Application: final.result.Application,
		Action:      final.result.Action,
	}, nil
}
