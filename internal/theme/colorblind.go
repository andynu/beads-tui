package theme

import "github.com/gdamore/tcell/v2"

// ColorblindTheme provides a color scheme optimized for colorblindness
// Designed specifically for deuteranopia (red-green colorblindness)
// Uses blue/orange/purple palette that remains distinguishable
type ColorblindTheme struct{}

func init() {
	Register(&ColorblindTheme{})
}

func (t *ColorblindTheme) Name() string {
	return "colorblind"
}

func (t *ColorblindTheme) PriorityColors() [5]string {
	return [5]string{
		"#D55E00", // P0: vermillion orange (critical) - safe for deuteranopia
		"#E69F00", // P1: orange (high)
		"#0072B2", // P2: blue (normal)
		"#999999", // P3: gray (low)
		"#666666", // P4: dark gray (lowest)
	}
}

func (t *ColorblindTheme) StatusOpen() string {
	return "#56B4E9" // sky blue (replaces green)
}

func (t *ColorblindTheme) StatusInProgress() string {
	return "#0072B2" // blue
}

func (t *ColorblindTheme) StatusBlocked() string {
	return "#E69F00" // orange (replaces yellow)
}

func (t *ColorblindTheme) StatusClosed() string {
	return "#999999" // gray
}

func (t *ColorblindTheme) DepBlocks() string {
	return "#D55E00" // vermillion orange
}

func (t *ColorblindTheme) DepRelated() string {
	return "#56B4E9" // sky blue
}

func (t *ColorblindTheme) DepParentChild() string {
	return "#009E73" // bluish green (safe for deuteranopia)
}

func (t *ColorblindTheme) DepDiscoveredFrom() string {
	return "#CC79A7" // reddish purple
}

func (t *ColorblindTheme) Success() string {
	return "#56B4E9" // sky blue (instead of green)
}

func (t *ColorblindTheme) Error() string {
	return "#D55E00" // vermillion orange (instead of red)
}

func (t *ColorblindTheme) Warning() string {
	return "#E69F00" // orange
}

func (t *ColorblindTheme) Info() string {
	return "#0072B2" // blue
}

func (t *ColorblindTheme) Muted() string {
	return "#999999" // gray
}

func (t *ColorblindTheme) Emphasis() string {
	return "#F0E442" // yellow (high luminance, safe)
}

func (t *ColorblindTheme) Accent() string {
	return "#CC79A7" // reddish purple
}

func (t *ColorblindTheme) SelectionBg() tcell.Color {
	return tcell.NewHexColor(0x2B3E50) // dark blue-gray
}

func (t *ColorblindTheme) SelectionFg() tcell.Color {
	return tcell.NewHexColor(0xECF0F1) // light gray (almost white)
}

func (t *ColorblindTheme) BorderNormal() tcell.Color {
	return tcell.NewHexColor(0x7F8C8D) // medium gray
}

func (t *ColorblindTheme) BorderFocused() tcell.Color {
	return tcell.NewHexColor(0xF0E442) // yellow
}

func (t *ColorblindTheme) AppBackground() tcell.Color {
	return tcell.NewHexColor(0x1C2833) // dark blue-gray
}

func (t *ColorblindTheme) AppForeground() tcell.Color {
	return tcell.NewHexColor(0xECF0F1) // light gray
}

func (t *ColorblindTheme) InputFieldBackground() tcell.Color {
	return tcell.NewHexColor(0x2B3E50) // slightly lighter dark blue-gray
}
