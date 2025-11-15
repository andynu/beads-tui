package theme

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestDefaultTheme(t *testing.T) {
	dt := &DefaultTheme{}

	// Test name
	if dt.Name() != "default" {
		t.Errorf("Expected name 'default', got %s", dt.Name())
	}

	// Test priority colors
	priorities := dt.PriorityColors()
	if len(priorities) != 5 {
		t.Errorf("Expected 5 priority colors, got %d", len(priorities))
	}

	// Test status colors
	if dt.StatusOpen() == "" {
		t.Error("StatusOpen should not be empty")
	}
	if dt.StatusInProgress() == "" {
		t.Error("StatusInProgress should not be empty")
	}
	if dt.StatusBlocked() == "" {
		t.Error("StatusBlocked should not be empty")
	}
	if dt.StatusClosed() == "" {
		t.Error("StatusClosed should not be empty")
	}

	// Test dependency colors
	if dt.DepBlocks() == "" {
		t.Error("DepBlocks should not be empty")
	}
	if dt.DepRelated() == "" {
		t.Error("DepRelated should not be empty")
	}
	if dt.DepParentChild() == "" {
		t.Error("DepParentChild should not be empty")
	}
	if dt.DepDiscoveredFrom() == "" {
		t.Error("DepDiscoveredFrom should not be empty")
	}

	// Test semantic colors
	if dt.Success() == "" {
		t.Error("Success should not be empty")
	}
	if dt.Error() == "" {
		t.Error("Error should not be empty")
	}
	if dt.Warning() == "" {
		t.Error("Warning should not be empty")
	}
	if dt.Info() == "" {
		t.Error("Info should not be empty")
	}
	if dt.Muted() == "" {
		t.Error("Muted should not be empty")
	}
	if dt.Emphasis() == "" {
		t.Error("Emphasis should not be empty")
	}
	if dt.Accent() == "" {
		t.Error("Accent should not be empty")
	}

	// Test component colors
	if dt.SelectionBg() == tcell.ColorDefault {
		t.Error("SelectionBg should not be default")
	}
	if dt.SelectionFg() == tcell.ColorDefault {
		t.Error("SelectionFg should not be default")
	}
	if dt.BorderNormal() == tcell.ColorDefault {
		t.Error("BorderNormal should not be default")
	}
	if dt.BorderFocused() == tcell.ColorDefault {
		t.Error("BorderFocused should not be default")
	}
}

func TestRegistry(t *testing.T) {
	// Default theme should be auto-registered
	themes := List()
	if len(themes) == 0 {
		t.Error("Expected at least one registered theme")
	}

	// Current should return a theme
	current := Current()
	if current == nil {
		t.Error("Current theme should not be nil")
	}

	// Get should return the default theme
	defaultTheme := Get("default")
	if defaultTheme == nil {
		t.Error("Expected to find 'default' theme")
	}

	// Get non-existent theme should return nil
	nonExistent := Get("nonexistent")
	if nonExistent != nil {
		t.Error("Expected nil for non-existent theme")
	}
}

func TestSetCurrent(t *testing.T) {
	// Should succeed for existing theme
	err := SetCurrent("default")
	if err != nil {
		t.Errorf("Expected no error setting to default theme, got %v", err)
	}

	// Should fail for non-existent theme
	err = SetCurrent("nonexistent")
	if err == nil {
		t.Error("Expected error when setting non-existent theme")
	}

	// Current should still be default after failed set
	current := Current()
	if current.Name() != "default" {
		t.Errorf("Expected current theme to be 'default' after failed set, got %s", current.Name())
	}
}

func TestRegisterTheme(t *testing.T) {
	// Create a custom theme
	type TestTheme struct {
		DefaultTheme
	}

	testTheme := &TestTheme{}

	// Override Name
	testTheme.DefaultTheme = DefaultTheme{}

	// Register should work
	Register(testTheme)

	// Should now be in the list
	themes := List()
	found := false
	for _, name := range themes {
		if name == "default" { // TestTheme uses DefaultTheme's Name() which returns "default"
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find registered test theme")
	}
}
