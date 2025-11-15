package theme

import "github.com/gdamore/tcell/v2"

// DefaultTheme provides the original beads-tui color scheme
type DefaultTheme struct{}

func init() {
	Register(&DefaultTheme{})
}

func (t *DefaultTheme) Name() string {
	return "default"
}

func (t *DefaultTheme) PriorityColors() [5]string {
	return [5]string{
		"red",           // P0: Critical - bright red
		"orangered",     // P1: High - orange-red for urgency
		"lightskyblue",  // P2: Normal - calm blue
		"darkgray",      // P3: Low - subdued gray
		"gray",          // P4: Lowest - very subdued
	}
}

func (t *DefaultTheme) StatusOpen() string {
	return "limegreen" // Bright green for ready work
}

func (t *DefaultTheme) StatusInProgress() string {
	return "deepskyblue" // Vibrant blue for active work
}

func (t *DefaultTheme) StatusBlocked() string {
	return "gold" // Gold/yellow for warning
}

func (t *DefaultTheme) StatusClosed() string {
	return "darkgray" // Muted gray
}

func (t *DefaultTheme) DepBlocks() string {
	return "red"
}

func (t *DefaultTheme) DepRelated() string {
	return "blue"
}

func (t *DefaultTheme) DepParentChild() string {
	return "green"
}

func (t *DefaultTheme) DepDiscoveredFrom() string {
	return "yellow"
}

func (t *DefaultTheme) Success() string {
	return "limegreen"
}

func (t *DefaultTheme) Error() string {
	return "red"
}

func (t *DefaultTheme) Warning() string {
	return "gold"
}

func (t *DefaultTheme) Info() string {
	return "cyan"
}

func (t *DefaultTheme) Muted() string {
	return "gray"
}

func (t *DefaultTheme) Emphasis() string {
	return "yellow"
}

func (t *DefaultTheme) Accent() string {
	return "cyan"
}

func (t *DefaultTheme) SelectionBg() tcell.Color {
	return tcell.ColorBlue
}

func (t *DefaultTheme) SelectionFg() tcell.Color {
	return tcell.ColorWhite
}

func (t *DefaultTheme) BorderNormal() tcell.Color {
	return tcell.ColorWhite
}

func (t *DefaultTheme) BorderFocused() tcell.Color {
	return tcell.ColorYellow
}

func (t *DefaultTheme) AppBackground() tcell.Color {
	return tcell.ColorBlack
}

func (t *DefaultTheme) AppForeground() tcell.Color {
	return tcell.ColorWhite
}

func (t *DefaultTheme) InputFieldBackground() tcell.Color {
	return tcell.NewHexColor(0x1a1a1a) // Slightly lighter than black background
}
