package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

		return fmt.Sprintf("[yellow]Beads TUI[-] - %s (%d issues)%s [SQLite] [%s View] [Mouse: %s] [Focus: %s] [Press ? for help, f for filters]",
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
	refreshIssues := func() {
		log.Printf("REFRESH: Starting issue refresh")
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

				// Refresh issues after a short delay
				time.AfterFunc(500*time.Millisecond, func() {
					refreshIssues()
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
					time.AfterFunc(500*time.Millisecond, func() {
						refreshIssues()
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

	// Helper function to show filter menu
	showFilterMenu := func() {
		form := tview.NewForm()

		// Priority checkboxes
		form.AddTextView("Priority Filter", "Toggle priorities to show:", 0, 1, false, false)
		for p := 0; p <= 4; p++ {
			priority := p // Capture for closure
			checked := appState.IsPriorityFiltered(priority)
			form.AddCheckbox(fmt.Sprintf("P%d", priority), checked, func(isChecked bool) {
				appState.TogglePriorityFilter(priority)
			})
		}

		// Type checkboxes
		form.AddTextView("Type Filter", "Toggle types to show:", 0, 1, false, false)
		types := []parser.IssueType{parser.TypeBug, parser.TypeFeature, parser.TypeTask, parser.TypeEpic, parser.TypeChore}
		for _, t := range types {
			issueType := t // Capture for closure
			checked := appState.IsTypeFiltered(issueType)
			form.AddCheckbox(string(issueType), checked, func(isChecked bool) {
				appState.ToggleTypeFilter(issueType)
			})
		}

		// Status checkboxes
		form.AddTextView("Status Filter", "Toggle statuses to show:", 0, 1, false, false)
		statuses := []parser.Status{parser.StatusOpen, parser.StatusInProgress, parser.StatusBlocked, parser.StatusClosed}
		for _, s := range statuses {
			status := s // Capture for closure
			checked := appState.IsStatusFiltered(status)
			form.AddCheckbox(string(status), checked, func(isChecked bool) {
				appState.ToggleStatusFilter(status)
			})
		}

		// Buttons
		form.AddButton("Apply", func() {
			pages.RemovePage("filter_menu")
			app.SetFocus(issueList)
			statusBar.SetText(getStatusBarText())
			populateIssueList()
		})
		form.AddButton("Clear All", func() {
			appState.ClearAllFilters()
			pages.RemovePage("filter_menu")
			app.SetFocus(issueList)
			statusBar.SetText(getStatusBarText())
			populateIssueList()
		})
		form.AddButton("Cancel", func() {
			pages.RemovePage("filter_menu")
			app.SetFocus(issueList)
		})

		form.SetBorder(true).SetTitle(" Filter Issues ").SetTitleAlign(tview.AlignCenter)
		form.SetCancelFunc(func() {
			pages.RemovePage("filter_menu")
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

		pages.AddPage("filter_menu", modal, true, true)
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

[cyan::b]View Controls[-::-]
  t           Toggle between list and tree view
  f           Open filter menu (priority, type, status)
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
			case '?':
				// Show help screen
				showHelpScreen()
				return nil
			case 'f':
				// Show filter menu
				showFilterMenu()
				return nil
			case '0', '1', '2', '3', '4':
				// Quick priority change
				if issue, ok := indexToIssue[issueList.GetCurrentItem()]; ok {
					priority := int(event.Rune() - '0')
					// Update priority via bd command
					cmd := fmt.Sprintf("bd update %s --priority %d", issue.ID, priority)
					log.Printf("BD COMMAND: Executing priority update: %s", cmd)
					err := exec.Command("sh", "-c", cmd).Run()
					if err != nil {
						log.Printf("BD COMMAND ERROR: Priority update failed: %v", err)
						statusBar.SetText(fmt.Sprintf("[red]Error updating priority: %v[-]", err))
					} else {
						log.Printf("BD COMMAND: Priority update successful for %s -> P%d", issue.ID, priority)
						statusBar.SetText(fmt.Sprintf("[green]✓ Set %s to P%d[-]", issue.ID, priority))
						// Refresh issues after a short delay
						log.Printf("BD COMMAND: Scheduling refresh in 500ms")
						time.AfterFunc(500*time.Millisecond, func() {
							log.Printf("BD COMMAND: Delayed refresh starting")
							refreshIssues()
						})
					}
				}
				return nil
			case 's':
				// Toggle status
				if issue, ok := indexToIssue[issueList.GetCurrentItem()]; ok {
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
					cmd := fmt.Sprintf("bd update %s --status %s", issue.ID, newStatus)
					log.Printf("BD COMMAND: Executing status update: %s", cmd)
					err := exec.Command("sh", "-c", cmd).Run()
					if err != nil {
						log.Printf("BD COMMAND ERROR: Status update failed: %v", err)
						statusBar.SetText(fmt.Sprintf("[red]Error updating status: %v[-]", err))
					} else {
						log.Printf("BD COMMAND: Status update successful for %s -> %s", issue.ID, newStatus)
						statusBar.SetText(fmt.Sprintf("[green]✓ Set %s to %s[-]", issue.ID, newStatus))
						// Refresh issues after a short delay
						log.Printf("BD COMMAND: Scheduling refresh in 500ms")
						time.AfterFunc(500*time.Millisecond, func() {
							log.Printf("BD COMMAND: Delayed refresh starting")
							refreshIssues()
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
