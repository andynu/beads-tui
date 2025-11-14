package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/andy/beads-tui/internal/parser"
	"github.com/andy/beads-tui/internal/state"
	"github.com/andy/beads-tui/internal/watcher"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	// Find .beads directory
	beadsDir, err := findBeadsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	jsonlPath := filepath.Join(beadsDir, "issues.jsonl")

	// Check if JSONL file exists
	if _, err := os.Stat(jsonlPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: %s not found\n", jsonlPath)
		fmt.Fprintf(os.Stderr, "Have you initialized beads? Run: bd init\n")
		os.Exit(1)
	}

	// Initialize state
	appState := state.New()

	// Create TUI application
	app := tview.NewApplication()

	// Status bar
	statusBar := tview.NewTextView().
		SetDynamicColors(true)

	// Issue list
	issueList := tview.NewList().
		ShowSecondaryText(false)
	issueList.SetBorder(true).SetTitle("Issues")

	// Function to load and display issues
	refreshIssues := func() {
		// Parse issues
		issues, err := parser.ParseFile(jsonlPath)
		if err != nil {
			// Show error in status bar
			app.QueueUpdateDraw(func() {
				statusBar.SetText(fmt.Sprintf("[red]Error parsing issues: %v[-]", err))
			})
			return
		}

		// Update state
		appState.LoadIssues(issues)

		// Update UI on main thread
		app.QueueUpdateDraw(func() {
			// Update status bar
			statusBar.SetText(fmt.Sprintf("[yellow]Beads TUI[-] - %s (%d issues) [Press ? for help, q to quit, r to refresh]",
				beadsDir, len(issues)))

			// Clear and rebuild issue list
			issueList.Clear()

			// Add ready issues
			readyIssues := appState.GetReadyIssues()
			if len(readyIssues) > 0 {
				issueList.AddItem(fmt.Sprintf("[green::b]READY (%d)[-::-]", len(readyIssues)), "", 0, nil)
				for _, issue := range readyIssues {
					priorityColor := getPriorityColor(issue.Priority)
					text := fmt.Sprintf("  [%s]●[-] %s [P%d] %s",
						priorityColor, issue.ID, issue.Priority, issue.Title)
					issueList.AddItem(text, "", 0, nil)
				}
			}

			// Add blocked issues
			blockedIssues := appState.GetBlockedIssues()
			if len(blockedIssues) > 0 {
				issueList.AddItem(fmt.Sprintf("\n[yellow::b]BLOCKED (%d)[-::-]", len(blockedIssues)), "", 0, nil)
				for _, issue := range blockedIssues {
					priorityColor := getPriorityColor(issue.Priority)
					text := fmt.Sprintf("  [%s]○[-] %s [P%d] %s",
						priorityColor, issue.ID, issue.Priority, issue.Title)
					issueList.AddItem(text, "", 0, nil)
				}
			}

			// Add in-progress issues
			inProgressIssues := appState.GetInProgressIssues()
			if len(inProgressIssues) > 0 {
				issueList.AddItem(fmt.Sprintf("\n[blue::b]IN PROGRESS (%d)[-::-]", len(inProgressIssues)), "", 0, nil)
				for _, issue := range inProgressIssues {
					priorityColor := getPriorityColor(issue.Priority)
					text := fmt.Sprintf("  [%s]◆[-] %s [P%d] %s",
						priorityColor, issue.ID, issue.Priority, issue.Title)
					issueList.AddItem(text, "", 0, nil)
				}
			}
		})
	}

	// Initial load
	refreshIssues()

	// Set up filesystem watcher
	fileWatcher, err := watcher.New(jsonlPath, 200*time.Millisecond, refreshIssues)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to set up file watcher: %v\n", err)
		fmt.Fprintf(os.Stderr, "Live updates will not work. Press 'r' to manually refresh.\n")
	} else {
		if err := fileWatcher.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to start file watcher: %v\n", err)
		}
		defer fileWatcher.Stop()
	}

	// Detail panel
	detailPanel := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	detailPanel.SetBorder(true).SetTitle("Details")
	detailPanel.SetText("[yellow]Select an issue to view details[-]")

	// Layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(statusBar, 1, 0, false).
		AddItem(tview.NewFlex().
			AddItem(issueList, 0, 1, true).
			AddItem(detailPanel, 0, 2, false),
			0, 1, true)

	// Set up key bindings
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil
			case 'r':
				// Manual refresh
				refreshIssues()
				return nil
			case 'j':
				// Down - simulate down arrow
				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
			case 'k':
				// Up - simulate up arrow
				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
			}
		}
		return event
	})

	// Run application
	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

// findBeadsDir searches for .beads directory in current and parent directories
func findBeadsDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		beadsDir := filepath.Join(dir, ".beads")
		if info, err := os.Stat(beadsDir); err == nil && info.IsDir() {
			return beadsDir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf(".beads directory not found")
}

// getPriorityColor returns a color code for the given priority
func getPriorityColor(priority int) string {
	switch priority {
	case 0:
		return "red"
	case 1:
		return "orange"
	case 2:
		return "white"
	case 3:
		return "gray"
	case 4:
		return "darkgray"
	default:
		return "white"
	}
}
