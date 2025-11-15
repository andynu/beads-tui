package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/andy/beads-tui/internal/formatting"
	"github.com/andy/beads-tui/internal/parser"
	"github.com/andy/beads-tui/internal/state"
	"github.com/andy/beads-tui/internal/storage"
	"github.com/andy/beads-tui/internal/ui"
	"github.com/andy/beads-tui/internal/watcher"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// AppContext holds all shared application state and provides methods
// for common operations. This replaces the closure-based architecture
// in the original main() function.
type AppContext struct {
	// TUI components
	App         *tview.Application
	StatusBar   *tview.TextView
	IssueList   *tview.List
	DetailPanel *tview.TextView
	Pages       *tview.Pages

	// Data layer
	State        *state.State
	SQLiteReader *storage.SQLiteReader
	BeadsDir     string
	Watcher      *watcher.Watcher

	// UI state
	IndexToIssue      map[int]*parser.Issue
	CurrentDetailIssue *parser.Issue
	DetailPanelFocused bool
	ShowClosedIssues   bool
	MouseEnabled       bool

	// Vim navigation state
	GGPressed bool

	// Shortcut state (for multi-key sequences like "bd", "gb", etc.)
	ShortcutTimer    *time.Timer
	LastShortcutKey  rune

	// Search state
	SearchMode         bool
	SearchQuery        string
	LastSearchResults  []int
	CurrentSearchIndex int
}

// New creates a new AppContext with initialized state
func New(beadsDir string, sqliteReader *storage.SQLiteReader) *AppContext {
	return &AppContext{
		BeadsDir:     beadsDir,
		SQLiteReader: sqliteReader,
		State:        state.New(),
		IndexToIssue: make(map[int]*parser.Issue),
		MouseEnabled: true, // Default enabled
	}
}

// RefreshIssues loads issues from SQLite and updates the UI
// If preserveIssueID is provided, attempts to restore selection to that issue
func (ctx *AppContext) RefreshIssues(preserveIssueID string) {
	ctx.App.QueueUpdateDraw(func() {
		// If no explicit issue ID provided, try to preserve current selection
		if preserveIssueID == "" {
			currentIndex := ctx.IssueList.GetCurrentItem()
			if issue, ok := ctx.IndexToIssue[currentIndex]; ok {
				preserveIssueID = issue.ID
			}
		}

		// Load issues from SQLite with timeout
		loadCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		issues, err := ctx.SQLiteReader.LoadIssues(loadCtx)
		if err != nil {
			// Show error in status bar
			ctx.StatusBar.SetText(fmt.Sprintf("[red]Error loading issues: %v[-]", err))
			log.Printf("ERROR: Failed to load issues: %v", err)
			return
		}

		// Update state
		ctx.State.LoadIssues(issues)

		// Update UI on main thread
		ctx.App.QueueUpdateDraw(func() {
			// Update status bar
			// Note: This will call getStatusBarText() which needs to be a method
			ctx.UpdateStatusBar()

			// Restore selection if requested
			if preserveIssueID != "" {
				ctx.RestoreSelection(preserveIssueID)
			}
		})
	})
}

// UpdateStatusBar refreshes the status bar text using current state
func (ctx *AppContext) UpdateStatusBar() {
	text := formatting.GetStatusBarText(
		ctx.BeadsDir,
		ctx.State,
		ctx.State.GetViewMode(),
		ctx.MouseEnabled,
		ctx.DetailPanelFocused,
		ctx.ShowClosedIssues,
	)
	ctx.StatusBar.SetText(text)
}

// PopulateIssueList clears and rebuilds the issue list from current state
func (ctx *AppContext) PopulateIssueList() {
	ctx.IndexToIssue = ui.PopulateIssueList(
		ctx.IssueList,
		ctx.State,
		ctx.ShowClosedIssues,
	)
}

// ShowIssueDetails formats and displays the details for the given issue
func (ctx *AppContext) ShowIssueDetails(issue *parser.Issue) {
	ctx.CurrentDetailIssue = issue
	details := formatting.FormatIssueDetails(issue)
	ctx.DetailPanel.SetText(details)
	ctx.DetailPanel.ScrollToBeginning()
}

// UpdatePanelFocus updates the visual indicators for panel focus
func (ctx *AppContext) UpdatePanelFocus() {
	ui.UpdatePanelFocus(ctx.IssueList, ctx.DetailPanel, ctx.DetailPanelFocused)
}

// RestoreSelection attempts to restore the list selection to a specific issue ID
func (ctx *AppContext) RestoreSelection(issueID string) {
	// Search through indexToIssue to find the list index for this issue ID
	for index, issue := range ctx.IndexToIssue {
		if issue.ID == issueID {
			ctx.IssueList.SetCurrentItem(index)
			return
		}
	}
}

// GetCurrentIssue returns the currently selected issue, or nil if none selected
func (ctx *AppContext) GetCurrentIssue() *parser.Issue {
	currentIndex := ctx.IssueList.GetCurrentItem()
	issue, _ := ctx.IndexToIssue[currentIndex]
	return issue
}

// SetStatusMessage displays a temporary message in the status bar
func (ctx *AppContext) SetStatusMessage(message string) {
	ctx.StatusBar.SetText(message)
}

// SetStatusMessageTimed displays a message and clears it after a delay
func (ctx *AppContext) SetStatusMessageTimed(message string, duration time.Duration) {
	ctx.StatusBar.SetText(message)
	time.AfterFunc(duration, func() {
		ctx.App.QueueUpdateDraw(func() {
			ctx.UpdateStatusBar()
		})
	})
}

// ToggleMouseMode enables/disables mouse support
func (ctx *AppContext) ToggleMouseMode() {
	ctx.MouseEnabled = !ctx.MouseEnabled
	ctx.App.SetMouseCapture(func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
		if ctx.MouseEnabled {
			return event, action
		}
		return nil, 0 // Disable mouse
	})

	status := "enabled"
	if !ctx.MouseEnabled {
		status = "disabled"
	}
	ctx.SetStatusMessageTimed(fmt.Sprintf("[yellow]Mouse mode %s[-]", status), 2*time.Second)
}

// ToggleClosedIssues shows/hides closed issues in list view
func (ctx *AppContext) ToggleClosedIssues() {
	ctx.ShowClosedIssues = !ctx.ShowClosedIssues
	// Repopulate list will be implemented later
	// populateIssueList()
	ctx.RefreshIssues("")
}
