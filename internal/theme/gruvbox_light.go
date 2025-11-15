package theme

import "github.com/gdamore/tcell/v2"

// GruvboxLightTheme provides the Gruvbox Light color scheme
// Based on: https://github.com/morhetz/gruvbox
type GruvboxLightTheme struct{}

func init() {
	Register(&GruvboxLightTheme{})
}

func (t *GruvboxLightTheme) Name() string {
	return "gruvbox-light"
}

func (t *GruvboxLightTheme) PriorityColors() [5]string {
	return [5]string{
		"#cc241d", // P0: red
		"#d65d0e", // P1: orange
		"#458588", // P2: blue
		"#928374", // P3: gray
		"#a89984", // P4: light gray
	}
}

func (t *GruvboxLightTheme) StatusOpen() string {
	return "#98971a" // green
}

func (t *GruvboxLightTheme) StatusInProgress() string {
	return "#458588" // blue
}

func (t *GruvboxLightTheme) StatusBlocked() string {
	return "#d79921" // yellow
}

func (t *GruvboxLightTheme) StatusClosed() string {
	return "#928374" // gray
}

func (t *GruvboxLightTheme) DepBlocks() string {
	return "#cc241d" // red
}

func (t *GruvboxLightTheme) DepRelated() string {
	return "#458588" // blue
}

func (t *GruvboxLightTheme) DepParentChild() string {
	return "#98971a" // green
}

func (t *GruvboxLightTheme) DepDiscoveredFrom() string {
	return "#d79921" // yellow
}

func (t *GruvboxLightTheme) Success() string {
	return "#98971a" // green
}

func (t *GruvboxLightTheme) Error() string {
	return "#cc241d" // red
}

func (t *GruvboxLightTheme) Warning() string {
	return "#d79921" // yellow
}

func (t *GruvboxLightTheme) Info() string {
	return "#458588" // blue
}

func (t *GruvboxLightTheme) Muted() string {
	return "#928374" // gray
}

func (t *GruvboxLightTheme) Emphasis() string {
	return "#d79921" // yellow
}

func (t *GruvboxLightTheme) Accent() string {
	return "#689d6a" // aqua
}

func (t *GruvboxLightTheme) SelectionBg() tcell.Color {
	return tcell.NewHexColor(0xd5c4a1) // light2
}

func (t *GruvboxLightTheme) SelectionFg() tcell.Color {
	return tcell.NewHexColor(0x3c3836) // fg
}

func (t *GruvboxLightTheme) BorderNormal() tcell.Color {
	return tcell.NewHexColor(0x928374) // gray
}

func (t *GruvboxLightTheme) BorderFocused() tcell.Color {
	return tcell.NewHexColor(0xd79921) // yellow
}

func (t *GruvboxLightTheme) AppBackground() tcell.Color {
	return tcell.NewHexColor(0xfbf1c7) // bg0
}

func (t *GruvboxLightTheme) AppForeground() tcell.Color {
	return tcell.NewHexColor(0x3c3836) // fg
}

func (t *GruvboxLightTheme) InputFieldBackground() tcell.Color {
	return tcell.NewHexColor(0xf2e5bc) // bg1 (slightly darker than bg0)
}
