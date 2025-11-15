package theme

import "github.com/gdamore/tcell/v2"

// TokyoNightTheme provides the Tokyo Night color scheme
// Based on: https://github.com/enkia/tokyo-night-vscode-theme
type TokyoNightTheme struct{}

func init() {
	Register(&TokyoNightTheme{})
}

func (t *TokyoNightTheme) Name() string {
	return "tokyo-night"
}

func (t *TokyoNightTheme) PriorityColors() [5]string {
	return [5]string{
		"#f7768e", // P0: red
		"#ff9e64", // P1: orange
		"#7aa2f7", // P2: blue
		"#565f89", // P3: comment
		"#414868", // P4: darker comment
	}
}

func (t *TokyoNightTheme) StatusOpen() string {
	return "#9ece6a" // green
}

func (t *TokyoNightTheme) StatusInProgress() string {
	return "#7aa2f7" // blue
}

func (t *TokyoNightTheme) StatusBlocked() string {
	return "#e0af68" // yellow
}

func (t *TokyoNightTheme) StatusClosed() string {
	return "#565f89" // comment
}

func (t *TokyoNightTheme) DepBlocks() string {
	return "#f7768e" // red
}

func (t *TokyoNightTheme) DepRelated() string {
	return "#7aa2f7" // blue
}

func (t *TokyoNightTheme) DepParentChild() string {
	return "#9ece6a" // green
}

func (t *TokyoNightTheme) DepDiscoveredFrom() string {
	return "#e0af68" // yellow
}

func (t *TokyoNightTheme) Success() string {
	return "#9ece6a" // green
}

func (t *TokyoNightTheme) Error() string {
	return "#f7768e" // red
}

func (t *TokyoNightTheme) Warning() string {
	return "#e0af68" // yellow
}

func (t *TokyoNightTheme) Info() string {
	return "#7dcfff" // cyan
}

func (t *TokyoNightTheme) Muted() string {
	return "#565f89" // comment
}

func (t *TokyoNightTheme) Emphasis() string {
	return "#bb9af7" // purple
}

func (t *TokyoNightTheme) Accent() string {
	return "#7dcfff" // cyan
}

func (t *TokyoNightTheme) SelectionBg() tcell.Color {
	return tcell.NewHexColor(0x283457) // selection
}

func (t *TokyoNightTheme) SelectionFg() tcell.Color {
	return tcell.NewHexColor(0xa9b1d6) // foreground
}

func (t *TokyoNightTheme) BorderNormal() tcell.Color {
	return tcell.NewHexColor(0x565f89) // comment
}

func (t *TokyoNightTheme) BorderFocused() tcell.Color {
	return tcell.NewHexColor(0x7dcfff) // cyan
}

func (t *TokyoNightTheme) AppBackground() tcell.Color {
	return tcell.NewHexColor(0x1a1b26) // background
}

func (t *TokyoNightTheme) AppForeground() tcell.Color {
	return tcell.NewHexColor(0xa9b1d6) // foreground
}

func (t *TokyoNightTheme) InputFieldBackground() tcell.Color {
	return tcell.NewHexColor(0x24283b) // background highlight
}
