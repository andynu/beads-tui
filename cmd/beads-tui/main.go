package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/andy/beads-tui/internal/parser"
	"github.com/andy/beads-tui/internal/state"
	"github.com/andy/beads-tui/internal/storage"
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

	dbPath := filepath.Join(beadsDir, "beads.db")

	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: %s not found\n", dbPath)
		fmt.Fprintf(os.Stderr, "Have you initialized beads? Run: bd init\n")
		os.Exit(1)
	}

	// Open SQLite database in read-only mode
	sqliteReader, err := storage.NewSQLiteReader(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer sqliteReader.Close()

	// Initialize state
	appState := state.New()

	// Create TUI application
	app := tview.NewApplication()

	// Status bar
	statusBar := tview.NewTextView().
		SetDynamicColors(true)

	// Issue list
	issueList := tview.NewList().
		ShowSecondaryText(false).
		SetSelectedBackgroundColor(tcell.ColorNavy).
		SetSelectedTextColor(tcell.ColorWhite)
	issueList.SetBorder(true).SetTitle("Issues")

	// Track mapping from list index to issue
	indexToIssue := make(map[int]*parser.Issue)

	// Vim navigation state
	var lastKeyWasG bool
	var searchMode bool
	var searchQuery string
	var searchMatches []int
	var currentSearchIndex int

	// Helper function to render tree node recursively
	var renderTreeNode func(node *state.TreeNode, prefix string, isLast bool, currentIndex *int)
	renderTreeNode = func(node *state.TreeNode, prefix string, isLast bool, currentIndex *int) {
		issue := node.Issue

		// Determine branch characters
		var branch, continuation string
		if node.Depth == 0 {
			branch = ""
			continuation = ""
		} else {
			if isLast {
				branch = "└── "
				continuation = "    "
			} else {
				branch = "├── "
				continuation = "│   "
			}
		}

		// Get status indicator
		var statusIcon string
		switch issue.Status {
		case parser.StatusOpen:
			statusIcon = "●"
		case parser.StatusBlocked:
			statusIcon = "○"
		case parser.StatusInProgress:
			statusIcon = "◆"
		default:
			statusIcon = "·"
		}

		// Format issue line
		priorityColor := getPriorityColor(issue.Priority)
		statusColor := getStatusColor(issue.Status)
		text := fmt.Sprintf("%s%s[%s]%s[-] [%s]%s[-] [P%d] %s",
			prefix, branch, statusColor, statusIcon, priorityColor, issue.ID, issue.Priority, issue.Title)

		issueList.AddItem(text, "", 0, nil)
		indexToIssue[*currentIndex] = issue
		*currentIndex++

		// Render children
		for i, child := range node.Children {
			isLastChild := i == len(node.Children)-1
			newPrefix := prefix + continuation
			renderTreeNode(child, newPrefix, isLastChild, currentIndex)
		}
	}

	// Helper function to populate issue list from state
	populateIssueList := func() {
		// Clear and rebuild issue list
		issueList.Clear()
		indexToIssue = make(map[int]*parser.Issue)
		currentIndex := 0

		// Check view mode
		if appState.GetViewMode() == state.ViewTree {
			// Tree view
			issueList.AddItem("[cyan::b]DEPENDENCY TREE[-::-]", "", 0, nil)
			currentIndex++

			treeNodes := appState.GetTreeNodes()
			for i, node := range treeNodes {
				isLast := i == len(treeNodes)-1
				renderTreeNode(node, "", isLast, &currentIndex)
			}
		} else {
			// List view (original behavior)
			// Add ready issues
			readyIssues := appState.GetReadyIssues()
			if len(readyIssues) > 0 {
				issueList.AddItem(fmt.Sprintf("[green::b]READY (%d)[-::-]", len(readyIssues)), "", 0, nil)
				currentIndex++

				for _, issue := range readyIssues {
					priorityColor := getPriorityColor(issue.Priority)
					text := fmt.Sprintf("  [%s]●[-] %s [P%d] %s",
						priorityColor, issue.ID, issue.Priority, issue.Title)
					issueList.AddItem(text, "", 0, nil)
					indexToIssue[currentIndex] = issue
					currentIndex++
				}
			}

			// Add blocked issues
			blockedIssues := appState.GetBlockedIssues()
			if len(blockedIssues) > 0 {
				issueList.AddItem(fmt.Sprintf("\n[yellow::b]BLOCKED (%d)[-::-]", len(blockedIssues)), "", 0, nil)
				currentIndex++

				for _, issue := range blockedIssues {
					priorityColor := getPriorityColor(issue.Priority)
					text := fmt.Sprintf("  [%s]○[-] %s [P%d] %s",
						priorityColor, issue.ID, issue.Priority, issue.Title)
					issueList.AddItem(text, "", 0, nil)
					indexToIssue[currentIndex] = issue
					currentIndex++
				}
			}

			// Add in-progress issues
			inProgressIssues := appState.GetInProgressIssues()
			if len(inProgressIssues) > 0 {
				issueList.AddItem(fmt.Sprintf("\n[blue::b]IN PROGRESS (%d)[-::-]", len(inProgressIssues)), "", 0, nil)
				currentIndex++

				for _, issue := range inProgressIssues {
					priorityColor := getPriorityColor(issue.Priority)
					text := fmt.Sprintf("  [%s]◆[-] %s [P%d] %s",
						priorityColor, issue.ID, issue.Priority, issue.Title)
					issueList.AddItem(text, "", 0, nil)
					indexToIssue[currentIndex] = issue
					currentIndex++
				}
			}
		}
	}

	// Function to load and display issues (for async updates after app starts)
	refreshIssues := func() {
		// Load issues from SQLite with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		issues, err := sqliteReader.LoadIssues(ctx)
		if err != nil {
			// Show error in status bar
			app.QueueUpdateDraw(func() {
				statusBar.SetText(fmt.Sprintf("[red]Error loading issues: %v[-]", err))
			})
			return
		}

		// Update state
		appState.LoadIssues(issues)

		// Update UI on main thread
		app.QueueUpdateDraw(func() {
			// Update status bar
			viewModeStr := "List"
			if appState.GetViewMode() == state.ViewTree {
				viewModeStr = "Tree"
			}
			statusBar.SetText(fmt.Sprintf("[yellow]Beads TUI[-] - %s (%d issues) [SQLite] [%s View] [Press ? for help, q to quit, r to refresh, t to toggle view]",
				beadsDir, len(issues), viewModeStr))

			populateIssueList()
		})
	}

	// Initial load (before app starts, no QueueUpdateDraw)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	issues, err := sqliteReader.LoadIssues(ctx)
	cancel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading issues: %v\n", err)
		os.Exit(1)
	}
	appState.LoadIssues(issues)
	viewModeStr := "List"
	if appState.GetViewMode() == state.ViewTree {
		viewModeStr = "Tree"
	}
	statusBar.SetText(fmt.Sprintf("[yellow]Beads TUI[-] - %s (%d issues) [SQLite] [%s View] [Press ? for help, q to quit, r to refresh, t to toggle view]",
		beadsDir, len(issues), viewModeStr))
	populateIssueList()

	// Set up filesystem watcher on the database
	fileWatcher, err := watcher.New(dbPath, 200*time.Millisecond, refreshIssues)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to set up database watcher: %v\n", err)
		fmt.Fprintf(os.Stderr, "Live updates will not work. Press 'r' to manually refresh.\n")
	} else {
		if err := fileWatcher.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to start database watcher: %v\n", err)
		}
		defer fileWatcher.Stop()
	}

	// Detail panel
	detailPanel := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	detailPanel.SetBorder(true).SetTitle("Details")
	detailPanel.SetText("[yellow]Navigate to an issue to view details[-]")

	// Function to show issue details
	showIssueDetails := func(issue *parser.Issue) {
		details := formatIssueDetails(issue)
		detailPanel.SetText(details)
		detailPanel.ScrollToBeginning()
	}

	// Set up change handler to auto-show details on selection change
	issueList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		// Check if the selected item is an issue (not a header)
		if issue, ok := indexToIssue[index]; ok {
			showIssueDetails(issue)
		}
	})

	// Layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(statusBar, 1, 0, false).
		AddItem(tview.NewFlex().
			AddItem(issueList, 0, 1, true).
			AddItem(detailPanel, 0, 2, false),
			0, 1, true)

	// Helper function to perform search
	performSearch := func(query string) {
		searchMatches = nil
		currentSearchIndex = -1

		if query == "" {
			return
		}

		// Search through all items in the list
		for i := 0; i < issueList.GetItemCount(); i++ {
			mainText, _ := issueList.GetItemText(i)
			// Simple case-insensitive substring search
			if len(mainText) > 0 && containsCaseInsensitive(mainText, query) {
				searchMatches = append(searchMatches, i)
			}
		}

		// Jump to first match if any
		if len(searchMatches) > 0 {
			currentSearchIndex = 0
			issueList.SetCurrentItem(searchMatches[0])
			statusBar.SetText(fmt.Sprintf("[yellow]Search:[-] %s [%d/%d matches] [Press n/N for next/prev, ESC to exit search]",
				query, 1, len(searchMatches)))
		} else {
			statusBar.SetText(fmt.Sprintf("[red]Search:[-] %s [No matches]", query))
		}
	}

	// Helper function for next search result
	nextSearchMatch := func() {
		if len(searchMatches) == 0 {
			return
		}
		currentSearchIndex = (currentSearchIndex + 1) % len(searchMatches)
		issueList.SetCurrentItem(searchMatches[currentSearchIndex])
		statusBar.SetText(fmt.Sprintf("[yellow]Search:[-] %s [%d/%d matches] [Press n/N for next/prev, ESC to exit search]",
			searchQuery, currentSearchIndex+1, len(searchMatches)))
	}

	// Helper function for previous search result
	prevSearchMatch := func() {
		if len(searchMatches) == 0 {
			return
		}
		currentSearchIndex--
		if currentSearchIndex < 0 {
			currentSearchIndex = len(searchMatches) - 1
		}
		issueList.SetCurrentItem(searchMatches[currentSearchIndex])
		statusBar.SetText(fmt.Sprintf("[yellow]Search:[-] %s [%d/%d matches] [Press n/N for next/prev, ESC to exit search]",
			searchQuery, currentSearchIndex+1, len(searchMatches)))
	}

	// Set up key bindings
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Handle search mode
		if searchMode {
			switch event.Key() {
			case tcell.KeyEscape:
				searchMode = false
				searchQuery = ""
				viewModeStr := "List"
				if appState.GetViewMode() == state.ViewTree {
					viewModeStr = "Tree"
				}
				statusBar.SetText(fmt.Sprintf("[yellow]Beads TUI[-] - %s (%d issues) [SQLite] [%s View] [Press ? for help, q to quit, r to refresh, t to toggle view]",
					beadsDir, len(appState.GetAllIssues()), viewModeStr))
				return nil
			case tcell.KeyEnter:
				performSearch(searchQuery)
				searchMode = false
				return nil
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				if len(searchQuery) > 0 {
					searchQuery = searchQuery[:len(searchQuery)-1]
					statusBar.SetText(fmt.Sprintf("[yellow]Search:[-] %s_", searchQuery))
				}
				return nil
			case tcell.KeyRune:
				searchQuery += string(event.Rune())
				statusBar.SetText(fmt.Sprintf("[yellow]Search:[-] %s_", searchQuery))
				return nil
			}
			return nil
		}

		// Normal mode key bindings
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
			case 'g':
				if lastKeyWasG {
					// gg - jump to top
					issueList.SetCurrentItem(0)
					lastKeyWasG = false
					return nil
				}
				lastKeyWasG = true
				return nil
			case 'G':
				// G - jump to bottom
				issueList.SetCurrentItem(issueList.GetItemCount() - 1)
				lastKeyWasG = false
				return nil
			case '/':
				// Start search mode
				searchMode = true
				searchQuery = ""
				statusBar.SetText("[yellow]Search:[-] _")
				return nil
			case 'n':
				// Next search result
				nextSearchMatch()
				return nil
			case 'N':
				// Previous search result
				prevSearchMatch()
				return nil
			case 't':
				// Toggle view mode
				appState.ToggleViewMode()
				viewModeStr := "List"
				if appState.GetViewMode() == state.ViewTree {
					viewModeStr = "Tree"
				}
				statusBar.SetText(fmt.Sprintf("[yellow]Beads TUI[-] - %s (%d issues) [SQLite] [%s View] [Press ? for help, q to quit, r to refresh, t to toggle view]",
					beadsDir, len(appState.GetAllIssues()), viewModeStr))
				populateIssueList()
				return nil
			default:
				// Reset g flag if any other key is pressed
				lastKeyWasG = false
			}
		case tcell.KeyEscape:
			// Clear search on ESC
			if len(searchMatches) > 0 {
				searchMatches = nil
				currentSearchIndex = -1
				viewModeStr := "List"
				if appState.GetViewMode() == state.ViewTree {
					viewModeStr = "Tree"
				}
				statusBar.SetText(fmt.Sprintf("[yellow]Beads TUI[-] - %s (%d issues) [SQLite] [%s View] [Press ? for help, q to quit, r to refresh, t to toggle view]",
					beadsDir, len(appState.GetAllIssues()), viewModeStr))
				return nil
			}
		default:
			lastKeyWasG = false
		}
		return event
	})

	// Run application
	// Note: Mouse disabled to allow terminal text selection (tui-p62)
	if err := app.SetRoot(flex, true).Run(); err != nil {
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

// formatIssueDetails formats full issue metadata for display
func formatIssueDetails(issue *parser.Issue) string {
	var result string

	// Header
	priorityColor := getPriorityColor(issue.Priority)
	statusColor := getStatusColor(issue.Status)
	typeIcon := getTypeIcon(issue.IssueType)

	result += fmt.Sprintf("[::b]%s %s[-::-]\n", typeIcon, issue.Title)
	result += fmt.Sprintf("[gray]ID:[-] %s  ", issue.ID)
	result += fmt.Sprintf("[%s]P%d[-]  ", priorityColor, issue.Priority)
	result += fmt.Sprintf("[%s]%s[-]\n\n", statusColor, issue.Status)

	// Description
	if issue.Description != "" {
		result += "[yellow::b]Description:[-::-]\n"
		result += issue.Description + "\n\n"
	}

	// Design notes
	if issue.Design != "" {
		result += "[yellow::b]Design:[-::-]\n"
		result += issue.Design + "\n\n"
	}

	// Acceptance criteria
	if issue.AcceptanceCriteria != "" {
		result += "[yellow::b]Acceptance Criteria:[-::-]\n"
		result += issue.AcceptanceCriteria + "\n\n"
	}

	// Notes
	if issue.Notes != "" {
		result += "[yellow::b]Notes:[-::-]\n"
		result += issue.Notes + "\n\n"
	}

	// Dependencies
	if len(issue.Dependencies) > 0 {
		result += "[yellow::b]Dependencies:[-::-]\n"
		for _, dep := range issue.Dependencies {
			result += fmt.Sprintf("  • [%s]%s[-] %s\n",
				getDependencyColor(dep.Type), dep.Type, dep.DependsOnID)
		}
		result += "\n"
	}

	// Labels
	if len(issue.Labels) > 0 {
		result += "[yellow::b]Labels:[-::-] "
		for i, label := range issue.Labels {
			if i > 0 {
				result += ", "
			}
			result += fmt.Sprintf("[cyan]%s[-]", label)
		}
		result += "\n\n"
	}

	// Metadata
	result += "[yellow::b]Metadata:[-::-]\n"
	result += fmt.Sprintf("  Created: %s\n", issue.CreatedAt.Format("2006-01-02 15:04"))
	result += fmt.Sprintf("  Updated: %s\n", issue.UpdatedAt.Format("2006-01-02 15:04"))

	if issue.ClosedAt != nil {
		result += fmt.Sprintf("  Closed: %s\n", issue.ClosedAt.Format("2006-01-02 15:04"))
	}

	if issue.Assignee != "" {
		result += fmt.Sprintf("  Assignee: %s\n", issue.Assignee)
	}

	if issue.EstimatedMinutes != nil {
		hours := *issue.EstimatedMinutes / 60
		mins := *issue.EstimatedMinutes % 60
		result += fmt.Sprintf("  Estimated: %dh %dm\n", hours, mins)
	}

	if issue.ExternalRef != nil {
		result += fmt.Sprintf("  External Ref: %s\n", *issue.ExternalRef)
	}

	// Comments
	if len(issue.Comments) > 0 {
		result += "\n[yellow::b]Comments:[-::-]\n"
		for _, comment := range issue.Comments {
			result += fmt.Sprintf("  [cyan]%s[-] (%s):\n", comment.Author, comment.CreatedAt.Format("2006-01-02 15:04"))
			result += fmt.Sprintf("    %s\n", comment.Text)
		}
	}

	return result
}

// getStatusColor returns color for status
func getStatusColor(status parser.Status) string {
	switch status {
	case parser.StatusOpen:
		return "green"
	case parser.StatusInProgress:
		return "blue"
	case parser.StatusBlocked:
		return "yellow"
	case parser.StatusClosed:
		return "gray"
	default:
		return "white"
	}
}

// getTypeIcon returns icon for issue type
func getTypeIcon(issueType parser.IssueType) string {
	switch issueType {
	case parser.TypeBug:
		return "🐛"
	case parser.TypeFeature:
		return "✨"
	case parser.TypeTask:
		return "📋"
	case parser.TypeEpic:
		return "🎯"
	case parser.TypeChore:
		return "🔧"
	default:
		return "•"
	}
}

// getDependencyColor returns color for dependency type
func getDependencyColor(depType parser.DependencyType) string {
	switch depType {
	case parser.DepBlocks:
		return "red"
	case parser.DepRelated:
		return "blue"
	case parser.DepParentChild:
		return "green"
	case parser.DepDiscoveredFrom:
		return "yellow"
	default:
		return "white"
	}
}

// containsCaseInsensitive checks if s contains substr (case-insensitive)
func containsCaseInsensitive(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return len(s) >= len(substr) && indexCaseInsensitive(s, substr) >= 0
}

// toLower converts string to lowercase
func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}

// indexCaseInsensitive finds the index of substr in s (case-insensitive)
func indexCaseInsensitive(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
