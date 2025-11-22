package theme

import "testing"

func TestGruvboxDarkTheme(t *testing.T) {
	theme := Get("gruvbox-dark")
	if theme == nil {
		t.Fatal("gruvbox-dark theme not found")
	}

	if theme.Name() != "gruvbox-dark" {
		t.Errorf("Expected name 'gruvbox-dark', got %s", theme.Name())
	}

	// Test that all colors are defined (not empty)
	priorities := theme.PriorityColors()
	if len(priorities) != 5 {
		t.Errorf("Expected 5 priority colors, got %d", len(priorities))
	}

	for i, color := range priorities {
		if color == "" {
			t.Errorf("Priority %d color is empty", i)
		}
	}

	// Test status colors
	if theme.StatusOpen() == "" {
		t.Error("StatusOpen is empty")
	}
	if theme.StatusInProgress() == "" {
		t.Error("StatusInProgress is empty")
	}
	if theme.StatusBlocked() == "" {
		t.Error("StatusBlocked is empty")
	}
	if theme.StatusClosed() == "" {
		t.Error("StatusClosed is empty")
	}
}

func TestGruvboxLightTheme(t *testing.T) {
	theme := Get("gruvbox-light")
	if theme == nil {
		t.Fatal("gruvbox-light theme not found")
	}

	if theme.Name() != "gruvbox-light" {
		t.Errorf("Expected name 'gruvbox-light', got %s", theme.Name())
	}

	// Test that all colors are defined (not empty)
	priorities := theme.PriorityColors()
	if len(priorities) != 5 {
		t.Errorf("Expected 5 priority colors, got %d", len(priorities))
	}

	for i, color := range priorities {
		if color == "" {
			t.Errorf("Priority %d color is empty", i)
		}
	}
}

func TestGruvboxThemesRegistered(t *testing.T) {
	// Check that gruvbox themes are registered
	gruvboxDark := Get("gruvbox-dark")
	if gruvboxDark == nil {
		t.Error("gruvbox-dark theme not registered")
	}

	gruvboxLight := Get("gruvbox-light")
	if gruvboxLight == nil {
		t.Error("gruvbox-light theme not registered")
	}

	// Verify they appear in the list
	themes := List()
	foundDark := false
	foundLight := false

	for _, name := range themes {
		if name == "gruvbox-dark" {
			foundDark = true
		}
		if name == "gruvbox-light" {
			foundLight = true
		}
	}

	if !foundDark {
		t.Error("gruvbox-dark not in theme list")
	}
	if !foundLight {
		t.Error("gruvbox-light not in theme list")
	}
}

func TestSwitchToGruvboxTheme(t *testing.T) {
	// Test switching to gruvbox-dark
	err := SetCurrent("gruvbox-dark")
	if err != nil {
		t.Errorf("Failed to set gruvbox-dark theme: %v", err)
	}

	current := Current()
	if current.Name() != "gruvbox-dark" {
		t.Errorf("Expected current theme to be gruvbox-dark, got %s", current.Name())
	}

	// Test switching to gruvbox-light
	err = SetCurrent("gruvbox-light")
	if err != nil {
		t.Errorf("Failed to set gruvbox-light theme: %v", err)
	}

	current = Current()
	if current.Name() != "gruvbox-light" {
		t.Errorf("Expected current theme to be gruvbox-light, got %s", current.Name())
	}
}
