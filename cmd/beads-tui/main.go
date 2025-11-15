package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/andy/beads-tui/internal/app"
	"github.com/andy/beads-tui/internal/formatting"
	"github.com/andy/beads-tui/internal/parser"
	"github.com/andy/beads-tui/internal/state"
	"github.com/andy/beads-tui/internal/storage"
	"github.com/andy/beads-tui/internal/theme"
	_ "github.com/andy/beads-tui/internal/theme" // Import to register themes
	"github.com/andy/beads-tui/internal/ui"
	"github.com/andy/beads-tui/internal/watcher"
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	// Parse command line flags
	debugMode := flag.Bool("debug", false, "Enable debug logging to file")
	themeName := flag.String("theme", "", "Color theme (default, gruvbox-dark, etc)")
	flag.Parse()

	// Set theme if specified
	if *themeName != "" {
		if err := theme.SetCurrent(*themeName); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v, using default theme\n", err)
		}
	}

	// Check environment variable for theme
	if envTheme := os.Getenv("BEADS_THEME"); envTheme != "" && *themeName == "" {
		if err := theme.SetCurrent(envTheme); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v, using default theme\n", err)
		}
	}

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
	beadsDir, err := app.FindBeadsDir()
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

	// Two-character shortcut state
	var lastKeyWasS bool // For status shortcuts (So, Si, Sb, Sc)

	// Mouse mode state (default: enabled)
	var mouseEnabled = true

	// Panel focus state (true = detail panel, false = issue list)
	var detailPanelFocused bool

	// Show closed issues in list view (default: false)
	var showClosedIssues bool

	// Layout orientation: true = vertical, false = horizontal (default)
	var verticalLayout bool

	// Detail pane visibility (default: true)
	var detailPaneVisible = true

	// Track currently displayed issue in detail panel (for clipboard copy)
	var currentDetailIssue *parser.Issue

	// Helper functions for themed messages
	successMsg := func(msg string) string {
		return fmt.Sprintf("[%s]%s[-]", formatting.GetSuccessColor(), msg)
	}
	errorMsg := func(msg string) string {
		return fmt.Sprintf("[%s]%s[-]", formatting.GetErrorColor(), msg)
	}
	_ = func(msg string) string { // emphasisMsg - reserved for future use
		return fmt.Sprintf("[%s]%s[-]", formatting.GetEmphasisColor(), msg)
	}

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
		visibleCount := len(appState.GetReadyIssues()) + len(appState.GetBlockedIssues()) + len(appState.GetInProgressIssues())
		if showClosedIssues {
			visibleCount += len(appState.GetClosedIssues())
		}

		filterText := ""
		if appState.HasActiveFilters() {
			filterText = fmt.Sprintf(" [Filters: %s]", appState.GetActiveFilters())
		}

		closedText := ""
		if showClosedIssues {
			closedText = " [Showing Closed]"
		}

		layoutStr := "Horizontal"
		if verticalLayout {
			layoutStr = "Vertical"
		}

		return fmt.Sprintf("[yellow]Beads TUI[-] - %s (%d issues)%s%s [SQLite] [%s View] [%s] [Mouse: %s] [Focus: %s] [Press ? for help, v to toggle layout]",
			beadsDir, visibleCount, filterText, closedText, viewModeStr, layoutStr, mouseStr, focusStr)
	}

	// Helper function to populate issue list from state
	populateIssueList := func() {
		indexToIssue = ui.PopulateIssueList(issueList, appState, showClosedIssues)
	}

	// Function to load and display issues (for async updates after app starts)
	// preserveIssueID: if provided, attempt to restore selection to this issue after refresh
	refreshIssues := func(preserveIssueID ...string) {
		log.Printf("REFRESH: Starting issue refresh")
		var targetIssueID string
		if len(preserveIssueID) > 0 {
			targetIssueID = preserveIssueID[0]
			log.Printf("REFRESH: Will attempt to preserve selection on issue: %s", targetIssueID)
		} else {
			// No explicit issue ID provided, try to preserve current selection
			currentIndex := issueList.GetCurrentItem()
			if currentIssue, ok := indexToIssue[currentIndex]; ok {
				targetIssueID = currentIssue.ID
				log.Printf("REFRESH: Auto-preserving current selection: %s", targetIssueID)
			}
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
				statusBar.SetText(errorMsg(fmt.Sprintf("Error loading issues: %v", err)))
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
					statusBar.SetText(fmt.Sprintf("[limegreen]✓ Copied[-] [white]%s[-] [limegreen]to clipboard[-]", currentDetailIssue.ID))
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
		details := formatting.FormatIssueDetails(issue)
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

	// Layout builder function
	buildLayout := func() *tview.Flex {
		var contentFlex *tview.Flex

		if !detailPaneVisible {
			// Detail pane hidden: show only issue list
			contentFlex = tview.NewFlex().
				AddItem(issueList, 0, 1, true)
		} else if verticalLayout {
			// Vertical: list on top (40%), details on bottom (60%)
			contentFlex = tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(issueList, 0, 40, !detailPanelFocused).
				AddItem(detailPanel, 0, 60, detailPanelFocused)
		} else {
			// Horizontal: list on left (1 part), details on right (2 parts)
			contentFlex = tview.NewFlex().
				AddItem(issueList, 0, 1, !detailPanelFocused).
				AddItem(detailPanel, 0, 2, detailPanelFocused)
		}

		return tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(statusBar, 1, 0, false).
			AddItem(contentFlex, 0, 1, true)
	}

	flex := buildLayout()

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
			if len(mainText) > 0 && formatting.ContainsCaseInsensitive(mainText, query) {
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
	// Create dialog helpers for all dialog functions
	dialogHelpers := &DialogHelpers{
		App:           app,
		Pages:         pages,
		IssueList:     issueList,
		IndexToIssue:  &indexToIssue,
		StatusBar:     statusBar,
		AppState:      appState,
		RefreshIssues: refreshIssues,
	}

	// Helper function to show comment dialog
	showCommentDialog := func() {
		dialogHelpers.ShowCommentDialog()
	}

	// Helper function to show rename dialog
	showRenameDialog := func() {
		dialogHelpers.ShowRenameDialog()
	}

	// Helper function to show quick filter (keyboard-friendly)
	showQuickFilter := func() {
		dialogHelpers.ShowQuickFilter()
		statusBar.SetText(getStatusBarText())
		populateIssueList()
	}

	// Helper function to show stats dashboard
	showStatsOverlay := func() {
		dialogHelpers.ShowStatsOverlay()
	}

	// Helper function to show help screen
	showHelpScreen := func() {
		dialogHelpers.ShowHelpScreen()
	}

	// Helper function to manage dependencies
	showDependencyDialog := func() {
		dialogHelpers.ShowDependencyDialog()
	}

	// Helper function to manage labels
	showLabelDialog := func() {
		dialogHelpers.ShowLabelDialog()
	}

	// Helper function to close issue with optional reason
	showCloseIssueDialog := func() {
		dialogHelpers.ShowCloseIssueDialog()
	}

	// Helper function to reopen closed issue with optional reason
	showReopenIssueDialog := func() {
		dialogHelpers.ShowReopenIssueDialog()
	}

	// Helper function to show edit form (in-TUI editing, similar to create issue form)
	showEditForm := func() {
		dialogHelpers.ShowEditForm()
	}

	// Helper function to show issue creation dialog
	showCreateIssueDialog := func() {
		dialogHelpers.ShowCreateIssueDialog()
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
				// Hide detail pane and return focus to issue list
				detailPaneVisible = false
				detailPanelFocused = false
				newFlex := buildLayout()
				pages.RemovePage("main")
				pages.AddPage("main", newFlex, true, true)
				statusBar.SetText(getStatusBarText())
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
			case tcell.KeyCtrlF:
				// Scroll down full page (vim style)
				detailPanel.InputHandler()(tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone), nil)
				return nil
			case tcell.KeyCtrlB:
				// Scroll up full page (vim style)
				detailPanel.InputHandler()(tcell.NewEventKey(tcell.KeyPgUp, 0, tcell.ModNone), nil)
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
			// If on an issue, show detail pane and focus it
			if _, ok := indexToIssue[issueList.GetCurrentItem()]; ok {
				if !detailPaneVisible {
					// Show detail pane
					detailPaneVisible = true
					newFlex := buildLayout()
					pages.RemovePage("main")
					pages.AddPage("main", newFlex, true, true)
				}
				detailPanelFocused = true
				updatePanelFocus()
				statusBar.SetText(getStatusBarText())
				return nil
			}
			return event
		case tcell.KeyCtrlB:
			// Scroll up full page (vim style)
			_, _, _, height := issueList.GetInnerRect()
			currentItem := issueList.GetCurrentItem()
			newItem := currentItem - height
			if newItem < 0 {
				newItem = 0
			}
			issueList.SetCurrentItem(newItem)
			return nil
		case tcell.KeyCtrlF:
			// Scroll down full page (vim style)
			_, _, _, height := issueList.GetInnerRect()
			currentItem := issueList.GetCurrentItem()
			maxItem := issueList.GetItemCount() - 1
			newItem := currentItem + height
			if newItem > maxItem {
				newItem = maxItem
			}
			issueList.SetCurrentItem(newItem)
			return nil
		case tcell.KeyRune:
			// Handle multi-key sequences FIRST before processing individual keys
			// This prevents conflicts with single-key handlers

			// Handle status shortcuts (S + second char)
			if lastKeyWasS {
				var newStatus string
				switch event.Rune() {
				case 'o':
					newStatus = "open"
				case 'i':
					newStatus = "in_progress"
				case 'b':
					newStatus = "blocked"
				case 'c':
					newStatus = "closed"
				default:
					// Invalid second key, reset and fall through
					lastKeyWasS = false
					statusBar.SetText(getStatusBarText())
					return nil
				}

				// Execute status update
				if issue, ok := indexToIssue[issueList.GetCurrentItem()]; ok {
					issueID := issue.ID
					log.Printf("BD COMMAND: Executing status update (S%c): bd update %s --status %s", event.Rune(), issueID, newStatus)
					updatedIssue, err := execBdJSONIssue("update", issueID, "--status", string(newStatus))
					if err != nil {
						statusBar.SetText(errorMsg(fmt.Sprintf("Error updating status: %v", err)))
					} else {
						statusBar.SetText(successMsg(fmt.Sprintf("✓ Set %s to %s", updatedIssue.ID, updatedIssue.Status)))
						time.AfterFunc(500*time.Millisecond, func() {
							refreshIssues(issueID)
						})
					}
				}
				lastKeyWasS = false
				return nil
			}

			// Normal single-key handling
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
			case 'v':
				// Toggle layout orientation (horizontal/vertical)
				verticalLayout = !verticalLayout
				newFlex := buildLayout()
				pages.RemovePage("main")
				pages.AddPage("main", newFlex, true, true)
				app.SetRoot(pages, true)
				statusBar.SetText(getStatusBarText())
				return nil
			case 'C':
				// Toggle showing closed issues
				showClosedIssues = !showClosedIssues
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
			case 'e':
				// Edit issue fields
				showEditForm()
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
						statusBar.SetText(successMsg(fmt.Sprintf("✓ Copied %s to clipboard", issue.ID)))
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
						statusBar.SetText(successMsg(fmt.Sprintf("✓ Copied '%s' to clipboard", text)))
						// Clear message after 2 seconds
						time.AfterFunc(2*time.Second, func() {
							app.QueueUpdateDraw(func() {
								statusBar.SetText(getStatusBarText())
							})
						})
					}
				}
				return nil
			case 'B':
				// Copy git branch name to clipboard
				if issue, ok := indexToIssue[issueList.GetCurrentItem()]; ok {
					branchName := issue.ID // Simple format: just the issue ID
					err := clipboard.WriteAll(branchName)
					if err != nil {
						log.Printf("CLIPBOARD ERROR: Failed to copy branch name: %v", err)
						statusBar.SetText(fmt.Sprintf("[red]Failed to copy: %v[-]", err))
					} else {
						log.Printf("CLIPBOARD: Copied branch name to clipboard: %s", branchName)
						statusBar.SetText(successMsg(fmt.Sprintf("✓ Copied branch name '%s' to clipboard", branchName)))
						// Clear message after 2 seconds
						time.AfterFunc(2*time.Second, func() {
							app.QueueUpdateDraw(func() {
								statusBar.SetText(getStatusBarText())
							})
						})
					}
				}
				return nil
			case 'R':
				// Rename issue (edit title)
				showRenameDialog()
				return nil
			case 'x':
				// Close issue with optional reason
				showCloseIssueDialog()
				return nil
			case 'X':
				// Reopen closed issue with optional reason
				showReopenIssueDialog()
				return nil
			case '?':
				// Show help screen
				showHelpScreen()
				return nil
			case 'f':
				// Show quick filter
				showQuickFilter()
				return nil
			case 'S':
				// Show stats dashboard
				showStatsOverlay()
				return nil
			case '0', '1', '2', '3', '4':
				// Quick priority change
				if issue, ok := indexToIssue[issueList.GetCurrentItem()]; ok {
					priority := int(event.Rune() - '0')
					issueID := issue.ID // Capture issue ID before refresh
					// Update priority via bd command with --json
					log.Printf("BD COMMAND: Executing priority update: bd update %s --priority %d", issueID, priority)
					updatedIssue, err := execBdJSONIssue("update", issueID, "--priority", fmt.Sprintf("%d", priority))
					if err != nil {
						log.Printf("BD COMMAND ERROR: Priority update failed: %v", err)
						statusBar.SetText(errorMsg(fmt.Sprintf("Error updating priority: %v", err)))
					} else {
						log.Printf("BD COMMAND: Priority update successful for %s -> P%d", updatedIssue.ID, updatedIssue.Priority)
						statusBar.SetText(successMsg(fmt.Sprintf("✓ Set %s to P%d", updatedIssue.ID, updatedIssue.Priority)))
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
				// Initiate status shortcut sequence
				lastKeyWasS = true
				statusBar.SetText("[yellow]Status shortcut: o/i/b/c[-]")
				// Reset after 2 seconds if no second key
				time.AfterFunc(2*time.Second, func() {
					app.QueueUpdateDraw(func() {
						if lastKeyWasS {
							lastKeyWasS = false
							statusBar.SetText(getStatusBarText())
						}
					})
				})
				return nil
			case 'c':
				// Add comment to issue
				showCommentDialog()
				return nil
			default:
				// Reset all multi-key flags if any other key is pressed
				lastKeyWasG = false
				lastKeyWasS = false
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

// Helper functions have been moved to internal packages:
// - formatting: color, status, details formatting
// - app: initialization and context management
// - ui: component creation and rendering
