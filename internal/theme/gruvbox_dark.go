package theme

import "github.com/gdamore/tcell/v2"

// GruvboxDarkTheme provides the Gruvbox Dark color scheme
// Based on: https://github.com/morhetz/gruvbox
type GruvboxDarkTheme struct{}

func init() {
	Register(&GruvboxDarkTheme{})
}

func (t *GruvboxDarkTheme) Name() string {
	return "gruvbox-dark"
}

func (t *GruvboxDarkTheme) PriorityColors() [5]string {
	return [5]string{
		"#fb4934", // P0: bright red
		"#fe8019", // P1: bright orange
		"#83a598", // P2: bright blue
		"#928374", // P3: gray
		"#665c54", // P4: dark gray
	}
}

func (t *GruvboxDarkTheme) StatusOpen() string {
	return "#b8bb26" // bright green
}

func (t *GruvboxDarkTheme) StatusInProgress() string {
	return "#83a598" // bright blue
}

func (t *GruvboxDarkTheme) StatusBlocked() string {
	return "#fabd2f" // bright yellow
}

func (t *GruvboxDarkTheme) StatusClosed() string {
	return "#928374" // gray
}

func (t *GruvboxDarkTheme) DepBlocks() string {
	return "#fb4934" // bright red
}

func (t *GruvboxDarkTheme) DepRelated() string {
	return "#83a598" // bright blue
}

func (t *GruvboxDarkTheme) DepParentChild() string {
	return "#b8bb26" // bright green
}

func (t *GruvboxDarkTheme) DepDiscoveredFrom() string {
	return "#fabd2f" // bright yellow
}

func (t *GruvboxDarkTheme) Success() string {
	return "#b8bb26" // bright green
}

func (t *GruvboxDarkTheme) Error() string {
	return "#fb4934" // bright red
}

func (t *GruvboxDarkTheme) Warning() string {
	return "#fabd2f" // bright yellow
}

func (t *GruvboxDarkTheme) Info() string {
	return "#83a598" // bright blue
}

func (t *GruvboxDarkTheme) Muted() string {
	return "#928374" // gray
}

func (t *GruvboxDarkTheme) Emphasis() string {
	return "#fabd2f" // bright yellow
}

func (t *GruvboxDarkTheme) Accent() string {
	return "#8ec07c" // bright aqua
}

func (t *GruvboxDarkTheme) SelectionBg() tcell.Color {
	return tcell.NewHexColor(0x504945) // dark2
}

func (t *GruvboxDarkTheme) SelectionFg() tcell.Color {
	return tcell.NewHexColor(0xebdbb2) // fg
}

func (t *GruvboxDarkTheme) BorderNormal() tcell.Color {
	return tcell.NewHexColor(0x928374) // gray
}

func (t *GruvboxDarkTheme) BorderFocused() tcell.Color {
	return tcell.NewHexColor(0xfabd2f) // bright yellow
}

func (t *GruvboxDarkTheme) AppBackground() tcell.Color {
	return tcell.NewHexColor(0x282828) // bg0
}

func (t *GruvboxDarkTheme) AppForeground() tcell.Color {
	return tcell.NewHexColor(0xebdbb2) // fg
}
