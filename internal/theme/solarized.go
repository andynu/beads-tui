package theme

import "github.com/gdamore/tcell/v2"

// SolarizedDarkTheme provides the Solarized Dark color scheme
// Based on: https://ethanschoonover.com/solarized/
type SolarizedDarkTheme struct{}

func init() {
	Register(&SolarizedDarkTheme{})
}

func (t *SolarizedDarkTheme) Name() string {
	return "solarized-dark"
}

func (t *SolarizedDarkTheme) PriorityColors() [5]string {
	return [5]string{
		"#dc322f", // P0: red
		"#cb4b16", // P1: orange
		"#268bd2", // P2: blue
		"#586e75", // P3: base01
		"#073642", // P4: base02
	}
}

func (t *SolarizedDarkTheme) StatusOpen() string {
	return "#859900" // green
}

func (t *SolarizedDarkTheme) StatusInProgress() string {
	return "#268bd2" // blue
}

func (t *SolarizedDarkTheme) StatusBlocked() string {
	return "#b58900" // yellow
}

func (t *SolarizedDarkTheme) StatusClosed() string {
	return "#586e75" // base01
}

func (t *SolarizedDarkTheme) DepBlocks() string {
	return "#dc322f" // red
}

func (t *SolarizedDarkTheme) DepRelated() string {
	return "#268bd2" // blue
}

func (t *SolarizedDarkTheme) DepParentChild() string {
	return "#859900" // green
}

func (t *SolarizedDarkTheme) DepDiscoveredFrom() string {
	return "#b58900" // yellow
}

func (t *SolarizedDarkTheme) Success() string {
	return "#859900" // green
}

func (t *SolarizedDarkTheme) Error() string {
	return "#dc322f" // red
}

func (t *SolarizedDarkTheme) Warning() string {
	return "#b58900" // yellow
}

func (t *SolarizedDarkTheme) Info() string {
	return "#2aa198" // cyan
}

func (t *SolarizedDarkTheme) Muted() string {
	return "#586e75" // base01
}

func (t *SolarizedDarkTheme) Emphasis() string {
	return "#b58900" // yellow
}

func (t *SolarizedDarkTheme) Accent() string {
	return "#2aa198" // cyan
}

func (t *SolarizedDarkTheme) SelectionBg() tcell.Color {
	return tcell.NewHexColor(0x073642) // base02
}

func (t *SolarizedDarkTheme) SelectionFg() tcell.Color {
	return tcell.NewHexColor(0x93a1a1) // base1
}

func (t *SolarizedDarkTheme) BorderNormal() tcell.Color {
	return tcell.NewHexColor(0x586e75) // base01
}

func (t *SolarizedDarkTheme) BorderFocused() tcell.Color {
	return tcell.NewHexColor(0x2aa198) // cyan
}

func (t *SolarizedDarkTheme) AppBackground() tcell.Color {
	return tcell.NewHexColor(0x002b36) // base03
}

func (t *SolarizedDarkTheme) AppForeground() tcell.Color {
	return tcell.NewHexColor(0x839496) // base0
}

// SolarizedLightTheme provides the Solarized Light color scheme
type SolarizedLightTheme struct{}

func init() {
	Register(&SolarizedLightTheme{})
}

func (t *SolarizedLightTheme) Name() string {
	return "solarized-light"
}

func (t *SolarizedLightTheme) PriorityColors() [5]string {
	return [5]string{
		"#dc322f", // P0: red
		"#cb4b16", // P1: orange
		"#268bd2", // P2: blue
		"#93a1a1", // P3: base1
		"#839496", // P4: base0
	}
}

func (t *SolarizedLightTheme) StatusOpen() string {
	return "#859900" // green
}

func (t *SolarizedLightTheme) StatusInProgress() string {
	return "#268bd2" // blue
}

func (t *SolarizedLightTheme) StatusBlocked() string {
	return "#b58900" // yellow
}

func (t *SolarizedLightTheme) StatusClosed() string {
	return "#93a1a1" // base1
}

func (t *SolarizedLightTheme) DepBlocks() string {
	return "#dc322f" // red
}

func (t *SolarizedLightTheme) DepRelated() string {
	return "#268bd2" // blue
}

func (t *SolarizedLightTheme) DepParentChild() string {
	return "#859900" // green
}

func (t *SolarizedLightTheme) DepDiscoveredFrom() string {
	return "#b58900" // yellow
}

func (t *SolarizedLightTheme) Success() string {
	return "#859900" // green
}

func (t *SolarizedLightTheme) Error() string {
	return "#dc322f" // red
}

func (t *SolarizedLightTheme) Warning() string {
	return "#b58900" // yellow
}

func (t *SolarizedLightTheme) Info() string {
	return "#2aa198" // cyan
}

func (t *SolarizedLightTheme) Muted() string {
	return "#93a1a1" // base1
}

func (t *SolarizedLightTheme) Emphasis() string {
	return "#b58900" // yellow
}

func (t *SolarizedLightTheme) Accent() string {
	return "#2aa198" // cyan
}

func (t *SolarizedLightTheme) SelectionBg() tcell.Color {
	return tcell.NewHexColor(0xeee8d5) // base2
}

func (t *SolarizedLightTheme) SelectionFg() tcell.Color {
	return tcell.NewHexColor(0x586e75) // base01
}

func (t *SolarizedLightTheme) BorderNormal() tcell.Color {
	return tcell.NewHexColor(0x93a1a1) // base1
}

func (t *SolarizedLightTheme) BorderFocused() tcell.Color {
	return tcell.NewHexColor(0x2aa198) // cyan
}

func (t *SolarizedLightTheme) AppBackground() tcell.Color {
	return tcell.NewHexColor(0xfdf6e3) // base3
}

func (t *SolarizedLightTheme) AppForeground() tcell.Color {
	return tcell.NewHexColor(0x657b83) // base00
}
