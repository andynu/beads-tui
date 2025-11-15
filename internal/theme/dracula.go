package theme

import "github.com/gdamore/tcell/v2"

// DraculaTheme provides the Dracula color scheme
// Based on: https://draculatheme.com/
type DraculaTheme struct{}

func init() {
	Register(&DraculaTheme{})
}

func (t *DraculaTheme) Name() string {
	return "dracula"
}

func (t *DraculaTheme) PriorityColors() [5]string {
	return [5]string{
		"#ff5555", // P0: red
		"#ffb86c", // P1: orange
		"#8be9fd", // P2: cyan
		"#6272a4", // P3: comment (purple-gray)
		"#44475a", // P4: current line (darker)
	}
}

func (t *DraculaTheme) StatusOpen() string {
	return "#50fa7b" // green
}

func (t *DraculaTheme) StatusInProgress() string {
	return "#bd93f9" // purple
}

func (t *DraculaTheme) StatusBlocked() string {
	return "#f1fa8c" // yellow
}

func (t *DraculaTheme) StatusClosed() string {
	return "#6272a4" // comment
}

func (t *DraculaTheme) DepBlocks() string {
	return "#ff5555" // red
}

func (t *DraculaTheme) DepRelated() string {
	return "#8be9fd" // cyan
}

func (t *DraculaTheme) DepParentChild() string {
	return "#50fa7b" // green
}

func (t *DraculaTheme) DepDiscoveredFrom() string {
	return "#f1fa8c" // yellow
}

func (t *DraculaTheme) Success() string {
	return "#50fa7b" // green
}

func (t *DraculaTheme) Error() string {
	return "#ff5555" // red
}

func (t *DraculaTheme) Warning() string {
	return "#f1fa8c" // yellow
}

func (t *DraculaTheme) Info() string {
	return "#8be9fd" // cyan
}

func (t *DraculaTheme) Muted() string {
	return "#6272a4" // comment
}

func (t *DraculaTheme) Emphasis() string {
	return "#ff79c6" // pink
}

func (t *DraculaTheme) Accent() string {
	return "#bd93f9" // purple
}

func (t *DraculaTheme) SelectionBg() tcell.Color {
	return tcell.NewHexColor(0x44475a) // current line
}

func (t *DraculaTheme) SelectionFg() tcell.Color {
	return tcell.NewHexColor(0xf8f8f2) // foreground
}

func (t *DraculaTheme) BorderNormal() tcell.Color {
	return tcell.NewHexColor(0x6272a4) // comment
}

func (t *DraculaTheme) BorderFocused() tcell.Color {
	return tcell.NewHexColor(0xff79c6) // pink
}

func (t *DraculaTheme) AppBackground() tcell.Color {
	return tcell.NewHexColor(0x282a36) // background
}

func (t *DraculaTheme) AppForeground() tcell.Color {
	return tcell.NewHexColor(0xf8f8f2) // foreground
}
