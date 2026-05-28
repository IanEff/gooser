// Test render.go for when we want to really lipgloss it.
package tui

import "testing"

func TestRenderGoosed(t *testing.T) {
	got := RenderGoosed("my-app")
	want := "Goosed my-app. 🪿"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenderTwiddle(t *testing.T) {
	tests := []struct {
		app     string
		enabled bool
		want    string
	}{
		{"my-app", true, "Enabled auto-sync for my-app."},
		{"my-app", false, "Disabled auto-sync for my-app."},
	}
	for _, tt := range tests {
		got := RenderTwiddle(tt.app, tt.enabled)
		if got != tt.want {
			t.Errorf("RenderTwiddle(%q, %v) = %q, want %q", tt.app, tt.enabled, got, tt.want)
		}
	}
}
