package theme

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/gdamore/tcell/v2"
)

//go:embed themes/*.toml
var embeddedThemes embed.FS

func init() {
	// Load all embedded TOML themes on package initialization
	if err := LoadAllEmbeddedThemes(); err != nil {
		// Don't panic, just log to stderr
		fmt.Fprintf(os.Stderr, "Warning: failed to load TOML themes: %v\n", err)
	}
}

// TOMLTheme represents a theme loaded from a TOML file
type TOMLTheme struct {
	themeName string
	config    tomlThemeConfig
}

// tomlThemeConfig matches the structure of TOML theme files
type tomlThemeConfig struct {
	Theme struct {
		Name        string `toml:"name"`
		Description string `toml:"description"`
	} `toml:"theme"`

	Priority struct {
		P0 string `toml:"p0"`
		P1 string `toml:"p1"`
		P2 string `toml:"p2"`
		P3 string `toml:"p3"`
		P4 string `toml:"p4"`
	} `toml:"priority"`

	Status struct {
		Open       string `toml:"open"`
		InProgress string `toml:"in_progress"`
		Blocked    string `toml:"blocked"`
		Closed     string `toml:"closed"`
	} `toml:"status"`

	Dependency struct {
		Blocks         string `toml:"blocks"`
		Related        string `toml:"related"`
		ParentChild    string `toml:"parent_child"`
		DiscoveredFrom string `toml:"discovered_from"`
	} `toml:"dependency"`

	UI struct {
		Success  string `toml:"success"`
		Error    string `toml:"error"`
		Warning  string `toml:"warning"`
		Info     string `toml:"info"`
		Muted    string `toml:"muted"`
		Emphasis string `toml:"emphasis"`
		Accent   string `toml:"accent"`
	} `toml:"ui"`

	Component struct {
		SelectionBg         string `toml:"selection_bg"`
		SelectionFg         string `toml:"selection_fg"`
		BorderNormal        string `toml:"border_normal"`
		BorderFocused       string `toml:"border_focused"`
		AppBackground       string `toml:"app_background"`
		AppForeground       string `toml:"app_foreground"`
		InputFieldBackground string `toml:"input_field_background"`
	} `toml:"component"`
}

// LoadTOMLTheme loads a theme from a TOML file (embedded or external)
func LoadTOMLTheme(name string) (*TOMLTheme, error) {
	var data []byte
	var err error

	// Try loading from embedded themes first
	embeddedPath := fmt.Sprintf("themes/%s.toml", name)
	data, err = embeddedThemes.ReadFile(embeddedPath)
	if err != nil {
		// Try loading from external user themes directory
		homeDir, err := os.UserHomeDir()
		if err == nil {
			externalPath := filepath.Join(homeDir, ".config", "beads-tui", "themes", name+".toml")
			data, err = os.ReadFile(externalPath)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load theme %s: %w", name, err)
	}

	var config tomlThemeConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse theme %s: %w", name, err)
	}

	// Validate that name matches
	if config.Theme.Name != name {
		return nil, fmt.Errorf("theme name mismatch: file=%s, config=%s", name, config.Theme.Name)
	}

	return &TOMLTheme{
		themeName: name,
		config:    config,
	}, nil
}

// LoadAllEmbeddedThemes loads all TOML themes from the embedded filesystem
func LoadAllEmbeddedThemes() error {
	entries, err := embeddedThemes.ReadDir("themes")
	if err != nil {
		return fmt.Errorf("failed to read themes directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}

		// Extract theme name (remove .toml extension)
		name := strings.TrimSuffix(entry.Name(), ".toml")

		theme, err := LoadTOMLTheme(name)
		if err != nil {
			return fmt.Errorf("failed to load theme %s: %w", name, err)
		}

		Register(theme)
	}

	return nil
}

// Theme interface implementations

func (t *TOMLTheme) Name() string {
	return t.themeName
}

func (t *TOMLTheme) PriorityColors() [5]string {
	return [5]string{
		t.config.Priority.P0,
		t.config.Priority.P1,
		t.config.Priority.P2,
		t.config.Priority.P3,
		t.config.Priority.P4,
	}
}

func (t *TOMLTheme) StatusOpen() string {
	return t.config.Status.Open
}

func (t *TOMLTheme) StatusInProgress() string {
	return t.config.Status.InProgress
}

func (t *TOMLTheme) StatusBlocked() string {
	return t.config.Status.Blocked
}

func (t *TOMLTheme) StatusClosed() string {
	return t.config.Status.Closed
}

func (t *TOMLTheme) DepBlocks() string {
	return t.config.Dependency.Blocks
}

func (t *TOMLTheme) DepRelated() string {
	return t.config.Dependency.Related
}

func (t *TOMLTheme) DepParentChild() string {
	return t.config.Dependency.ParentChild
}

func (t *TOMLTheme) DepDiscoveredFrom() string {
	return t.config.Dependency.DiscoveredFrom
}

func (t *TOMLTheme) Success() string {
	return t.config.UI.Success
}

func (t *TOMLTheme) Error() string {
	return t.config.UI.Error
}

func (t *TOMLTheme) Warning() string {
	return t.config.UI.Warning
}

func (t *TOMLTheme) Info() string {
	return t.config.UI.Info
}

func (t *TOMLTheme) Muted() string {
	return t.config.UI.Muted
}

func (t *TOMLTheme) Emphasis() string {
	return t.config.UI.Emphasis
}

func (t *TOMLTheme) Accent() string {
	return t.config.UI.Accent
}

func (t *TOMLTheme) SelectionBg() tcell.Color {
	return parseHexColor(t.config.Component.SelectionBg)
}

func (t *TOMLTheme) SelectionFg() tcell.Color {
	return parseHexColor(t.config.Component.SelectionFg)
}

func (t *TOMLTheme) BorderNormal() tcell.Color {
	return parseHexColor(t.config.Component.BorderNormal)
}

func (t *TOMLTheme) BorderFocused() tcell.Color {
	return parseHexColor(t.config.Component.BorderFocused)
}

func (t *TOMLTheme) AppBackground() tcell.Color {
	return parseHexColor(t.config.Component.AppBackground)
}

func (t *TOMLTheme) AppForeground() tcell.Color {
	return parseHexColor(t.config.Component.AppForeground)
}

func (t *TOMLTheme) InputFieldBackground() tcell.Color {
	return parseHexColor(t.config.Component.InputFieldBackground)
}

// parseHexColor converts a hex color string to tcell.Color
func parseHexColor(hex string) tcell.Color {
	// Remove # prefix if present
	hex = strings.TrimPrefix(hex, "#")

	// Parse hex string to uint32
	var r, g, b uint32
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)

	// Combine into 24-bit color value
	color := (r << 16) | (g << 8) | b

	return tcell.NewHexColor(int32(color))
}
