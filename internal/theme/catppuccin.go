package theme

import "github.com/gdamore/tcell/v2"

// CatppuccinMochaTheme provides the Catppuccin Mocha color scheme
// Based on: https://github.com/catppuccin/catppuccin
type CatppuccinMochaTheme struct{}

func init() {
	Register(&CatppuccinMochaTheme{})
}

func (t *CatppuccinMochaTheme) Name() string {
	return "catppuccin-mocha"
}

func (t *CatppuccinMochaTheme) PriorityColors() [5]string {
	return [5]string{
		"#f38ba8", // P0: red
		"#fab387", // P1: peach (orange)
		"#89b4fa", // P2: blue
		"#6c7086", // P3: overlay1
		"#585b70", // P4: overlay0
	}
}

func (t *CatppuccinMochaTheme) StatusOpen() string {
	return "#a6e3a1" // green
}

func (t *CatppuccinMochaTheme) StatusInProgress() string {
	return "#89b4fa" // blue
}

func (t *CatppuccinMochaTheme) StatusBlocked() string {
	return "#f9e2af" // yellow
}

func (t *CatppuccinMochaTheme) StatusClosed() string {
	return "#6c7086" // overlay1
}

func (t *CatppuccinMochaTheme) DepBlocks() string {
	return "#f38ba8" // red
}

func (t *CatppuccinMochaTheme) DepRelated() string {
	return "#89b4fa" // blue
}

func (t *CatppuccinMochaTheme) DepParentChild() string {
	return "#a6e3a1" // green
}

func (t *CatppuccinMochaTheme) DepDiscoveredFrom() string {
	return "#f9e2af" // yellow
}

func (t *CatppuccinMochaTheme) Success() string {
	return "#a6e3a1" // green
}

func (t *CatppuccinMochaTheme) Error() string {
	return "#f38ba8" // red
}

func (t *CatppuccinMochaTheme) Warning() string {
	return "#f9e2af" // yellow
}

func (t *CatppuccinMochaTheme) Info() string {
	return "#94e2d5" // teal
}

func (t *CatppuccinMochaTheme) Muted() string {
	return "#6c7086" // overlay1
}

func (t *CatppuccinMochaTheme) Emphasis() string {
	return "#cba6f7" // mauve (purple)
}

func (t *CatppuccinMochaTheme) Accent() string {
	return "#94e2d5" // teal
}

func (t *CatppuccinMochaTheme) SelectionBg() tcell.Color {
	return tcell.NewHexColor(0x313244) // surface0
}

func (t *CatppuccinMochaTheme) SelectionFg() tcell.Color {
	return tcell.NewHexColor(0xcdd6f4) // text
}

func (t *CatppuccinMochaTheme) BorderNormal() tcell.Color {
	return tcell.NewHexColor(0x6c7086) // overlay1
}

func (t *CatppuccinMochaTheme) BorderFocused() tcell.Color {
	return tcell.NewHexColor(0xcba6f7) // mauve
}
