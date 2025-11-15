package theme

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
)

// Theme defines the color scheme for the TUI
type Theme interface {
	// Name returns the theme's display name
	Name() string

	// Priority colors (P0-P4)
	PriorityColors() [5]string

	// Status colors (tview color strings)
	StatusOpen() string
	StatusInProgress() string
	StatusBlocked() string
	StatusClosed() string

	// Dependency type colors (tview color strings)
	DepBlocks() string
	DepRelated() string
	DepParentChild() string
	DepDiscoveredFrom() string

	// UI semantic colors (tview color strings)
	Success() string
	Error() string
	Warning() string
	Info() string
	Muted() string
	Emphasis() string
	Accent() string

	// Component colors (tcell.Color for tview style properties)
	SelectionBg() tcell.Color
	SelectionFg() tcell.Color
	BorderNormal() tcell.Color
	BorderFocused() tcell.Color
}

var (
	registry      = make(map[string]Theme)
	currentTheme  Theme
	registryMutex sync.RWMutex
)

// Register adds a theme to the global registry
func Register(t Theme) {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	registry[t.Name()] = t

	// Set as current if it's the first theme registered
	if currentTheme == nil {
		currentTheme = t
	}
}

// SetCurrent switches to the named theme
func SetCurrent(name string) error {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	t, exists := registry[name]
	if !exists {
		return fmt.Errorf("theme not found: %s", name)
	}

	currentTheme = t
	return nil
}

// Current returns the currently active theme
func Current() Theme {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	return currentTheme
}

// List returns the names of all registered themes
func List() []string {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// Get returns the theme with the given name, or nil if not found
func Get(name string) Theme {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	return registry[name]
}
