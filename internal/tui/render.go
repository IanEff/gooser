// Package tui provides terminal user interface rendering functions for the Gooser CLI.
package tui

import "fmt"

// RenderGoosed returns a string indicating that the given application has been goosed.
func RenderGoosed(application string) string {
	return fmt.Sprintf("Goosed %s. 🪿", application)
}

// RenderTwiddle returns a string indicating the auto-sync status of the given application.
func RenderTwiddle(application string, enabled bool) string {
	verb := "Disabled"
	if enabled {
		verb = "Enabled"
	}
	return fmt.Sprintf("%s auto-sync for %s.", verb, application)
}
