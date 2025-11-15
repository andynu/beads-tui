package theme

import "testing"

func TestAllThemesRegistered(t *testing.T) {
	expectedThemes := []string{
		"default",
		"gruvbox-dark",
		"gruvbox-light",
		"nord",
		"solarized-dark",
		"solarized-light",
		"dracula",
		"tokyo-night",
		"catppuccin-mocha",
	}

	themes := List()
	themeMap := make(map[string]bool)
	for _, name := range themes {
		themeMap[name] = true
	}

	for _, expected := range expectedThemes {
		if !themeMap[expected] {
			t.Errorf("Theme %s not registered", expected)
		}

		// Also verify Get returns a non-nil theme
		theme := Get(expected)
		if theme == nil {
			t.Errorf("Get(%s) returned nil", expected)
		}
	}

	t.Logf("Successfully registered %d themes: %v", len(themes), themes)
}

func TestAllThemesCanBeSet(t *testing.T) {
	themes := []string{
		"default",
		"gruvbox-dark",
		"gruvbox-light",
		"nord",
		"solarized-dark",
		"solarized-light",
		"dracula",
		"tokyo-night",
		"catppuccin-mocha",
	}

	for _, themeName := range themes {
		err := SetCurrent(themeName)
		if err != nil {
			t.Errorf("Failed to set theme %s: %v", themeName, err)
		}

		current := Current()
		if current.Name() != themeName {
			t.Errorf("Expected current theme to be %s, got %s", themeName, current.Name())
		}
	}
}

func TestAllThemesHaveCompleteColors(t *testing.T) {
	themes := []string{
		"default",
		"gruvbox-dark",
		"gruvbox-light",
		"nord",
		"solarized-dark",
		"solarized-light",
		"dracula",
		"tokyo-night",
		"catppuccin-mocha",
	}

	for _, themeName := range themes {
		theme := Get(themeName)
		if theme == nil {
			t.Fatalf("Theme %s not found", themeName)
		}

		// Test priority colors
		priorities := theme.PriorityColors()
		if len(priorities) != 5 {
			t.Errorf("%s: Expected 5 priority colors, got %d", themeName, len(priorities))
		}
		for i, color := range priorities {
			if color == "" {
				t.Errorf("%s: Priority %d color is empty", themeName, i)
			}
		}

		// Test status colors
		if theme.StatusOpen() == "" {
			t.Errorf("%s: StatusOpen is empty", themeName)
		}
		if theme.StatusInProgress() == "" {
			t.Errorf("%s: StatusInProgress is empty", themeName)
		}
		if theme.StatusBlocked() == "" {
			t.Errorf("%s: StatusBlocked is empty", themeName)
		}
		if theme.StatusClosed() == "" {
			t.Errorf("%s: StatusClosed is empty", themeName)
		}

		// Test dependency colors
		if theme.DepBlocks() == "" {
			t.Errorf("%s: DepBlocks is empty", themeName)
		}
		if theme.DepRelated() == "" {
			t.Errorf("%s: DepRelated is empty", themeName)
		}
		if theme.DepParentChild() == "" {
			t.Errorf("%s: DepParentChild is empty", themeName)
		}
		if theme.DepDiscoveredFrom() == "" {
			t.Errorf("%s: DepDiscoveredFrom is empty", themeName)
		}

		// Test semantic colors
		if theme.Success() == "" {
			t.Errorf("%s: Success is empty", themeName)
		}
		if theme.Error() == "" {
			t.Errorf("%s: Error is empty", themeName)
		}
		if theme.Warning() == "" {
			t.Errorf("%s: Warning is empty", themeName)
		}
		if theme.Info() == "" {
			t.Errorf("%s: Info is empty", themeName)
		}
		if theme.Muted() == "" {
			t.Errorf("%s: Muted is empty", themeName)
		}
		if theme.Emphasis() == "" {
			t.Errorf("%s: Emphasis is empty", themeName)
		}
		if theme.Accent() == "" {
			t.Errorf("%s: Accent is empty", themeName)
		}
	}
}
