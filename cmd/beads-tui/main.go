package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/andy/beads-tui/internal/parser"
	"github.com/andy/beads-tui/internal/state"
	"github.com/andy/beads-tui/internal/storage"
	"github.com/andy/beads-tui/internal/watcher"
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	// Parse command line flags
	debugMode := flag.Bool("debug", false, "Enable debug logging to file")
	flag.Parse()

	// Set up logging
	var logFile *os.File
	if *debugMode {
		logDir := filepath.Join(os.Getenv("HOME"), ".beads-tui")
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create log directory: %v\n", err)
		} else {
			logPath := filepath.Join(logDir, fmt.Sprintf("debug-%s.log", time.Now().Format("2006-01-02-15-04-05")))
			var err error
			logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to open log file: %v\n", err)
			} else {
				log.SetOutput(logFile)
				log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
				defer logFile.Close()
				log.Printf("=== beads-tui started in debug mode ===")
				log.Printf("Log file: %s", logPath)
				fmt.Fprintf(os.Stderr, "Debug logging enabled: %s\n", logPath)
			}
		}
	} else {
		// Disable logging completely when not in debug mode
		log.SetOutput(io.Discard)
		log.SetFlags(0)
	}

	log.Printf("Finding .beads directory")
	// Find .beads directory
	beadsDir, err := findBeadsDir()
	if err != nil {
		log.Printf("ERROR: Failed to find .beads directory: %v", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	log.Printf("Found .beads directory: %s", beadsDir)

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
		SetSelectedBackgroundColor(tcell.ColorDarkCyan).
		SetSelectedTextColor(tcell.ColorBlack)
	issueList.SetBorder(true).SetTitle("Issues")

	// Track mapping from list index to issue
	indexToIssue := make(map[int]*parser.Issue)

	// Vim navigation state
	var lastKeyWasG bool
	var searchMode bool
	var searchQuery string
	var searchMatches []int
	var currentSearchIndex int

	// Mouse mode state (default: enabled)
	var mouseEnabled = true

	// Panel focus state (true = detail panel, false = issue list)
	var detailPanelFocused bool

	// Track currently displayed issue in detail panel (for clipboard copy)
	var currentDetailIssue *parser.Issue

	// Helper function to generate status bar text
	getStatusBarText := func() string {
		viewModeStr := "List"
		if appState.GetViewMode() == state.ViewTree {
			viewModeStr = "Tree"
		}
		mouseStr := "OFF"
		if mouseEnabled {
			mouseStr = "ON"
		}
		focusStr := "List"
		if detailPanelFocused {
			focusStr = "Details"
		}

		// Count visible issues after filtering
		visibleCount := len(appState.GetReadyIssues()) + len(appState.GetBlockedIssues()) + len(appState.GetInProgressIssues()) + len(appState.GetClosedIssues())

		filterText := ""
		if appState.HasActiveFilters() {
			filterText = fmt.Sprintf(" [Filters: %s]", appState.GetActiveFilters())
		}

		return fmt.Sprintf("[yellow]Beads TUI[-] - %s (%d issues)%s [SQLite] [%s View] [Mouse: %s] [Focus: %s] [Press ? for help, f for quick filter]",
			beadsDir, visibleCount, filterText, viewModeStr, mouseStr, focusStr)
	}

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

			// Add closed issues
			closedIssues := appState.GetClosedIssues()
			if len(closedIssues) > 0 {
				issueList.AddItem(fmt.Sprintf("\n[gray::b]CLOSED (%d)[-::-]", len(closedIssues)), "", 0, nil)
				currentIndex++

				for _, issue := range closedIssues {
					priorityColor := getPriorityColor(issue.Priority)
					text := fmt.Sprintf("  [%s]✓[-] %s [P%d] %s",
						priorityColor, issue.ID, issue.Priority, issue.Title)
					issueList.AddItem(text, "", 0, nil)
					indexToIssue[currentIndex] = issue
					currentIndex++
				}
			}
		}
	}

	// Function to load and display issues (for async updates after app starts)
	// preserveIssueID: if provided, attempt to restore selection to this issue after refresh
	refreshIssues := func(preserveIssueID ...string) {
		log.Printf("REFRESH: Starting issue refresh")
		var targetIssueID string
		if len(preserveIssueID) > 0 {
			targetIssueID = preserveIssueID[0]
			log.Printf("REFRESH: Will attempt to preserve selection on issue: %s", targetIssueID)
		}

		// Load issues from SQLite with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		log.Printf("REFRESH: Loading issues from SQLite (timeout=5s)")
		issues, err := sqliteReader.LoadIssues(ctx)
		if err != nil {
			log.Printf("REFRESH ERROR: Failed to load issues: %v", err)
			// Show error in status bar
			app.QueueUpdateDraw(func() {
				statusBar.SetText(fmt.Sprintf("[red]Error loading issues: %v[-]", err))
			})
			return
		}
		log.Printf("REFRESH: Loaded %d issues from database", len(issues))

		// Update state
		appState.LoadIssues(issues)
		log.Printf("REFRESH: Updated app state")

		// Update UI on main thread
		log.Printf("REFRESH: Queueing UI update")
		app.QueueUpdateDraw(func() {
			log.Printf("REFRESH: UI update executing")
			// Update status bar
			statusBar.SetText(getStatusBarText())

			populateIssueList()

			// Restore selection if requested
			if targetIssueID != "" {
				log.Printf("REFRESH: Searching for issue %s to restore selection", targetIssueID)
				for idx, issue := range indexToIssue {
					if issue.ID == targetIssueID {
						log.Printf("REFRESH: Found issue %s at index %d, restoring selection", targetIssueID, idx)
						issueList.SetCurrentItem(idx)
						break
					}
				}
			}

			log.Printf("REFRESH: UI update complete")
		})
		log.Printf("REFRESH: Issue refresh complete")
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
	statusBar.SetText(getStatusBarText())
	populateIssueList()

	// Set up filesystem watcher on the database
	log.Printf("Setting up file watcher on: %s", dbPath)
	fileWatcher, err := watcher.New(dbPath, 200*time.Millisecond, func() {
		log.Printf("WATCHER: File change detected, triggering refresh")
		refreshIssues()
	})
	if err != nil {
		log.Printf("WATCHER ERROR: Failed to create watcher: %v", err)
		fmt.Fprintf(os.Stderr, "Warning: failed to set up database watcher: %v\n", err)
		fmt.Fprintf(os.Stderr, "Live updates will not work. Press 'r' to manually refresh.\n")
	} else {
		if err := fileWatcher.Start(); err != nil {
			log.Printf("WATCHER ERROR: Failed to start watcher: %v", err)
			fmt.Fprintf(os.Stderr, "Warning: failed to start database watcher: %v\n", err)
		} else {
			log.Printf("WATCHER: File watcher started successfully")
		}
		defer func() {
			log.Printf("WATCHER: Stopping file watcher")
			fileWatcher.Stop()
		}()
	}

	// Detail panel
	detailPanel := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	detailPanel.SetBorder(true).SetTitle("Details")
	detailPanel.SetText("[yellow]Navigate to an issue to view details[-]")

	// Add mouse click handler for copying issue ID
	detailPanel.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseLeftClick && currentDetailIssue != nil {
			// Get click position
			_, y := event.Position()
			// Get the detail panel's position
			_, panelY, _, _ := detailPanel.GetInnerRect()

			// Calculate relative position within the text view
			relativeY := y - panelY

			// The issue ID is on line 2 (0-indexed line 1) of the detail text
			// Format: "ID: <issue-id>  P<priority>  <status>"
			if relativeY == 1 && currentDetailIssue != nil {
				// Copy issue ID to clipboard
				err := clipboard.WriteAll(currentDetailIssue.ID)
				if err != nil {
					log.Printf("CLIPBOARD ERROR: Failed to copy to clipboard: %v", err)
					statusBar.SetText(fmt.Sprintf("[red]Failed to copy: %v[-]", err))
				} else {
					log.Printf("CLIPBOARD: Copied issue ID to clipboard: %s", currentDetailIssue.ID)
					statusBar.SetText(fmt.Sprintf("[green]✓ Copied %s to clipboard[-]", currentDetailIssue.ID))
					// Clear message after 2 seconds
					time.AfterFunc(2*time.Second, func() {
						app.QueueUpdateDraw(func() {
							statusBar.SetText(getStatusBarText())
						})
					})
				}
			}
		}
		return action, event
	})

	// Helper function to update panel focus indicators
	updatePanelFocus := func() {
		if detailPanelFocused {
			issueList.SetBorderColor(tcell.ColorGray)
			issueList.SetTitle("Issues")
			detailPanel.SetBorderColor(tcell.ColorYellow)
			detailPanel.SetTitle("Details [FOCUSED - Use Ctrl-d/u to scroll, ESC to return]")
			app.SetFocus(detailPanel)
		} else {
			issueList.SetBorderColor(tcell.ColorDefault)
			issueList.SetTitle("Issues")
			detailPanel.SetBorderColor(tcell.ColorGray)
			detailPanel.SetTitle("Details [Press Tab or Enter to focus]")
			app.SetFocus(issueList)
		}
		statusBar.SetText(getStatusBarText())
	}
	// Set initial focus state
	updatePanelFocus()

	// Function to show issue details
	showIssueDetails := func(issue *parser.Issue) {
		currentDetailIssue = issue
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

	// Pages for modal dialogs
	pages := tview.NewPages().
		AddPage("main", flex, true, true)

	// Set up signal handler for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Run signal handler in goroutine
	go func() {
		sig := <-sigChan
		log.Printf("SIGNAL: Received signal %v, initiating graceful shutdown", sig)

		// Stop the TUI application
		app.Stop()

		// Give deferred cleanup functions time to execute
		// If they don't complete within 5 seconds, force exit
		cleanupDone := make(chan struct{})
		go func() {
			// This will be reached after app.Stop() returns and we're back in the main goroutine
			time.Sleep(100 * time.Millisecond) // Small delay to allow main() to return
			close(cleanupDone)
		}()

		select {
		case <-cleanupDone:
			log.Printf("SIGNAL: Graceful shutdown complete")
		case <-time.After(5 * time.Second):
			log.Printf("SIGNAL: Shutdown timeout, forcing exit")
			if logFile != nil {
				logFile.Sync()
				logFile.Close()
			}
			os.Exit(1)
		}
	}()

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

	// Helper function to show comment dialog
	showCommentDialog := func() {
		// Get current issue
		currentIndex := issueList.GetCurrentItem()
		issue, ok := indexToIssue[currentIndex]
		if !ok {
			statusBar.SetText("[red]No issue selected[-]")
			return
		}

		form := tview.NewForm()
		var commentText string

		form.AddTextView("Adding comment to", issue.ID+" - "+issue.Title, 0, 2, false, false)
		form.AddTextArea("Comment", "", 60, 8, 0, func(text string) {
			commentText = text
		})

		form.AddButton("Save (Ctrl-S)", func() {
			if commentText == "" {
				statusBar.SetText("[red]Error: Comment cannot be empty[-]")
				return
			}

			// Execute bd comment command
			cmd := fmt.Sprintf("bd comment %s %q", issue.ID, commentText)
			log.Printf("BD COMMAND: Adding comment: %s", cmd)
			output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
			if err != nil {
				log.Printf("BD COMMAND ERROR: Comment failed: %v, output: %s", err, string(output))
				statusBar.SetText(fmt.Sprintf("[red]Error adding comment: %v[-]", err))
			} else {
				log.Printf("BD COMMAND: Comment added successfully: %s", string(output))
				statusBar.SetText("[green]✓ Comment added successfully[-]")

				// Close dialog
				pages.RemovePage("comment_dialog")
				app.SetFocus(issueList)

				// Refresh issues after a short delay, preserving selection
				issueID := issue.ID
				time.AfterFunc(500*time.Millisecond, func() {
					refreshIssues(issueID)
				})
			}
		})
		form.AddButton("Cancel", func() {
			pages.RemovePage("comment_dialog")
			app.SetFocus(issueList)
		})

		form.SetBorder(true).SetTitle(" Add Comment ").SetTitleAlign(tview.AlignCenter)
		form.SetCancelFunc(func() {
			pages.RemovePage("comment_dialog")
			app.SetFocus(issueList)
		})

		// Add Ctrl-S handler for save
		form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyCtrlS {
				// Save comment directly
				if commentText == "" {
					statusBar.SetText("[red]Error: Comment cannot be empty[-]")
					return nil
				}

				cmd := fmt.Sprintf("bd comment %s %q", issue.ID, commentText)
				log.Printf("BD COMMAND: Adding comment: %s", cmd)
				output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
				if err != nil {
					log.Printf("BD COMMAND ERROR: Comment failed: %v, output: %s", err, string(output))
					statusBar.SetText(fmt.Sprintf("[red]Error adding comment: %v[-]", err))
				} else {
					log.Printf("BD COMMAND: Comment added successfully: %s", string(output))
					statusBar.SetText("[green]✓ Comment added successfully[-]")
					pages.RemovePage("comment_dialog")
					app.SetFocus(issueList)
					issueID := issue.ID
					time.AfterFunc(500*time.Millisecond, func() {
						refreshIssues(issueID)
					})
				}
				return nil
			}
			return event
		})

		// Create modal (centered)
		modal := tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(form, 0, 3, true).
				AddItem(nil, 0, 1, false), 0, 3, true).
			AddItem(nil, 0, 1, false)

		pages.AddPage("comment_dialog", modal, true, true)
		app.SetFocus(form)
	}

	// Helper function to show quick filter (keyboard-friendly)
	showQuickFilter := func() {
		form := tview.NewForm()
		var filterQuery string

		helpText := `[yellow]Quick Filter Syntax:[-]
  p0-p4    Priority (e.g., 'p1' or 'p1,p2')
  bug, feature, task, epic, chore    Types
  open, in_progress, blocked, closed    Statuses

[cyan]Examples:[-]
  p1 bug          P1 bugs only
  feature,task    Features and tasks
  p0,p1 open      High priority open issues

[gray]Leave empty to clear all filters[-]`

		form.AddTextView("", helpText, 0, 11, false, false)
		form.AddInputField("Filter", "", 50, nil, func(text string) {
			filterQuery = text
		})

		// Apply filter function
		applyQuickFilter := func() {
			// Clear existing filters
			appState.ClearAllFilters()

			if filterQuery == "" {
				// Empty query = clear all filters
				pages.RemovePage("quick_filter")
				app.SetFocus(issueList)
				statusBar.SetText(getStatusBarText())
				populateIssueList()
				return
			}

			// Parse filter query (space or comma separated)
			query := strings.ToLower(strings.TrimSpace(filterQuery))
			tokens := strings.FieldsFunc(query, func(r rune) bool {
				return r == ' ' || r == ','
			})

			// Process each token
			for _, token := range tokens {
				token = strings.TrimSpace(token)
				if token == "" {
					continue
				}

				// Check for priority (p0-p4)
				if len(token) == 2 && token[0] == 'p' && token[1] >= '0' && token[1] <= '4' {
					priority := int(token[1] - '0')
					appState.TogglePriorityFilter(priority)
					continue
				}

				// Check for type
				switch token {
				case "bug":
					appState.ToggleTypeFilter(parser.TypeBug)
				case "feature":
					appState.ToggleTypeFilter(parser.TypeFeature)
				case "task":
					appState.ToggleTypeFilter(parser.TypeTask)
				case "epic":
					appState.ToggleTypeFilter(parser.TypeEpic)
				case "chore":
					appState.ToggleTypeFilter(parser.TypeChore)
				}

				// Check for status
				switch token {
				case "open":
					appState.ToggleStatusFilter(parser.StatusOpen)
				case "in_progress", "inprogress":
					appState.ToggleStatusFilter(parser.StatusInProgress)
				case "blocked":
					appState.ToggleStatusFilter(parser.StatusBlocked)
				case "closed":
					appState.ToggleStatusFilter(parser.StatusClosed)
				}
			}

			// Apply filters
			pages.RemovePage("quick_filter")
			app.SetFocus(issueList)
			statusBar.SetText(getStatusBarText())
			populateIssueList()
		}

		form.AddButton("Apply (Enter)", applyQuickFilter)
		form.AddButton("Clear All", func() {
			appState.ClearAllFilters()
			pages.RemovePage("quick_filter")
			app.SetFocus(issueList)
			statusBar.SetText(getStatusBarText())
			populateIssueList()
		})
		form.AddButton("Cancel (ESC)", func() {
			pages.RemovePage("quick_filter")
			app.SetFocus(issueList)
		})

		form.SetBorder(true).SetTitle(" Quick Filter ").SetTitleAlign(tview.AlignCenter)
		form.SetCancelFunc(func() {
			pages.RemovePage("quick_filter")
			app.SetFocus(issueList)
		})

		// Add Enter key handler to apply filter
		form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEnter {
				applyQuickFilter()
				return nil
			}
			return event
		})

		// Create modal (centered)
		modal := tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(form, 0, 2, true).
				AddItem(nil, 0, 1, false), 0, 2, true).
			AddItem(nil, 0, 1, false)

		pages.AddPage("quick_filter", modal, true, true)
		app.SetFocus(form)
	}

	// Helper function to show help screen
	showHelpScreen := func() {
		helpText := `[yellow::b]beads-tui Keyboard Shortcuts[-::-]

[cyan::b]Navigation[-::-]
  j / ↓       Move down
  k / ↑       Move up
  gg          Jump to top
  G           Jump to bottom
  Tab         Focus detail panel for scrolling
  Enter       Focus detail panel (when on issue)
  ESC         Return focus to issue list

[cyan::b]Search[-::-]
  /           Start search mode
  n           Next search result
  N           Previous search result
  ESC         Exit search mode

[cyan::b]Quick Actions[-::-]
  0-4         Set priority (P0=critical, P1=high, P2=normal, P3=low, P4=lowest)
  s           Cycle status (open → in_progress → blocked → closed → open)
  a           Create new issue (vim-style "add")
  c           Add comment to selected issue
  e           Edit issue fields (description, design, acceptance, notes) in $EDITOR
  D           Manage dependencies (add/remove blocks, parent-child, related)
  L           Manage labels (add/remove labels)
  y           Yank (copy) issue ID to clipboard
  Y           Yank (copy) issue ID with title to clipboard

[cyan::b]View Controls[-::-]
  t           Toggle between list and tree view
  f           Quick filter (type: p1 bug, feature, etc.)
  m           Toggle mouse mode on/off
  r           Manual refresh

[cyan::b]Detail Panel Scrolling (when focused)[-::-]
  Ctrl-d      Scroll down half page
  Ctrl-u      Scroll up half page
  Ctrl-e      Scroll down one line
  Ctrl-y      Scroll up one line
  PageDown    Scroll down full page
  PageUp      Scroll up full page
  Home        Jump to top of details
  End         Jump to bottom of details

[cyan::b]General[-::-]
  ?           Show this help screen
  q           Quit

[cyan::b]Status Icons[-::-]
  ●           Open/Ready
  ○           Blocked
  ◆           In Progress
  ·           Other

[cyan::b]Priority Colors[-::-]
  [red]P0[-]          Critical
  [orange]P1[-]          High
  P2          Normal (white)
  [gray]P3[-]          Low
  [darkgray]P4[-]          Lowest

[cyan::b]Status Colors[-::-]
  [green]●[-]           Ready
  [yellow]○[-]           Blocked
  [blue]◆[-]           In Progress

[gray]Press ESC or ? to close this help screen[-]`

		// Create help text view
		helpTextView := tview.NewTextView().
			SetDynamicColors(true).
			SetText(helpText).
			SetTextAlign(tview.AlignLeft)
		helpTextView.SetBorder(true).
			SetTitle(" Help - Keyboard Shortcuts ").
			SetTitleAlign(tview.AlignCenter)

		// Create modal (centered)
		modal := tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(helpTextView, 0, 3, true).
				AddItem(nil, 0, 1, false), 0, 2, true).
			AddItem(nil, 0, 1, false)

		// Add input capture to close on ESC or ?
		modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEscape || (event.Key() == tcell.KeyRune && event.Rune() == '?') {
				pages.RemovePage("help")
				app.SetFocus(issueList)
				return nil
			}
			return event
		})

		pages.AddPage("help", modal, true, true)
		app.SetFocus(modal)
	}

	// Helper function to edit a field in $EDITOR
	editFieldInEditor := func(issue *parser.Issue, fieldName string, currentValue string) {
		// Get editor from environment or default to vim
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}

		// Create temp file
		tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("beads-tui-%s-%s.md", fieldName, issue.ID))
		template := fmt.Sprintf(`# Edit %s for %s
# Lines starting with # are ignored
# Save and exit to update, or exit without saving to cancel

%s`, fieldName, issue.ID, currentValue)

		if err := os.WriteFile(tempFile, []byte(template), 0600); err != nil {
			log.Printf("EDITOR ERROR: Failed to create temp file: %v", err)
			statusBar.SetText(fmt.Sprintf("[red]Error creating temp file: %v[-]", err))
			return
		}
		defer os.Remove(tempFile)

		// Suspend TUI
		app.Suspend(func() {
			log.Printf("EDITOR: Spawning editor: %s %s", editor, tempFile)
			// Spawn editor
			cmd := exec.Command(editor, tempFile)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				log.Printf("EDITOR ERROR: Editor exited with error: %v", err)
				fmt.Fprintf(os.Stderr, "\nEditor exited with error: %v\nPress Enter to continue...", err)
				fmt.Scanln()
				return
			}

			// Read temp file
			content, err := os.ReadFile(tempFile)
			if err != nil {
				log.Printf("EDITOR ERROR: Failed to read temp file: %v", err)
				fmt.Fprintf(os.Stderr, "\nFailed to read temp file: %v\nPress Enter to continue...", err)
				fmt.Scanln()
				return
			}

			// Strip comment lines and trim
			var lines []string
			for _, line := range strings.Split(string(content), "\n") {
				if !strings.HasPrefix(strings.TrimSpace(line), "#") {
					lines = append(lines, line)
				}
			}
			strippedContent := strings.TrimSpace(strings.Join(lines, "\n"))

			// If content is empty, ask for confirmation
			if strippedContent == "" && currentValue != "" {
				fmt.Print("\nContent is empty. Clear this field? (y/n): ")
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					log.Printf("EDITOR: User cancelled clearing field")
					return
				}
			}

			// Update via bd command
			var bdFieldFlag string
			switch strings.ToLower(fieldName) {
			case "description":
				bdFieldFlag = "--description"
			case "design":
				bdFieldFlag = "--design"
			case "acceptance":
				bdFieldFlag = "--acceptance"
			case "notes":
				bdFieldFlag = "--notes"
			default:
				log.Printf("EDITOR ERROR: Unknown field name: %s", fieldName)
				fmt.Fprintf(os.Stderr, "\nUnknown field: %s\nPress Enter to continue...", fieldName)
				fmt.Scanln()
				return
			}

			// Write content to temp file for bd command (avoid shell escaping issues)
			contentFile := filepath.Join(os.TempDir(), fmt.Sprintf("beads-tui-content-%s.txt", issue.ID))
			if err := os.WriteFile(contentFile, []byte(strippedContent), 0600); err != nil {
				log.Printf("EDITOR ERROR: Failed to write content file: %v", err)
				fmt.Fprintf(os.Stderr, "\nFailed to write content file: %v\nPress Enter to continue...", err)
				fmt.Scanln()
				return
			}
			defer os.Remove(contentFile)

			bdCmd := fmt.Sprintf("bd update %s %s \"$(cat %s)\"", issue.ID, bdFieldFlag, contentFile)
			log.Printf("BD COMMAND: Updating %s: %s", fieldName, bdCmd)
			output, err := exec.Command("sh", "-c", bdCmd).CombinedOutput()
			if err != nil {
				log.Printf("BD COMMAND ERROR: Update failed: %v, output: %s", err, string(output))
				fmt.Fprintf(os.Stderr, "\nError updating %s: %v\nOutput: %s\nPress Enter to continue...", fieldName, err, string(output))
				fmt.Scanln()
			} else {
				log.Printf("BD COMMAND: Update successful for %s", fieldName)
				fmt.Fprintf(os.Stderr, "\n✓ Updated %s for %s\n", fieldName, issue.ID)
				time.Sleep(500 * time.Millisecond)
			}
		})

		// Resume TUI and refresh
		if err := app.Draw(); err != nil {
			log.Printf("APP ERROR: Failed to redraw after editor: %v", err)
		}
		statusBar.SetText(fmt.Sprintf("[green]✓ Updated %s for %s[-]", fieldName, issue.ID))
		time.AfterFunc(500*time.Millisecond, func() {
			refreshIssues(issue.ID)
		})
	}

	// Helper function to manage dependencies
	showDependencyDialog := func() {
		// Get current issue
		currentIndex := issueList.GetCurrentItem()
		issue, ok := indexToIssue[currentIndex]
		if !ok {
			statusBar.SetText("[red]No issue selected[-]")
			return
		}

		form := tview.NewForm()
		form.AddTextView("Managing dependencies for", issue.ID+" - "+issue.Title, 0, 2, false, false)

		// Show current dependencies
		if len(issue.Dependencies) > 0 {
			depText := "Current Dependencies:\n"
			for _, dep := range issue.Dependencies {
				depText += fmt.Sprintf("  %s → %s\n", dep.Type, dep.DependsOnID)
			}
			form.AddTextView("", depText, 0, len(issue.Dependencies)+1, false, false)
		} else {
			form.AddTextView("", "No dependencies", 0, 1, false, false)
		}

		// Add new dependency fields
		var targetID, depType string
		form.AddInputField("Add Dependency (Issue ID)", "", 20, nil, func(text string) {
			targetID = text
		})
		form.AddDropDown("Dependency Type", []string{"blocks", "parent-child", "related", "discovered-from"}, 0, func(option string, index int) {
			depType = option
		})

		// Add button
		form.AddButton("Add Dependency", func() {
			if targetID == "" {
				statusBar.SetText("[red]Error: Issue ID required[-]")
				return
			}

			// Validate target issue exists
			if appState.GetIssueByID(targetID) == nil {
				statusBar.SetText(fmt.Sprintf("[red]Error: Issue %s not found[-]", targetID))
				return
			}

			issueID := issue.ID // Capture before potential refresh
			cmd := fmt.Sprintf("bd dep add %s %s --type %s", issueID, targetID, depType)
			log.Printf("BD COMMAND: Adding dependency: %s", cmd)
			output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
			if err != nil {
				log.Printf("BD COMMAND ERROR: Dependency add failed: %v, output: %s", err, string(output))
				statusBar.SetText(fmt.Sprintf("[red]Error adding dependency: %v[-]", err))
			} else {
				log.Printf("BD COMMAND: Dependency added successfully")
				statusBar.SetText(fmt.Sprintf("[green]✓ Added %s dependency to %s[-]", depType, targetID))
				pages.RemovePage("dependency_dialog")
				app.SetFocus(issueList)
				time.AfterFunc(500*time.Millisecond, func() {
					refreshIssues(issueID)
				})
			}
		})

		// Remove dependency buttons
		if len(issue.Dependencies) > 0 {
			form.AddTextView("", "\nRemove Dependencies:", 0, 1, false, false)
			for _, dep := range issue.Dependencies {
				// Capture dep in closure
				depToRemove := dep
				buttonLabel := fmt.Sprintf("Remove %s → %s", depToRemove.Type, depToRemove.DependsOnID)
				form.AddButton(buttonLabel, func() {
					issueID := issue.ID
					cmd := fmt.Sprintf("bd dep remove %s %s --type %s", issueID, depToRemove.DependsOnID, depToRemove.Type)
					log.Printf("BD COMMAND: Removing dependency: %s", cmd)
					output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
					if err != nil {
						log.Printf("BD COMMAND ERROR: Dependency remove failed: %v, output: %s", err, string(output))
						statusBar.SetText(fmt.Sprintf("[red]Error removing dependency: %v[-]", err))
					} else {
						log.Printf("BD COMMAND: Dependency removed successfully")
						statusBar.SetText(fmt.Sprintf("[green]✓ Removed %s dependency to %s[-]", depToRemove.Type, depToRemove.DependsOnID))
						pages.RemovePage("dependency_dialog")
						app.SetFocus(issueList)
						time.AfterFunc(500*time.Millisecond, func() {
							refreshIssues(issueID)
						})
					}
				})
			}
		}

		// Close button
		form.AddButton("Close", func() {
			pages.RemovePage("dependency_dialog")
			app.SetFocus(issueList)
		})

		form.SetBorder(true).SetTitle(" Manage Dependencies ").SetTitleAlign(tview.AlignCenter)
		form.SetCancelFunc(func() {
			pages.RemovePage("dependency_dialog")
			app.SetFocus(issueList)
		})

		// Create modal (centered)
		modal := tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(form, 0, 3, true).
				AddItem(nil, 0, 1, false), 0, 2, true).
			AddItem(nil, 0, 1, false)

		pages.AddPage("dependency_dialog", modal, true, true)
		app.SetFocus(form)
	}

	// Helper function to manage labels
	showLabelDialog := func() {
		// Get current issue
		currentIndex := issueList.GetCurrentItem()
		issue, ok := indexToIssue[currentIndex]
		if !ok {
			statusBar.SetText("[red]No issue selected[-]")
			return
		}

		form := tview.NewForm()
		form.AddTextView("Managing labels for", issue.ID+" - "+issue.Title, 0, 2, false, false)

		// Show current labels
		if len(issue.Labels) > 0 {
			labelText := "Current Labels:\n  "
			for i, label := range issue.Labels {
				if i > 0 {
					labelText += ", "
				}
				labelText += label
			}
			form.AddTextView("", labelText, 0, 2, false, false)
		} else {
			form.AddTextView("", "No labels", 0, 1, false, false)
		}

		// Add new label field
		var newLabel string
		form.AddInputField("Add Label", "", 30, nil, func(text string) {
			newLabel = text
		})

		// Add button
		form.AddButton("Add Label", func() {
			trimmedLabel := strings.TrimSpace(newLabel)
			if trimmedLabel == "" {
				statusBar.SetText("[red]Error: Label cannot be empty[-]")
				return
			}

			// Check if label already exists
			for _, existing := range issue.Labels {
				if existing == trimmedLabel {
					statusBar.SetText(fmt.Sprintf("[red]Error: Label '%s' already exists[-]", trimmedLabel))
					return
				}
			}

			issueID := issue.ID // Capture before potential refresh
			cmd := fmt.Sprintf("bd label %s %q", issueID, trimmedLabel)
			log.Printf("BD COMMAND: Adding label: %s", cmd)
			output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
			if err != nil {
				log.Printf("BD COMMAND ERROR: Label add failed: %v, output: %s", err, string(output))
				statusBar.SetText(fmt.Sprintf("[red]Error adding label: %v[-]", err))
			} else {
				log.Printf("BD COMMAND: Label added successfully")
				statusBar.SetText(fmt.Sprintf("[green]✓ Added label '%s'[-]", trimmedLabel))
				pages.RemovePage("label_dialog")
				app.SetFocus(issueList)
				time.AfterFunc(500*time.Millisecond, func() {
					refreshIssues(issueID)
				})
			}
		})

		// Remove label buttons
		if len(issue.Labels) > 0 {
			form.AddTextView("", "\nRemove Labels:", 0, 1, false, false)
			for _, label := range issue.Labels {
				// Capture label in closure
				labelToRemove := label
				buttonLabel := fmt.Sprintf("Remove '%s'", labelToRemove)
				form.AddButton(buttonLabel, func() {
					issueID := issue.ID
					cmd := fmt.Sprintf("bd label %s --remove %q", issueID, labelToRemove)
					log.Printf("BD COMMAND: Removing label: %s", cmd)
					output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
					if err != nil {
						log.Printf("BD COMMAND ERROR: Label remove failed: %v, output: %s", err, string(output))
						statusBar.SetText(fmt.Sprintf("[red]Error removing label: %v[-]", err))
					} else {
						log.Printf("BD COMMAND: Label removed successfully")
						statusBar.SetText(fmt.Sprintf("[green]✓ Removed label '%s'[-]", labelToRemove))
						pages.RemovePage("label_dialog")
						app.SetFocus(issueList)
						time.AfterFunc(500*time.Millisecond, func() {
							refreshIssues(issueID)
						})
					}
				})
			}
		}

		// Close button
		form.AddButton("Close", func() {
			pages.RemovePage("label_dialog")
			app.SetFocus(issueList)
		})

		form.SetBorder(true).SetTitle(" Manage Labels ").SetTitleAlign(tview.AlignCenter)
		form.SetCancelFunc(func() {
			pages.RemovePage("label_dialog")
			app.SetFocus(issueList)
		})

		// Create modal (centered)
		modal := tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(form, 0, 3, true).
				AddItem(nil, 0, 1, false), 0, 2, true).
			AddItem(nil, 0, 1, false)

		pages.AddPage("label_dialog", modal, true, true)
		app.SetFocus(form)
	}

	// Helper function to show edit menu
	showEditMenu := func() {
		// Get current issue
		currentIndex := issueList.GetCurrentItem()
		issue, ok := indexToIssue[currentIndex]
		if !ok {
			statusBar.SetText("[red]No issue selected[-]")
			return
		}

		form := tview.NewForm()
		form.AddTextView("Editing", issue.ID+" - "+issue.Title, 0, 2, false, false)
		form.AddButton("Edit Description", func() {
			pages.RemovePage("edit_menu")
			app.SetFocus(issueList)
			editFieldInEditor(issue, "description", issue.Description)
		})
		form.AddButton("Edit Design", func() {
			pages.RemovePage("edit_menu")
			app.SetFocus(issueList)
			editFieldInEditor(issue, "design", issue.Design)
		})
		form.AddButton("Edit Acceptance Criteria", func() {
			pages.RemovePage("edit_menu")
			app.SetFocus(issueList)
			editFieldInEditor(issue, "acceptance", issue.AcceptanceCriteria)
		})
		form.AddButton("Edit Notes", func() {
			pages.RemovePage("edit_menu")
			app.SetFocus(issueList)
			editFieldInEditor(issue, "notes", issue.Notes)
		})
		form.AddButton("Cancel", func() {
			pages.RemovePage("edit_menu")
			app.SetFocus(issueList)
		})

		form.SetBorder(true).SetTitle(" Edit Issue Fields ").SetTitleAlign(tview.AlignCenter)
		form.SetCancelFunc(func() {
			pages.RemovePage("edit_menu")
			app.SetFocus(issueList)
		})

		// Create modal (centered)
		modal := tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(form, 0, 2, true).
				AddItem(nil, 0, 1, false), 0, 2, true).
			AddItem(nil, 0, 1, false)

		pages.AddPage("edit_menu", modal, true, true)
		app.SetFocus(form)
	}

	// Helper function to show issue creation dialog
	showCreateIssueDialog := func() {
		// Create form
		form := tview.NewForm()

		var title, description, priority, issueType string
		priority = "2" // Default to P2
		issueType = "feature" // Default to feature

		// Get current issue for potential parent
		var currentIssueID string
		if issue, ok := indexToIssue[issueList.GetCurrentItem()]; ok {
			currentIssueID = issue.ID
		}

		// Add form fields
		form.AddInputField("Title", "", 50, nil, func(text string) {
			title = text
		})
		form.AddTextArea("Description", "", 60, 5, 0, func(text string) {
			description = text
		})
		form.AddDropDown("Priority", []string{"P0 (Critical)", "P1 (High)", "P2 (Normal)", "P3 (Low)", "P4 (Lowest)"}, 2, func(option string, index int) {
			priority = fmt.Sprintf("%d", index)
		})
		form.AddDropDown("Type", []string{"bug", "feature", "task", "epic", "chore"}, 1, func(option string, index int) {
			issueType = option
		})
		if currentIssueID != "" {
			form.AddCheckbox("Add as child of "+currentIssueID, false, nil)
		}

		// Add buttons
		form.AddButton("Create", func() {
			if title == "" {
				statusBar.SetText("[red]Error: Title is required[-]")
				return
			}

			// Build bd create command
			cmd := fmt.Sprintf("bd create %q -p %s -t %s", title, priority, issueType)
			if description != "" {
				cmd += fmt.Sprintf(" --description %q", description)
			}

			// Check if we should add parent relationship
			if currentIssueID != "" {
				// Check checkbox state
				formItem := form.GetFormItemByLabel("Add as child of " + currentIssueID)
				if checkbox, ok := formItem.(*tview.Checkbox); ok && checkbox.IsChecked() {
					cmd += fmt.Sprintf(" --parent %s", currentIssueID)
				}
			}

			log.Printf("BD COMMAND: Creating issue: %s", cmd)
			output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
			if err != nil {
				log.Printf("BD COMMAND ERROR: Issue creation failed: %v, output: %s", err, string(output))
				statusBar.SetText(fmt.Sprintf("[red]Error creating issue: %v[-]", err))
			} else {
				log.Printf("BD COMMAND: Issue created successfully: %s", string(output))
				statusBar.SetText("[green]✓ Issue created successfully[-]")

				// Close dialog
				pages.RemovePage("create_issue")
				app.SetFocus(issueList)

				// Refresh issues after a short delay
				time.AfterFunc(500*time.Millisecond, func() {
					refreshIssues()
				})
			}
		})
		form.AddButton("Cancel", func() {
			pages.RemovePage("create_issue")
			app.SetFocus(issueList)
		})

		form.SetBorder(true).SetTitle(" Create New Issue ").SetTitleAlign(tview.AlignCenter)
		form.SetCancelFunc(func() {
			pages.RemovePage("create_issue")
			app.SetFocus(issueList)
		})

		// Create modal (centered)
		modal := tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(form, 0, 3, true).
				AddItem(nil, 0, 1, false), 0, 3, true).
			AddItem(nil, 0, 1, false)

		pages.AddPage("create_issue", modal, true, true)
		app.SetFocus(form)
	}

	// Set up key bindings
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Log all keyboard events in debug mode
		log.Printf("KEY EVENT: key=%v rune=%q mod=%v searchMode=%v detailFocus=%v",
			event.Key(), event.Rune(), event.Modifiers(), searchMode, detailPanelFocused)

		// If a modal is showing (not on main page), let the modal handle all input
		currentPage, _ := pages.GetFrontPage()
		if currentPage != "main" {
			return event
		}

		// Handle search mode
		if searchMode {
			switch event.Key() {
			case tcell.KeyEscape:
				searchMode = false
				searchQuery = ""
				statusBar.SetText(getStatusBarText())
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

		// Handle detail panel scrolling when focused
		if detailPanelFocused {
			switch event.Key() {
			case tcell.KeyEscape:
				// Return focus to issue list
				detailPanelFocused = false
				updatePanelFocus()
				return nil
			case tcell.KeyCtrlD:
				// Scroll down half page
				_, _, _, height := detailPanel.GetInnerRect()
				for i := 0; i < height/2; i++ {
					detailPanel.InputHandler()(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone), nil)
				}
				return nil
			case tcell.KeyCtrlU:
				// Scroll up half page
				_, _, _, height := detailPanel.GetInnerRect()
				for i := 0; i < height/2; i++ {
					detailPanel.InputHandler()(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone), nil)
				}
				return nil
			case tcell.KeyCtrlE:
				// Scroll down one line
				detailPanel.InputHandler()(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone), nil)
				return nil
			case tcell.KeyCtrlY:
				// Scroll up one line
				detailPanel.InputHandler()(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone), nil)
				return nil
			case tcell.KeyPgDn:
				// Page down
				detailPanel.InputHandler()(tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone), nil)
				return nil
			case tcell.KeyPgUp:
				// Page up
				detailPanel.InputHandler()(tcell.NewEventKey(tcell.KeyPgUp, 0, tcell.ModNone), nil)
				return nil
			case tcell.KeyHome:
				// Jump to top
				detailPanel.ScrollToBeginning()
				return nil
			case tcell.KeyEnd:
				// Jump to end
				detailPanel.ScrollToEnd()
				return nil
			}
			// Allow other keys to pass through
			return event
		}

		// Normal mode key bindings (issue list focused)
		switch event.Key() {
		case tcell.KeyTab:
			// Focus detail panel
			detailPanelFocused = true
			updatePanelFocus()
			return nil
		case tcell.KeyEnter:
			// If on an issue, focus detail panel (alternative to Tab)
			if _, ok := indexToIssue[issueList.GetCurrentItem()]; ok {
				detailPanelFocused = true
				updatePanelFocus()
				return nil
			}
			return event
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil
			case 'r':
				// Manual refresh - run in goroutine to avoid blocking UI
				statusBar.SetText("[yellow]Refreshing...[-]")
				go refreshIssues()
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
				statusBar.SetText(getStatusBarText())
				populateIssueList()
				return nil
			case 'm':
				// Toggle mouse mode
				mouseEnabled = !mouseEnabled
				app.EnableMouse(mouseEnabled)
				statusBar.SetText(getStatusBarText())
				return nil
			case 'a':
				// Open issue creation dialog
				showCreateIssueDialog()
				return nil
			case 'c':
				// Open comment dialog
				showCommentDialog()
				return nil
			case 'e':
				// Open edit menu for current issue
				showEditMenu()
				return nil
			case 'D':
				// Open dependency management dialog
				showDependencyDialog()
				return nil
			case 'L':
				// Open label management dialog
				showLabelDialog()
				return nil
			case 'y':
				// Yank (copy) issue ID to clipboard
				if issue, ok := indexToIssue[issueList.GetCurrentItem()]; ok {
					err := clipboard.WriteAll(issue.ID)
					if err != nil {
						log.Printf("CLIPBOARD ERROR: Failed to copy to clipboard: %v", err)
						statusBar.SetText(fmt.Sprintf("[red]Failed to copy: %v[-]", err))
					} else {
						log.Printf("CLIPBOARD: Copied issue ID to clipboard: %s", issue.ID)
						statusBar.SetText(fmt.Sprintf("[green]✓ Copied %s to clipboard[-]", issue.ID))
						// Clear message after 2 seconds
						time.AfterFunc(2*time.Second, func() {
							app.QueueUpdateDraw(func() {
								statusBar.SetText(getStatusBarText())
							})
						})
					}
				}
				return nil
			case 'Y':
				// Yank (copy) issue ID with title to clipboard
				if issue, ok := indexToIssue[issueList.GetCurrentItem()]; ok {
					text := fmt.Sprintf("%s - %s", issue.ID, issue.Title)
					err := clipboard.WriteAll(text)
					if err != nil {
						log.Printf("CLIPBOARD ERROR: Failed to copy to clipboard: %v", err)
						statusBar.SetText(fmt.Sprintf("[red]Failed to copy: %v[-]", err))
					} else {
						log.Printf("CLIPBOARD: Copied issue ID with title to clipboard: %s", text)
						statusBar.SetText(fmt.Sprintf("[green]✓ Copied '%s' to clipboard[-]", text))
						// Clear message after 2 seconds
						time.AfterFunc(2*time.Second, func() {
							app.QueueUpdateDraw(func() {
								statusBar.SetText(getStatusBarText())
							})
						})
					}
				}
				return nil
			case '?':
				// Show help screen
				showHelpScreen()
				return nil
			case 'f':
				// Show quick filter
				showQuickFilter()
				return nil
			case '0', '1', '2', '3', '4':
				// Quick priority change
				if issue, ok := indexToIssue[issueList.GetCurrentItem()]; ok {
					priority := int(event.Rune() - '0')
					issueID := issue.ID // Capture issue ID before refresh
					// Update priority via bd command
					cmd := fmt.Sprintf("bd update %s --priority %d", issueID, priority)
					log.Printf("BD COMMAND: Executing priority update: %s", cmd)
					err := exec.Command("sh", "-c", cmd).Run()
					if err != nil {
						log.Printf("BD COMMAND ERROR: Priority update failed: %v", err)
						statusBar.SetText(fmt.Sprintf("[red]Error updating priority: %v[-]", err))
					} else {
						log.Printf("BD COMMAND: Priority update successful for %s -> P%d", issueID, priority)
						statusBar.SetText(fmt.Sprintf("[green]✓ Set %s to P%d[-]", issueID, priority))
						// Refresh issues after a short delay, preserving selection
						log.Printf("BD COMMAND: Scheduling refresh in 500ms")
						time.AfterFunc(500*time.Millisecond, func() {
							log.Printf("BD COMMAND: Delayed refresh starting")
							refreshIssues(issueID)
						})
					}
				}
				return nil
			case 's':
				// Toggle status
				if issue, ok := indexToIssue[issueList.GetCurrentItem()]; ok {
					issueID := issue.ID // Capture issue ID before refresh
					// Cycle through statuses: open -> in_progress -> blocked -> closed -> open
					var newStatus string
					switch issue.Status {
					case parser.StatusOpen:
						newStatus = "in_progress"
					case parser.StatusInProgress:
						newStatus = "blocked"
					case parser.StatusBlocked:
						newStatus = "closed"
					case parser.StatusClosed:
						newStatus = "open"
					default:
						newStatus = "in_progress"
					}
					// Update status via bd command
					cmd := fmt.Sprintf("bd update %s --status %s", issueID, newStatus)
					log.Printf("BD COMMAND: Executing status update: %s", cmd)
					err := exec.Command("sh", "-c", cmd).Run()
					if err != nil {
						log.Printf("BD COMMAND ERROR: Status update failed: %v", err)
						statusBar.SetText(fmt.Sprintf("[red]Error updating status: %v[-]", err))
					} else {
						log.Printf("BD COMMAND: Status update successful for %s -> %s", issueID, newStatus)
						statusBar.SetText(fmt.Sprintf("[green]✓ Set %s to %s[-]", issueID, newStatus))
						// Refresh issues after a short delay, preserving selection
						log.Printf("BD COMMAND: Scheduling refresh in 500ms")
						time.AfterFunc(500*time.Millisecond, func() {
							log.Printf("BD COMMAND: Delayed refresh starting")
							refreshIssues(issueID)
						})
					}
				}
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
				statusBar.SetText(getStatusBarText())
				return nil
			}
		default:
			lastKeyWasG = false
		}
		return event
	})

	// Run application
	// Enable mouse by default (can be toggled with 'm' key)
	app.EnableMouse(mouseEnabled)
	log.Printf("APP: Starting tview application main loop")
	if err := app.SetRoot(pages, true).Run(); err != nil {
		log.Printf("APP ERROR: Application crashed: %v", err)
		panic(err)
	}
	log.Printf("APP: Application exited normally")
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
	result += fmt.Sprintf("[gray]ID:[-] %s [blue](click to copy)[-]  ", issue.ID)
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
