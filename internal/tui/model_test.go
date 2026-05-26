package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/ianeff/gooser/internal/gooser"
)

// threeApps is a stable fixture used across model tests.
func threeApps() []gooser.Application {
	return []gooser.Application{
		{Name: "alpha", Sync: "Synced", Health: "Healthy"},
		{Name: "beta", Sync: "OutOfSync", Health: "Degraded"},
		{Name: "gamma", Sync: "Synced", Health: "Progressing"},
	}
}

// keyPress constructs a KeyPressMsg whose String() returns the given key string.
// Supports rune keys ("g", "t", …), "up", "down", "enter", and "ctrl+c".
func keyPress(key string) tea.KeyPressMsg {
	switch key {
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "ctrl+c":
		return tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	default:
		r := []rune(key)
		return tea.KeyPressMsg{Code: r[0], Text: key}
	}
}

// updateModel is a thin wrapper so tests don't have to cast.
func updateModel(m model, key string) model {
	next, _ := m.Update(keyPress(key))
	return next.(model)
}

// TestUpdate_Navigation verifies cursor movement and its boundary clamping.
func TestUpdate_Navigation(t *testing.T) {
	apps := threeApps() // indices 0, 1, 2

	t.Run("initial cursor is zero", func(t *testing.T) {
		m := initialModel(apps)
		if m.cursor != 0 {
			t.Errorf("cursor = %d, want 0", m.cursor)
		}
	})

	t.Run("down moves cursor forward", func(t *testing.T) {
		m := updateModel(initialModel(apps), "down")
		if m.cursor != 1 {
			t.Errorf("cursor = %d, want 1", m.cursor)
		}
	})

	t.Run("j moves cursor forward", func(t *testing.T) {
		m := updateModel(initialModel(apps), "j")
		if m.cursor != 1 {
			t.Errorf("cursor = %d, want 1", m.cursor)
		}
	})

	t.Run("down clamps at last item", func(t *testing.T) {
		m := initialModel(apps)
		// Move to end then try to go further.
		for range apps {
			m = updateModel(m, "down")
		}
		if m.cursor != len(apps)-1 {
			t.Errorf("cursor = %d, want %d (clamped)", m.cursor, len(apps)-1)
		}
	})

	t.Run("up from zero stays at zero", func(t *testing.T) {
		m := updateModel(initialModel(apps), "up")
		if m.cursor != 0 {
			t.Errorf("cursor = %d, want 0 (clamped)", m.cursor)
		}
	})

	t.Run("k moves cursor backward", func(t *testing.T) {
		m := initialModel(apps)
		m = updateModel(m, "down") // cursor=1
		m = updateModel(m, "k")    // cursor=0
		if m.cursor != 0 {
			t.Errorf("cursor = %d, want 0", m.cursor)
		}
	})

	t.Run("up moves cursor backward", func(t *testing.T) {
		m := initialModel(apps)
		m = updateModel(m, "down") // cursor=1
		m = updateModel(m, "up")   // cursor=0
		if m.cursor != 0 {
			t.Errorf("cursor = %d, want 0", m.cursor)
		}
	})
}

// TestUpdate_Actions verifies that action keys set the right Result and signal quit.
func TestUpdate_Actions(t *testing.T) {
	tests := []struct {
		key        string
		wantAction Action
	}{
		{"g", ActionGoose},
		{"enter", ActionGoose},
		{"t", ActionTwiddleOn},
		{"o", ActionTwiddleOff},
	}

	for _, tt := range tests {
		t.Run("key "+tt.key+" at cursor 0 selects first app", func(t *testing.T) {
			m := updateModel(initialModel(threeApps()), tt.key)

			if !m.quitting {
				t.Error("quitting = false, want true")
			}
			if m.result.Action != tt.wantAction {
				t.Errorf("result.Action = %v, want %v", m.result.Action, tt.wantAction)
			}
			if m.result.Application != "alpha" {
				t.Errorf("result.Application = %q, want %q", m.result.Application, "alpha")
			}
		})

		t.Run("key "+tt.key+" at cursor 1 selects second app", func(t *testing.T) {
			m := initialModel(threeApps())
			m = updateModel(m, "down") // cursor=1 → "beta"
			m = updateModel(m, tt.key)

			if m.result.Application != "beta" {
				t.Errorf("result.Application = %q, want %q", m.result.Application, "beta")
			}
			if m.result.Action != tt.wantAction {
				t.Errorf("result.Action = %v, want %v", m.result.Action, tt.wantAction)
			}
		})
	}
}

// TestUpdate_Quit verifies that q and ctrl+c set quitting without touching result.
func TestUpdate_Quit(t *testing.T) {
	for _, key := range []string{"q", "ctrl+c"} {
		t.Run("key "+key+" sets quitting", func(t *testing.T) {
			m := updateModel(initialModel(threeApps()), key)
			if !m.quitting {
				t.Error("quitting = false, want true")
			}
			// Quit should not record an application selection.
			if m.result.Application != "" {
				t.Errorf("result.Application = %q, want empty", m.result.Application)
			}
		})
	}
}
