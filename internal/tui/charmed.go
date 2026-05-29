// Package tui provides a superslick charmbracelet-based TUI for the gooser's pleasure.
package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/ianeff/gooser/internal/gooser"
)

// Action represents the action to perform on an Application.
type Action int

// These Actions are the verbs available to the user.
const (
	ActionGoose Action = iota
	ActionTwiddle
)

// Result holds the result of a TUI selection.
type Result struct {
	Application string
	Action      Action
}

type model struct {
	apps     []gooser.Application
	cursor   int
	result   Result
	quitting bool
}

func initialModel(apps []gooser.Application) model {
	return model{
		apps: apps,
	}
}

// Init is called when the TUI is first initialized.
func (m model) Init() tea.Cmd {
	// From the docs: 'Just return `nil`, which means "no I/O right now, please"'
	return nil
}

// Update handles incoming messages from the TUI.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
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
			m.quitting = true
			return m, tea.Quit

		case "t":
			m.result.Application = m.apps[m.cursor].Name
			m.result.Action = ActionTwiddle
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the TUI view.
func (m model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	var sb strings.Builder

	sb.WriteString(titleStyle.Render("🪿 Gooser — ArgoCD applications"))
	sb.WriteString("\n")

	for i, app := range m.apps {
		cursor := "  "
		if m.cursor == i {
			cursor = cursorStyle.Render("❯") + " "
		}

		var name string
		if m.cursor == i {
			name = selectedNameStyle.Render(app.Name)
		} else {
			name = nameStyle.Render(app.Name)
		}

		sb.WriteString(cursor)
		sb.WriteString(name)
		sb.WriteString(labelStyle.Render("sync="))
		sb.WriteString(syncCell(app.Sync))
		sb.WriteString("  ")
		sb.WriteString(labelStyle.Render("health="))
		sb.WriteString(healthCell(app.Health))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(
		hintStyle.Render("Select an application to [") +
			keyStyle.Render("g") +
			hintStyle.Render("]oose · [") +
			keyStyle.Render("t") +
			hintStyle.Render("]widdle auto-sync · [") +
			keyStyle.Render("q") +
			hintStyle.Render("]uit"),
	)

	return tea.NewView(borderStyle.Render(sb.String()))
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
