// Package tui provides a superslick charmbracelet-based TUI for the gooser's pleasure.
package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Catppuccin Frappé palette — https://catppuccin.com/palette
// Only the swatches we actually use.
var (
	// Foregrounds.
	colText     = lipgloss.Color("#c6d0f5")
	colSubtext0 = lipgloss.Color("#a5adce")

	// Accents.
	colMauve    = lipgloss.Color("#ca9ee6")
	colPink     = lipgloss.Color("#f4b8e4")
	colLavender = lipgloss.Color("#babbf1")

	// Status colours.
	colGreen  = lipgloss.Color("#a6d189")
	colRed    = lipgloss.Color("#e78284")
	colYellow = lipgloss.Color("#e5c890")

	// Pre-built styles for the Frappé-themed TUI.

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colMauve).
			PaddingBottom(1)

	borderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colLavender).
			Padding(1, 2)

	cursorStyle = lipgloss.NewStyle().Foreground(colPink)

	nameStyle = lipgloss.NewStyle().
			Width(nameColWidth).
			Foreground(colText)

	selectedNameStyle = lipgloss.NewStyle().
				Width(nameColWidth).
				Bold(true).
				Foreground(colMauve)

	labelStyle = lipgloss.NewStyle().Foreground(colSubtext0)
	keyStyle   = lipgloss.NewStyle().Foreground(colLavender).Bold(true)
	hintStyle  = lipgloss.NewStyle().Foreground(colSubtext0)
)

// Column widths for the application list.
const (
	nameColWidth   = 36
	statusColWidth = 12
)

// statusColor returns the Frappé accent colour that best represents an ArgoCD status string.
func statusColor(s string) color.Color {
	switch s {
	case "Synced", "Healthy":
		return colGreen
	case "OutOfSync", "Degraded", "Missing":
		return colRed
	case "Progressing", "Suspended":
		return colYellow
	default:
		return colSubtext0
	}
}

// syncCell renders a sync-status string coloured and padded to statusColWidth.
func syncCell(s string) string {
	return lipgloss.NewStyle().Foreground(statusColor(s)).Width(statusColWidth).Render(s)
}

// healthCell renders a health-status string in its status colour (no fixed width; last column).
func healthCell(s string) string {
	return lipgloss.NewStyle().Foreground(statusColor(s)).Render(s)
}
