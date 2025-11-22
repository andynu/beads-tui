package theme

import "github.com/gdamore/tcell/v2"

// HighContrastTheme provides maximum contrast for accessibility
// Designed for users with low vision or visual impairments
// Uses pure black/white with bold saturated colors
type HighContrastTheme struct{}

func init() {
	Register(&HighContrastTheme{})
}

func (t *HighContrastTheme) Name() string {
	return "high-contrast"
}

func (t *HighContrastTheme) PriorityColors() [5]string {
	return [5]string{
		"#FF0000", // P0: pure red (critical)
		"#FF8800", // P1: bright orange (high)
		"#00DDFF", // P2: cyan (normal)
		"#AAAAAA", // P3: light gray (low)
		"#666666", // P4: dark gray (lowest)
	}
}

func (t *HighContrastTheme) StatusOpen() string {
	return "#00FF00" // bright green
}

func (t *HighContrastTheme) StatusInProgress() string {
	return "#00DDFF" // bright cyan
}

func (t *HighContrastTheme) StatusBlocked() string {
	return "#FFFF00" // bright yellow
}

func (t *HighContrastTheme) StatusClosed() string {
	return "#888888" // medium gray
}

func (t *HighContrastTheme) DepBlocks() string {
	return "#FF0000" // bright red
}

func (t *HighContrastTheme) DepRelated() string {
	return "#00DDFF" // bright cyan
}

func (t *HighContrastTheme) DepParentChild() string {
	return "#00FF00" // bright green
}

func (t *HighContrastTheme) DepDiscoveredFrom() string {
	return "#FFFF00" // bright yellow
}

func (t *HighContrastTheme) Success() string {
	return "#00FF00" // bright green
}

func (t *HighContrastTheme) Error() string {
	return "#FF0000" // bright red
}

func (t *HighContrastTheme) Warning() string {
	return "#FFFF00" // bright yellow
}

func (t *HighContrastTheme) Info() string {
	return "#00DDFF" // bright cyan
}

func (t *HighContrastTheme) Muted() string {
	return "#AAAAAA" // light gray
}

func (t *HighContrastTheme) Emphasis() string {
	return "#FFFF00" // bright yellow
}

func (t *HighContrastTheme) Accent() string {
	return "#FF00FF" // bright magenta
}

func (t *HighContrastTheme) SelectionBg() tcell.Color {
	return tcell.NewHexColor(0x333333) // dark gray
}

func (t *HighContrastTheme) SelectionFg() tcell.Color {
	return tcell.NewHexColor(0xFFFFFF) // pure white
}

func (t *HighContrastTheme) BorderNormal() tcell.Color {
	return tcell.NewHexColor(0x888888) // medium gray
}

func (t *HighContrastTheme) BorderFocused() tcell.Color {
	return tcell.NewHexColor(0xFFFF00) // bright yellow
}

func (t *HighContrastTheme) AppBackground() tcell.Color {
	return tcell.NewHexColor(0x000000) // pure black
}

func (t *HighContrastTheme) AppForeground() tcell.Color {
	return tcell.NewHexColor(0xFFFFFF) // pure white
}

func (t *HighContrastTheme) InputFieldBackground() tcell.Color {
	return tcell.NewHexColor(0x222222) // very dark gray
}
