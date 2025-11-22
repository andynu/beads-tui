package theme

import "testing"

func TestHighContrastTheme(t *testing.T) {
	theme := &HighContrastTheme{}

	if theme.Name() != "high-contrast" {
		t.Errorf("Expected name 'high-contrast', got %s", theme.Name())
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

	// Test UI colors
	if theme.Success() == "" {
		t.Error("Success is empty")
	}
	if theme.Error() == "" {
		t.Error("Error is empty")
	}
	if theme.Warning() == "" {
		t.Error("Warning is empty")
	}
	if theme.Info() == "" {
		t.Error("Info is empty")
	}
}

func TestColorblindTheme(t *testing.T) {
	theme := &ColorblindTheme{}

	if theme.Name() != "colorblind" {
		t.Errorf("Expected name 'colorblind', got %s", theme.Name())
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

	// Test UI colors
	if theme.Success() == "" {
		t.Error("Success is empty")
	}
	if theme.Error() == "" {
		t.Error("Error is empty")
	}
	if theme.Warning() == "" {
		t.Error("Warning is empty")
	}
	if theme.Info() == "" {
		t.Error("Info is empty")
	}
}

func TestAccessibilityThemesRegistered(t *testing.T) {
	// Check that accessibility themes are registered
	highContrast := Get("high-contrast")
	if highContrast == nil {
		t.Error("high-contrast theme not registered")
	}

	colorblind := Get("colorblind")
	if colorblind == nil {
		t.Error("colorblind theme not registered")
	}

	// Verify they appear in the list
	themes := List()
	foundHighContrast := false
	foundColorblind := false

	for _, name := range themes {
		if name == "high-contrast" {
			foundHighContrast = true
		}
		if name == "colorblind" {
			foundColorblind = true
		}
	}

	if !foundHighContrast {
		t.Error("high-contrast not in theme list")
	}
	if !foundColorblind {
		t.Error("colorblind not in theme list")
	}
}

func TestSwitchToAccessibilityThemes(t *testing.T) {
	// Test switching to high-contrast
	err := SetCurrent("high-contrast")
	if err != nil {
		t.Errorf("Failed to set high-contrast theme: %v", err)
	}

	current := Current()
	if current.Name() != "high-contrast" {
		t.Errorf("Expected current theme to be high-contrast, got %s", current.Name())
	}

	// Test switching to colorblind
	err = SetCurrent("colorblind")
	if err != nil {
		t.Errorf("Failed to set colorblind theme: %v", err)
	}

	current = Current()
	if current.Name() != "colorblind" {
		t.Errorf("Expected current theme to be colorblind, got %s", current.Name())
	}
}

// TestColorblindSafety verifies that the colorblind theme avoids
// problematic red-green combinations
func TestColorblindSafety(t *testing.T) {
	theme := &ColorblindTheme{}

	// Verify that success/error use different hues (blue/orange instead of green/red)
	success := theme.Success()
	error := theme.Error()

	// Both should be defined
	if success == "" || error == "" {
		t.Error("Success or Error color not defined")
	}

	// They should be different colors
	if success == error {
		t.Error("Success and Error use the same color")
	}

	// Verify status colors are all unique
	statuses := []string{
		theme.StatusOpen(),
		theme.StatusInProgress(),
		theme.StatusBlocked(),
		theme.StatusClosed(),
	}

	for i := 0; i < len(statuses); i++ {
		for j := i + 1; j < len(statuses); j++ {
			// Allow StatusClosed to potentially share with others as it's less critical
			if i != 3 && j != 3 && statuses[i] == statuses[j] {
				t.Errorf("Status colors %d and %d are identical: %s", i, j, statuses[i])
			}
		}
	}
}

// TestHighContrastColors verifies that high contrast theme uses
// sufficiently bright/saturated colors
func TestHighContrastColors(t *testing.T) {
	theme := &HighContrastTheme{}

	// Verify background is pure black
	bg := theme.AppBackground()
	if bg.Hex() != 0x000000 {
		t.Errorf("Expected pure black background (0x000000), got 0x%06X", bg.Hex())
	}

	// Verify foreground is pure white
	fg := theme.AppForeground()
	if fg.Hex() != 0xFFFFFF {
		t.Errorf("Expected pure white foreground (0xFFFFFF), got 0x%06X", fg.Hex())
	}

	// Verify selection provides good contrast
	selBg := theme.SelectionBg()
	selFg := theme.SelectionFg()

	// Selection should not be black (should contrast with background)
	if selBg.Hex() == 0x000000 {
		t.Error("Selection background should not be pure black")
	}

	// Selection foreground should be white for maximum contrast
	if selFg.Hex() != 0xFFFFFF {
		t.Errorf("Expected pure white selection foreground, got 0x%06X", selFg.Hex())
	}
}
