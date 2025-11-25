package theme

import (
	"testing"
)

func TestLoadTOMLTheme(t *testing.T) {
	// Test loading gruvbox-dark theme
	theme, err := LoadTOMLTheme("gruvbox-dark")
	if err != nil {
		t.Fatalf("Failed to load gruvbox-dark theme: %v", err)
	}

	if theme.Name() != "gruvbox-dark" {
		t.Errorf("Expected name 'gruvbox-dark', got %s", theme.Name())
	}

	// Test that colors are defined
	if len(theme.PriorityColors()) != 5 {
		t.Errorf("Expected 5 priority colors, got %d", len(theme.PriorityColors()))
	}

	// Test status colors
	if theme.StatusOpen() == "" {
		t.Error("StatusOpen is empty")
	}
	if theme.StatusInProgress() == "" {
		t.Error("StatusInProgress is empty")
	}
}

func TestLoadAllEmbeddedThemes(t *testing.T) {
	// This test verifies that all TOML themes are loaded automatically
	// The init() function should have already loaded them

	// Check for expected themes
	expectedThemes := []string{
		"default",
		"gruvbox-dark",
		"high-contrast",
		"colorblind",
		"nord",
		"dracula",
	}

	for _, name := range expectedThemes {
		theme := Get(name)
		if theme == nil {
			t.Errorf("Theme %s not loaded", name)
		}
	}
}

func TestTOMLThemeRegistration(t *testing.T) {
	// Verify TOML themes appear in the registry
	themes := List()

	foundGruvbox := false
	foundHighContrast := false

	for _, name := range themes {
		if name == "gruvbox-dark" {
			foundGruvbox = true
		}
		if name == "high-contrast" {
			foundHighContrast = true
		}
	}

	if !foundGruvbox {
		t.Error("gruvbox-dark TOML theme not in registry")
	}
	if !foundHighContrast {
		t.Error("high-contrast TOML theme not in registry")
	}
}

func TestSwitchToTOMLTheme(t *testing.T) {
	// Test switching to a TOML theme
	err := SetCurrent("nord")
	if err != nil {
		t.Errorf("Failed to set nord theme: %v", err)
	}

	current := Current()
	if current.Name() != "nord" {
		t.Errorf("Expected current theme to be nord, got %s", current.Name())
	}
}

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		input    string
		expected int32
	}{
		{"#FF0000", 0xFF0000},  // red
		{"#00FF00", 0x00FF00},  // green
		{"#0000FF", 0x0000FF},  // blue
		{"#282828", 0x282828},  // gruvbox bg
		{"#FFFFFF", 0xFFFFFF},  // white
		{"#000000", 0x000000},  // black
		{"FF0000", 0xFF0000},   // without # prefix
	}

	for _, tt := range tests {
		color := parseHexColor(tt.input)
		if color.Hex() != tt.expected {
			t.Errorf("parseHexColor(%s) = 0x%06X, expected 0x%06X", tt.input, color.Hex(), tt.expected)
		}
	}
}
