package main

import (
	"fmt"
	"sort"

	"github.com/andy/beads-tui/internal/theme"
	_ "github.com/andy/beads-tui/internal/theme" // Import to register themes
)

func main() {
	themes := theme.List()
	sort.Strings(themes)

	fmt.Printf("Available themes (%d):\n", len(themes))
	for _, name := range themes {
		t := theme.Get(name)
		if t != nil {
			fmt.Printf("  - %s\n", name)
		}
	}

	// Test switching to each theme
	fmt.Println("\nTesting theme switching:")
	for _, name := range themes {
		err := theme.SetCurrent(name)
		if err != nil {
			fmt.Printf("  ✗ %s: %v\n", name, err)
		} else {
			current := theme.Current()
			fmt.Printf("  ✓ %s (P0 color: %s)\n", name, current.PriorityColors()[0])
		}
	}
}
