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

// CatppuccinLatteTheme provides the Catppuccin Latte (light) color scheme
// Based on: https://github.com/catppuccin/catppuccin
type CatppuccinLatteTheme struct{}

func init() {
	Register(&CatppuccinLatteTheme{})
}

func (t *CatppuccinLatteTheme) Name() string {
	return "catppuccin-latte"
}

func (t *CatppuccinLatteTheme) PriorityColors() [5]string {
	return [5]string{
		"#d20f39", // P0: red
		"#fe640b", // P1: peach (orange)
		"#1e66f5", // P2: blue
		"#9ca0b0", // P3: overlay1
		"#acb0be", // P4: overlay0
	}
}

func (t *CatppuccinLatteTheme) StatusOpen() string {
	return "#40a02b" // green
}

func (t *CatppuccinLatteTheme) StatusInProgress() string {
	return "#1e66f5" // blue
}

func (t *CatppuccinLatteTheme) StatusBlocked() string {
	return "#df8e1d" // yellow
}

func (t *CatppuccinLatteTheme) StatusClosed() string {
	return "#9ca0b0" // overlay1
}

func (t *CatppuccinLatteTheme) DepBlocks() string {
	return "#d20f39" // red
}

func (t *CatppuccinLatteTheme) DepRelated() string {
	return "#1e66f5" // blue
}

func (t *CatppuccinLatteTheme) DepParentChild() string {
	return "#40a02b" // green
}

func (t *CatppuccinLatteTheme) DepDiscoveredFrom() string {
	return "#df8e1d" // yellow
}

func (t *CatppuccinLatteTheme) Success() string {
	return "#40a02b" // green
}

func (t *CatppuccinLatteTheme) Error() string {
	return "#d20f39" // red
}

func (t *CatppuccinLatteTheme) Warning() string {
	return "#df8e1d" // yellow
}

func (t *CatppuccinLatteTheme) Info() string {
	return "#179299" // teal
}

func (t *CatppuccinLatteTheme) Muted() string {
	return "#9ca0b0" // overlay1
}

func (t *CatppuccinLatteTheme) Emphasis() string {
	return "#8839ef" // mauve (purple)
}

func (t *CatppuccinLatteTheme) Accent() string {
	return "#179299" // teal
}

func (t *CatppuccinLatteTheme) SelectionBg() tcell.Color {
	return tcell.NewHexColor(0xe6e9ef) // surface0
}

func (t *CatppuccinLatteTheme) SelectionFg() tcell.Color {
	return tcell.NewHexColor(0x4c4f69) // text
}

func (t *CatppuccinLatteTheme) BorderNormal() tcell.Color {
	return tcell.NewHexColor(0x9ca0b0) // overlay1
}

func (t *CatppuccinLatteTheme) BorderFocused() tcell.Color {
	return tcell.NewHexColor(0x8839ef) // mauve
}
