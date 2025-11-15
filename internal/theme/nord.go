package theme

import "github.com/gdamore/tcell/v2"

// NordTheme provides the Nord color scheme
// Based on: https://www.nordtheme.com/
type NordTheme struct{}

func init() {
	Register(&NordTheme{})
}

func (t *NordTheme) Name() string {
	return "nord"
}

func (t *NordTheme) PriorityColors() [5]string {
	return [5]string{
		"#bf616a", // P0: nord11 (red)
		"#d08770", // P1: nord12 (orange)
		"#81a1c1", // P2: nord9 (blue)
		"#4c566a", // P3: nord3 (dark gray)
		"#3b4252", // P4: nord1 (darker gray)
	}
}

func (t *NordTheme) StatusOpen() string {
	return "#a3be8c" // nord14 (green)
}

func (t *NordTheme) StatusInProgress() string {
	return "#88c0d0" // nord8 (cyan)
}

func (t *NordTheme) StatusBlocked() string {
	return "#ebcb8b" // nord13 (yellow)
}

func (t *NordTheme) StatusClosed() string {
	return "#4c566a" // nord3 (gray)
}

func (t *NordTheme) DepBlocks() string {
	return "#bf616a" // nord11 (red)
}

func (t *NordTheme) DepRelated() string {
	return "#81a1c1" // nord9 (blue)
}

func (t *NordTheme) DepParentChild() string {
	return "#a3be8c" // nord14 (green)
}

func (t *NordTheme) DepDiscoveredFrom() string {
	return "#ebcb8b" // nord13 (yellow)
}

func (t *NordTheme) Success() string {
	return "#a3be8c" // nord14 (green)
}

func (t *NordTheme) Error() string {
	return "#bf616a" // nord11 (red)
}

func (t *NordTheme) Warning() string {
	return "#ebcb8b" // nord13 (yellow)
}

func (t *NordTheme) Info() string {
	return "#81a1c1" // nord9 (blue)
}

func (t *NordTheme) Muted() string {
	return "#4c566a" // nord3 (gray)
}

func (t *NordTheme) Emphasis() string {
	return "#88c0d0" // nord8 (cyan)
}

func (t *NordTheme) Accent() string {
	return "#5e81ac" // nord10 (frost blue)
}

func (t *NordTheme) SelectionBg() tcell.Color {
	return tcell.NewHexColor(0x3b4252) // nord1
}

func (t *NordTheme) SelectionFg() tcell.Color {
	return tcell.NewHexColor(0xeceff4) // nord6
}

func (t *NordTheme) BorderNormal() tcell.Color {
	return tcell.NewHexColor(0x4c566a) // nord3
}

func (t *NordTheme) BorderFocused() tcell.Color {
	return tcell.NewHexColor(0x88c0d0) // nord8
}
