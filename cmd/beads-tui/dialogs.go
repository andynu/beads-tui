package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/andy/beads-tui/internal/formatting"
	"github.com/andy/beads-tui/internal/parser"
	"github.com/andy/beads-tui/internal/state"
	"github.com/andy/beads-tui/internal/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// DialogHelpers holds references to UI components needed by dialog functions
type DialogHelpers struct {
	App           *tview.Application
	Pages         *tview.Pages
	IssueList     *tview.List
	IndexToIssue  *map[int]*parser.Issue
	StatusBar     *tview.TextView
	AppState      *state.State
	RefreshIssues func(...string)
}

// ShowCommentDialog displays a dialog to add a comment to the current issue
func (h *DialogHelpers) ShowCommentDialog() {
	// Get current issue
	currentIndex := h.IssueList.GetCurrentItem()
	issue, ok := (*h.IndexToIssue)[currentIndex]
	if !ok {
		h.StatusBar.SetText(fmt.Sprintf("[%s]No issue selected[-]", formatting.GetErrorColor()))
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
			h.StatusBar.SetText(fmt.Sprintf("[%s]Error: Comment cannot be empty[-]", formatting.GetErrorColor()))
			return
		}

		// Execute bd comment command with --json
		log.Printf("BD COMMAND: Adding comment: bd comment %s %q", issue.ID, commentText)
		comment, err := execBdJSONComment("comment", issue.ID, commentText)
		if err != nil {
			log.Printf("BD COMMAND ERROR: Comment failed: %v", err)
			h.StatusBar.SetText(fmt.Sprintf("[%s]Error adding comment: %v[-]", formatting.GetErrorColor(), err))
		} else {
			log.Printf("BD COMMAND: Comment added successfully: ID %d", comment.ID)
			h.StatusBar.SetText(fmt.Sprintf("[%s]✓ Comment added successfully[-]", formatting.GetSuccessColor()))

			// Close dialog
			h.Pages.RemovePage("comment_dialog")
			h.App.SetFocus(h.IssueList)

			// Refresh issues after a short delay, preserving selection
			issueID := issue.ID
			time.AfterFunc(500*time.Millisecond, func() {
				h.RefreshIssues(issueID)
			})
		}
	})
	form.AddButton("Cancel", func() {
		h.Pages.RemovePage("comment_dialog")
		h.App.SetFocus(h.IssueList)
	})

	form.SetBorder(true).SetTitle(" Add Comment ").SetTitleAlign(tview.AlignCenter)
	form.SetCancelFunc(func() {
		h.Pages.RemovePage("comment_dialog")
		h.App.SetFocus(h.IssueList)
	})

	// Add Ctrl-S handler for save
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlS {
			// Save comment directly
			if commentText == "" {
				h.StatusBar.SetText(fmt.Sprintf("[%s]Error: Comment cannot be empty[-]", formatting.GetErrorColor()))
				return nil
			}

			log.Printf("BD COMMAND: Adding comment: bd comment %s %q", issue.ID, commentText)
			comment, err := execBdJSONComment("comment", issue.ID, commentText)
			if err != nil {
				log.Printf("BD COMMAND ERROR: Comment failed: %v", err)
				h.StatusBar.SetText(fmt.Sprintf("[%s]Error adding comment: %v[-]", formatting.GetErrorColor(), err))
			} else {
				log.Printf("BD COMMAND: Comment added successfully: ID %d", comment.ID)
				h.StatusBar.SetText(fmt.Sprintf("[%s]✓ Comment added successfully[-]", formatting.GetSuccessColor()))
				h.Pages.RemovePage("comment_dialog")
				h.App.SetFocus(h.IssueList)
				issueID := issue.ID
				time.AfterFunc(500*time.Millisecond, func() {
					h.RefreshIssues(issueID)
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

	h.Pages.AddPage("comment_dialog", modal, true, true)
	h.App.SetFocus(form)
}

// ShowRenameDialog displays a dialog to rename the current issue
func (h *DialogHelpers) ShowRenameDialog() {
	// Get current issue
	currentIndex := h.IssueList.GetCurrentItem()
	issue, ok := (*h.IndexToIssue)[currentIndex]
	if !ok {
		h.StatusBar.SetText(fmt.Sprintf("[%s]No issue selected[-]", formatting.GetErrorColor()))
		return
	}

	form := tview.NewForm()
	var newTitle string

	form.AddTextView("Renaming issue", issue.ID, 0, 1, false, false)
	form.AddInputField("New Title", issue.Title, 60, nil, func(text string) {
		newTitle = text
	})

	form.AddButton("Save (Ctrl-S)", func() {
		if newTitle == "" {
			h.StatusBar.SetText(fmt.Sprintf("[%s]Error: Title cannot be empty[-]", formatting.GetErrorColor()))
			return
		}

		// Execute bd update command with --json
		log.Printf("BD COMMAND: Renaming issue: bd update %s --title %q", issue.ID, newTitle)
		updatedIssue, err := execBdJSONIssue("update", issue.ID, "--title", newTitle)
		if err != nil {
			log.Printf("BD COMMAND ERROR: Rename failed: %v", err)
			h.StatusBar.SetText(fmt.Sprintf("[%s]Error renaming issue: %v[-]", formatting.GetErrorColor(), err))
		} else {
			log.Printf("BD COMMAND: Issue renamed successfully: %s", updatedIssue.Title)
			h.StatusBar.SetText(fmt.Sprintf("[%s]✓ Renamed %s[-]", formatting.GetSuccessColor(), updatedIssue.ID))

			// Close dialog
			h.Pages.RemovePage("rename_dialog")
			h.App.SetFocus(h.IssueList)

			// Refresh issues after a short delay, preserving selection
			issueID := issue.ID
			time.AfterFunc(500*time.Millisecond, func() {
				h.RefreshIssues(issueID)
			})
		}
	})
	form.AddButton("Cancel", func() {
		h.Pages.RemovePage("rename_dialog")
		h.App.SetFocus(h.IssueList)
	})

	form.SetBorder(true).SetTitle(" Rename Issue ").SetTitleAlign(tview.AlignCenter)
	form.SetCancelFunc(func() {
		h.Pages.RemovePage("rename_dialog")
		h.App.SetFocus(h.IssueList)
	})

	// Add Ctrl-S handler for save
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlS {
			// Save directly
			if newTitle == "" {
				h.StatusBar.SetText(fmt.Sprintf("[%s]Error: Title cannot be empty[-]", formatting.GetErrorColor()))
				return nil
			}

			log.Printf("BD COMMAND: Renaming issue: bd update %s --title %q", issue.ID, newTitle)
			updatedIssue, err := execBdJSONIssue("update", issue.ID, "--title", newTitle)
			if err != nil {
				log.Printf("BD COMMAND ERROR: Rename failed: %v", err)
				h.StatusBar.SetText(fmt.Sprintf("[%s]Error renaming issue: %v[-]", formatting.GetErrorColor(), err))
			} else {
				log.Printf("BD COMMAND: Issue renamed successfully: %s", updatedIssue.Title)
				h.StatusBar.SetText(fmt.Sprintf("[%s]✓ Renamed %s[-]", formatting.GetSuccessColor(), updatedIssue.ID))
				h.Pages.RemovePage("rename_dialog")
				h.App.SetFocus(h.IssueList)
				issueID := issue.ID
				time.AfterFunc(500*time.Millisecond, func() {
					h.RefreshIssues(issueID)
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
			AddItem(form, 12, 1, true).
			AddItem(nil, 0, 1, false), 80, 1, true).
		AddItem(nil, 0, 1, false)

	h.Pages.AddPage("rename_dialog", modal, true, true)
	h.App.SetFocus(form)
}

// ShowQuickFilter displays a dialog for quick filtering of issues
func (h *DialogHelpers) ShowQuickFilter() {
	form := tview.NewForm()
	var filterQuery string

	emphasisColor := formatting.GetEmphasisColor()
	accentColor := formatting.GetAccentColor()
	mutedColor := formatting.GetMutedColor()

	helpText := fmt.Sprintf(`[%s]Quick Filter Syntax:[-]
  p0-p4    Priority (e.g., 'p1' or 'p1,p2')
  bug, feature, task, epic, chore    Types
  open, in_progress, blocked, closed    Statuses
  #label   Label (e.g., '#ui' or '#bug,#urgent')

[%s]Examples:[-]
  p1 bug          P1 bugs only
  feature,task    Features and tasks
  p0,p1 open      High priority open issues
  #ui #urgent     Issues with 'ui' or 'urgent' labels

[%s]Leave empty to clear all filters[-]`, emphasisColor, accentColor, mutedColor)

	form.AddTextView("", helpText, 0, 11, false, false)
	form.AddInputField("Filter", "", 50, nil, func(text string) {
		filterQuery = text
	})

	// Apply filter function
	applyQuickFilter := func() {
		// Clear existing filters
		h.AppState.ClearAllFilters()

		if filterQuery == "" {
			// Empty query = clear all filters
			h.Pages.RemovePage("quick_filter")
			h.App.SetFocus(h.IssueList)
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

			// Check for label (starts with #)
			if strings.HasPrefix(token, "#") {
				label := strings.TrimPrefix(token, "#")
				if label != "" {
					h.AppState.ToggleLabelFilter(label)
				}
				continue
			}

			// Check for priority (p0-p4)
			if len(token) == 2 && token[0] == 'p' && token[1] >= '0' && token[1] <= '4' {
				priority := int(token[1] - '0')
				h.AppState.TogglePriorityFilter(priority)
				continue
			}

			// Check for type
			switch token {
			case "bug":
				h.AppState.ToggleTypeFilter(parser.TypeBug)
			case "feature":
				h.AppState.ToggleTypeFilter(parser.TypeFeature)
			case "task":
				h.AppState.ToggleTypeFilter(parser.TypeTask)
			case "epic":
				h.AppState.ToggleTypeFilter(parser.TypeEpic)
			case "chore":
				h.AppState.ToggleTypeFilter(parser.TypeChore)
			}

			// Check for status
			switch token {
			case "open":
				h.AppState.ToggleStatusFilter(parser.StatusOpen)
			case "in_progress", "inprogress":
				h.AppState.ToggleStatusFilter(parser.StatusInProgress)
			case "blocked":
				h.AppState.ToggleStatusFilter(parser.StatusBlocked)
			case "closed":
				h.AppState.ToggleStatusFilter(parser.StatusClosed)
			}
		}

		// Apply filters
		h.Pages.RemovePage("quick_filter")
		h.App.SetFocus(h.IssueList)
	}

	form.AddButton("Apply (Enter)", applyQuickFilter)
	form.AddButton("Clear All", func() {
		h.AppState.ClearAllFilters()
		h.Pages.RemovePage("quick_filter")
		h.App.SetFocus(h.IssueList)
	})
	form.AddButton("Cancel (ESC)", func() {
		h.Pages.RemovePage("quick_filter")
		h.App.SetFocus(h.IssueList)
	})

	form.SetBorder(true).SetTitle(" Quick Filter ").SetTitleAlign(tview.AlignCenter)
	form.SetCancelFunc(func() {
		h.Pages.RemovePage("quick_filter")
		h.App.SetFocus(h.IssueList)
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

	h.Pages.AddPage("quick_filter", modal, true, true)
	h.App.SetFocus(form)
}

// ShowStatsOverlay displays a statistics dashboard
func (h *DialogHelpers) ShowStatsOverlay() {
	allIssues := h.AppState.GetAllIssues()

	// Calculate statistics
	stats := struct {
		total        int
		byStatus     map[parser.Status]int
		byPriority   map[int]int
		byType       map[parser.IssueType]int
		totalDeps    int
		avgDepsPerIssue float64
	}{
		byStatus:   make(map[parser.Status]int),
		byPriority: make(map[int]int),
		byType:     make(map[parser.IssueType]int),
	}

	stats.total = len(allIssues)
	totalDeps := 0

	for _, issue := range allIssues {
		stats.byStatus[issue.Status]++
		stats.byPriority[issue.Priority]++
		stats.byType[issue.IssueType]++
		totalDeps += len(issue.Dependencies)
	}

	stats.totalDeps = totalDeps
	if stats.total > 0 {
		stats.avgDepsPerIssue = float64(totalDeps) / float64(stats.total)
	}

	// Build stats text
	var sb strings.Builder
	emphasisColor := formatting.GetEmphasisColor()
	accentColor := formatting.GetAccentColor()
	mutedColor := formatting.GetMutedColor()
	priorityColors := [5]string{
		formatting.GetPriorityColor(0),
		formatting.GetPriorityColor(1),
		formatting.GetPriorityColor(2),
		formatting.GetPriorityColor(3),
		formatting.GetPriorityColor(4),
	}

	sb.WriteString(fmt.Sprintf("[%s::b]Issue Statistics Dashboard[-::-]\n\n", emphasisColor))

	// Overall stats
	sb.WriteString(fmt.Sprintf("[%s::b]Total Issues:[-::-] %d\n\n", accentColor, stats.total))

	// By Status
	sb.WriteString(fmt.Sprintf("[%s::b]By Status:[-::-]\n", accentColor))
	sb.WriteString(fmt.Sprintf("  [%s]Ready[-]:        %3d  (%.1f%%)\n",
		formatting.GetStatusColor(parser.StatusOpen),
		stats.byStatus[parser.StatusOpen],
		float64(stats.byStatus[parser.StatusOpen])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  [%s]In Progress[-]: %3d  (%.1f%%)\n",
		formatting.GetStatusColor(parser.StatusInProgress),
		stats.byStatus[parser.StatusInProgress],
		float64(stats.byStatus[parser.StatusInProgress])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  [%s]Blocked[-]:     %3d  (%.1f%%)\n",
		formatting.GetStatusColor(parser.StatusBlocked),
		stats.byStatus[parser.StatusBlocked],
		float64(stats.byStatus[parser.StatusBlocked])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  [%s]Closed[-]:      %3d  (%.1f%%)\n\n",
		formatting.GetStatusColor(parser.StatusClosed),
		stats.byStatus[parser.StatusClosed],
		float64(stats.byStatus[parser.StatusClosed])/float64(stats.total)*100))

	// By Priority
	sb.WriteString(fmt.Sprintf("[%s::b]By Priority:[-::-]\n", accentColor))
	sb.WriteString(fmt.Sprintf("  [%s]P0 (Critical)[-]: %3d  (%.1f%%)\n",
		priorityColors[0],
		stats.byPriority[0],
		float64(stats.byPriority[0])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  [%s]P1 (High)[-]:     %3d  (%.1f%%)\n",
		priorityColors[1],
		stats.byPriority[1],
		float64(stats.byPriority[1])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  [%s]P2 (Normal)[-]:   %3d  (%.1f%%)\n",
		priorityColors[2],
		stats.byPriority[2],
		float64(stats.byPriority[2])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  [%s]P3 (Low)[-]:      %3d  (%.1f%%)\n",
		priorityColors[3],
		stats.byPriority[3],
		float64(stats.byPriority[3])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  [%s]P4 (Lowest)[-]:   %3d  (%.1f%%)\n\n",
		priorityColors[4],
		stats.byPriority[4],
		float64(stats.byPriority[4])/float64(stats.total)*100))

	// By Type
	sb.WriteString(fmt.Sprintf("[%s::b]By Type:[-::-]\n", accentColor))
	sb.WriteString(fmt.Sprintf("  Bug:      %3d  (%.1f%%)\n",
		stats.byType[parser.TypeBug],
		float64(stats.byType[parser.TypeBug])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  Feature:  %3d  (%.1f%%)\n",
		stats.byType[parser.TypeFeature],
		float64(stats.byType[parser.TypeFeature])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  Task:     %3d  (%.1f%%)\n",
		stats.byType[parser.TypeTask],
		float64(stats.byType[parser.TypeTask])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  Epic:     %3d  (%.1f%%)\n",
		stats.byType[parser.TypeEpic],
		float64(stats.byType[parser.TypeEpic])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  Chore:    %3d  (%.1f%%)\n\n",
		stats.byType[parser.TypeChore],
		float64(stats.byType[parser.TypeChore])/float64(stats.total)*100))

	// Dependencies
	sb.WriteString(fmt.Sprintf("[%s::b]Dependencies:[-::-]\n", accentColor))
	sb.WriteString(fmt.Sprintf("  Total:           %d\n", stats.totalDeps))
	sb.WriteString(fmt.Sprintf("  Avg per issue:   %.2f\n", stats.avgDepsPerIssue))

	sb.WriteString(fmt.Sprintf("\n[%s]━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━[-]\n", mutedColor))
	sb.WriteString(fmt.Sprintf("[%s]Press ESC or S to close[-]", emphasisColor))

	// Create stats text view
	statsTextView := tview.NewTextView().
		SetDynamicColors(true).
		SetText(sb.String()).
		SetTextAlign(tview.AlignLeft)
	statsTextView.SetBorder(true).
		SetTitle(" Statistics Dashboard ").
		SetTitleAlign(tview.AlignCenter)

	// Create modal (centered, slightly smaller than help)
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(statsTextView, 0, 2, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)

	// Add input capture to close on ESC or S
	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || (event.Key() == tcell.KeyRune && (event.Rune() == 'S' || event.Rune() == 's')) {
			h.Pages.RemovePage("stats")
			h.App.SetFocus(h.IssueList)
			return nil
		}
		return event
	})

	// Show modal
	h.Pages.AddPage("stats", modal, true, true)
	h.App.SetFocus(modal)
}

// ShowHelpScreen displays the keyboard shortcuts help screen
func (h *DialogHelpers) ShowHelpScreen() {
	// Note: This help screen uses hardcoded colors for documentation purposes
	// showing the current theme's colors as examples
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
  R           Rename issue (edit title)
  a           Create new issue (vim-style "add")
  c           Add comment to selected issue
  e           Edit issue (title, description, design, acceptance, notes, priority, type)
  x           Close issue with optional reason
  X           Reopen closed issue with optional reason
  D           Manage dependencies (add/remove blocks, parent-child, related)
  L           Manage labels (add/remove labels)
  y           Yank (copy) issue ID to clipboard
  Y           Yank (copy) issue ID with title to clipboard
  B           Copy git branch name to clipboard

[cyan::b]Two-Character Shortcuts[-::-]
  So          Set status to open
  Si          Set status to in_progress
  Sb          Set status to blocked
  Sc          Set status to closed

[cyan::b]View Controls[-::-]
  t           Toggle between list and tree view
  C           Toggle showing closed issues in list view
  f           Quick filter (type: p1 bug, feature, etc.)
  S           Show statistics dashboard
  m           Toggle mouse mode on/off
  r           Manual refresh

[cyan::b]Detail Panel Scrolling (when focused)[-::-]
  Ctrl-d      Scroll down half page
  Ctrl-u      Scroll up half page
  Ctrl-f      Scroll down full page (vim)
  Ctrl-b      Scroll up full page (vim)
  Ctrl-e      Scroll down one line
  Ctrl-y      Scroll up one line
  PageDown    Scroll down full page
  PageUp      Scroll up full page
  Home        Jump to top of details
  End         Jump to bottom of details

[cyan::b]General[-::-]
  ?           Show this help screen
  q           Quit

[cyan::b]Command Line Options[-::-]
  --theme <name>      Set color theme
    beads-tui --theme gruvbox-dark

  --view <mode>       Start in list or tree view
    beads-tui --view tree

  --issue <id>        Show only a specific issue
    beads-tui --issue tui-abc

  --debug             Enable debug logging

[cyan::b]Themes[-::-]
  Available themes: default, gruvbox-dark, gruvbox-light, nord,
  solarized-dark, solarized-light, dracula, tokyo-night,
  catppuccin-mocha, catppuccin-latte

  Set via environment variable:
    export BEADS_THEME=gruvbox-dark

[cyan::b]Status Icons[-::-]
  ●           Open/Ready
  ○           Blocked
  ◆           In Progress
  ·           Other

[cyan::b]Priority Colors[-::-]
  [red]P0[-]          Critical
  [orangered]P1[-]          High
  [lightskyblue]P2[-]          Normal
  [gray]P3[-]          Low
  [gray]P4[-]          Lowest

[cyan::b]Status Colors[-::-]
  [limegreen]●[-]           Ready
  [gold]○[-]           Blocked
  [deepskyblue]◆[-]           In Progress
  [gray]·[-]           Closed

[gray]━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━[-]
[yellow]Press ESC or ? to close this help[-]`

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
			h.Pages.RemovePage("help")
			h.App.SetFocus(h.IssueList)
			return nil
		}
		return event
	})

	h.Pages.AddPage("help", modal, true, true)
	h.App.SetFocus(modal)
}

// ShowDependencyDialog displays a dialog for managing dependencies
func (h *DialogHelpers) ShowDependencyDialog() {
	// Get current issue
	currentIndex := h.IssueList.GetCurrentItem()
	issue, ok := (*h.IndexToIssue)[currentIndex]
	if !ok {
		h.StatusBar.SetText(fmt.Sprintf("[%s]No issue selected[-]", formatting.GetErrorColor()))
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
			h.StatusBar.SetText(fmt.Sprintf("[%s]Error: Issue ID required[-]", formatting.GetErrorColor()))
			return
		}

		// Validate target issue exists
		if h.AppState.GetIssueByID(targetID) == nil {
			h.StatusBar.SetText(fmt.Sprintf("[%s]Error: Issue %s not found[-]", formatting.GetErrorColor(), targetID))
			return
		}

		issueID := issue.ID // Capture before potential refresh
		log.Printf("BD COMMAND: Adding dependency: bd dep add %s %s --type %s", issueID, targetID, depType)
		updatedIssue, err := execBdJSONIssue("dep", "add", issueID, targetID, "--type", depType)
		if err != nil {
			log.Printf("BD COMMAND ERROR: Dependency add failed: %v", err)
			h.StatusBar.SetText(fmt.Sprintf("[%s]Error adding dependency: %v[-]", formatting.GetErrorColor(), err))
		} else {
			log.Printf("BD COMMAND: Dependency added successfully to %s", updatedIssue.ID)
			h.StatusBar.SetText(fmt.Sprintf("[%s]✓ Added [%s]%s[-] dependency to [%s]%s[-][-]", formatting.GetSuccessColor(), formatting.GetEmphasisColor(), depType, formatting.GetAccentColor(), targetID))
			h.Pages.RemovePage("dependency_dialog")
			h.App.SetFocus(h.IssueList)
			time.AfterFunc(500*time.Millisecond, func() {
				h.RefreshIssues(issueID)
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
				log.Printf("BD COMMAND: Removing dependency: bd dep remove %s %s --type %s", issueID, depToRemove.DependsOnID, depToRemove.Type)
				updatedIssue, err := execBdJSONIssue("dep", "remove", issueID, depToRemove.DependsOnID, "--type", string(depToRemove.Type))
				if err != nil {
					log.Printf("BD COMMAND ERROR: Dependency remove failed: %v", err)
					h.StatusBar.SetText(fmt.Sprintf("[%s]Error removing dependency: %v[-]", formatting.GetErrorColor(), err))
				} else {
					log.Printf("BD COMMAND: Dependency removed successfully from %s", updatedIssue.ID)
					h.StatusBar.SetText(fmt.Sprintf("[%s]✓ Removed [%s]%s[-] dependency to [%s]%s[-][-]", formatting.GetSuccessColor(), formatting.GetEmphasisColor(), depToRemove.Type, formatting.GetAccentColor(), depToRemove.DependsOnID))
					h.Pages.RemovePage("dependency_dialog")
					h.App.SetFocus(h.IssueList)
					time.AfterFunc(500*time.Millisecond, func() {
						h.RefreshIssues(issueID)
					})
				}
			})
		}
	}

	// Close button
	form.AddButton("Close", func() {
		h.Pages.RemovePage("dependency_dialog")
		h.App.SetFocus(h.IssueList)
	})

	form.SetBorder(true).SetTitle(" Manage Dependencies ").SetTitleAlign(tview.AlignCenter)
	form.SetCancelFunc(func() {
		h.Pages.RemovePage("dependency_dialog")
		h.App.SetFocus(h.IssueList)
	})

	// Create modal (centered)
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 0, 3, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)

	h.Pages.AddPage("dependency_dialog", modal, true, true)
	h.App.SetFocus(form)
}

// ShowLabelDialog displays a dialog for managing labels
func (h *DialogHelpers) ShowLabelDialog() {
	// Get current issue
	currentIndex := h.IssueList.GetCurrentItem()
	issue, ok := (*h.IndexToIssue)[currentIndex]
	if !ok {
		h.StatusBar.SetText(fmt.Sprintf("[%s]No issue selected[-]", formatting.GetErrorColor()))
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
			h.StatusBar.SetText(fmt.Sprintf("[%s]Error: Label cannot be empty[-]", formatting.GetErrorColor()))
			return
		}

		// Check if label already exists
		for _, existing := range issue.Labels {
			if existing == trimmedLabel {
				h.StatusBar.SetText(fmt.Sprintf("[%s]Error: Label '%s' already exists[-]", formatting.GetErrorColor(), trimmedLabel))
				return
			}
		}

		issueID := issue.ID // Capture before potential refresh
		log.Printf("BD COMMAND: Adding label: bd label add %s %q", issueID, trimmedLabel)
		updatedIssue, err := execBdJSONIssue("label", "add", issueID, trimmedLabel)
		if err != nil {
			log.Printf("BD COMMAND ERROR: Label add failed: %v", err)
			h.StatusBar.SetText(fmt.Sprintf("[%s]Error adding label: %v[-]", formatting.GetErrorColor(), err))
		} else {
			log.Printf("BD COMMAND: Label added successfully to %s", updatedIssue.ID)
			h.StatusBar.SetText(fmt.Sprintf("[%s]✓ Added label [%s]'%s'[-][-]", formatting.GetSuccessColor(), formatting.GetEmphasisColor(), trimmedLabel))
			h.Pages.RemovePage("label_dialog")
			h.App.SetFocus(h.IssueList)
			time.AfterFunc(500*time.Millisecond, func() {
				h.RefreshIssues(issueID)
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
				log.Printf("BD COMMAND: Removing label: bd label remove %s %q", issueID, labelToRemove)
				updatedIssue, err := execBdJSONIssue("label", "remove", issueID, labelToRemove)
				if err != nil {
					log.Printf("BD COMMAND ERROR: Label remove failed: %v", err)
					h.StatusBar.SetText(fmt.Sprintf("[%s]Error removing label: %v[-]", formatting.GetErrorColor(), err))
				} else {
					log.Printf("BD COMMAND: Label removed successfully from %s", updatedIssue.ID)
					h.StatusBar.SetText(fmt.Sprintf("[%s]✓ Removed label [%s]'%s'[-][-]", formatting.GetSuccessColor(), formatting.GetEmphasisColor(), labelToRemove))
					h.Pages.RemovePage("label_dialog")
					h.App.SetFocus(h.IssueList)
					time.AfterFunc(500*time.Millisecond, func() {
						h.RefreshIssues(issueID)
					})
				}
			})
		}
	}

	// Close button
	form.AddButton("Close", func() {
		h.Pages.RemovePage("label_dialog")
		h.App.SetFocus(h.IssueList)
	})

	form.SetBorder(true).SetTitle(" Manage Labels ").SetTitleAlign(tview.AlignCenter)
	form.SetCancelFunc(func() {
		h.Pages.RemovePage("label_dialog")
		h.App.SetFocus(h.IssueList)
	})

	// Create modal (centered)
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 0, 3, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)

	h.Pages.AddPage("label_dialog", modal, true, true)
	h.App.SetFocus(form)
}

// ShowCloseIssueDialog displays a dialog for closing an issue
func (h *DialogHelpers) ShowCloseIssueDialog() {
	// Get current issue
	currentIndex := h.IssueList.GetCurrentItem()
	issue, ok := (*h.IndexToIssue)[currentIndex]
	if !ok {
		h.StatusBar.SetText(fmt.Sprintf("[%s]No issue selected[-]", formatting.GetErrorColor()))
		return
	}

	// Don't allow closing already closed issues
	if issue.Status == parser.StatusClosed {
		h.StatusBar.SetText(fmt.Sprintf("[%s]Issue is already closed[-]", formatting.GetWarningColor()))
		return
	}

	form := tview.NewForm()
	var reason string

	form.AddTextView("Closing", issue.ID+" - "+issue.Title, 0, 2, false, false)
	form.AddInputField("Reason (optional)", "", 60, nil, func(text string) {
		reason = text
	})

	form.AddButton("Close Issue", func() {
		issueID := issue.ID // Capture before potential refresh
		args := []string{"close", issueID}
		if reason != "" {
			args = append(args, "--reason", reason)
		}
		log.Printf("BD COMMAND: Closing issue: bd %s", strings.Join(args, " "))
		closedIssue, err := execBdJSONIssue(args...)
		if err != nil {
			log.Printf("BD COMMAND ERROR: Close failed: %v", err)
			h.StatusBar.SetText(fmt.Sprintf("[%s]Error closing issue: %v[-]", formatting.GetErrorColor(), err))
		} else {
			log.Printf("BD COMMAND: Issue closed successfully: %s", closedIssue.ID)
			h.StatusBar.SetText(fmt.Sprintf("[%s]✓ Closed [%s]%s[-][-]", formatting.GetSuccessColor(), formatting.GetAccentColor(), closedIssue.ID))
			h.Pages.RemovePage("close_issue_dialog")
			h.App.SetFocus(h.IssueList)
			time.AfterFunc(500*time.Millisecond, func() {
				h.RefreshIssues(issueID)
			})
		}
	})
	form.AddButton("Cancel", func() {
		h.Pages.RemovePage("close_issue_dialog")
		h.App.SetFocus(h.IssueList)
	})

	form.SetBorder(true).SetTitle(" Close Issue (Enter to submit) ").SetTitleAlign(tview.AlignCenter)
	form.SetCancelFunc(func() {
		h.Pages.RemovePage("close_issue_dialog")
		h.App.SetFocus(h.IssueList)
	})

	// Add Enter key handler to close
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			issueID := issue.ID
			args := []string{"close", issueID}
			if reason != "" {
				args = append(args, "--reason", reason)
			}
			log.Printf("BD COMMAND: Closing issue (Enter): bd %s", strings.Join(args, " "))
			closedIssue, err := execBdJSONIssue(args...)
			if err != nil {
				log.Printf("BD COMMAND ERROR: Close failed: %v", err)
				h.StatusBar.SetText(fmt.Sprintf("[%s]Error closing issue: %v[-]", formatting.GetErrorColor(), err))
			} else {
				log.Printf("BD COMMAND: Issue closed successfully: %s", closedIssue.ID)
				h.StatusBar.SetText(fmt.Sprintf("[%s]✓ Closed [%s]%s[-][-]", formatting.GetSuccessColor(), formatting.GetAccentColor(), closedIssue.ID))
				h.Pages.RemovePage("close_issue_dialog")
				h.App.SetFocus(h.IssueList)
				time.AfterFunc(500*time.Millisecond, func() {
					h.RefreshIssues(issueID)
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
			AddItem(form, 0, 2, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)

	h.Pages.AddPage("close_issue_dialog", modal, true, true)
	h.App.SetFocus(form)
}

// ShowReopenIssueDialog displays a dialog for reopening a closed issue
func (h *DialogHelpers) ShowReopenIssueDialog() {
	// Get current issue
	currentIndex := h.IssueList.GetCurrentItem()
	issue, ok := (*h.IndexToIssue)[currentIndex]
	if !ok {
		h.StatusBar.SetText(fmt.Sprintf("[%s]No issue selected[-]", formatting.GetErrorColor()))
		return
	}

	// Only allow reopening closed issues
	if issue.Status != parser.StatusClosed {
		h.StatusBar.SetText(fmt.Sprintf("[%s]Issue is not closed[-]", formatting.GetWarningColor()))
		return
	}

	form := tview.NewForm()
	var reason string

	form.AddTextView("Reopening", issue.ID+" - "+issue.Title, 0, 2, false, false)
	form.AddInputField("Reason (optional)", "", 60, nil, func(text string) {
		reason = text
	})

	form.AddButton("Reopen Issue", func() {
		issueID := issue.ID // Capture before potential refresh
		args := []string{"reopen", issueID}
		if reason != "" {
			args = append(args, "--reason", reason)
		}
		log.Printf("BD COMMAND: Reopening issue: bd %s", strings.Join(args, " "))
		reopenedIssue, err := execBdJSONIssue(args...)
		if err != nil {
			log.Printf("BD COMMAND ERROR: Reopen failed: %v", err)
			h.StatusBar.SetText(fmt.Sprintf("[%s]Error reopening issue: %v[-]", formatting.GetErrorColor(), err))
		} else {
			log.Printf("BD COMMAND: Issue reopened successfully: %s", reopenedIssue.ID)
			h.StatusBar.SetText(fmt.Sprintf("[%s]✓ Reopened [%s]%s[-][-]", formatting.GetSuccessColor(), formatting.GetAccentColor(), reopenedIssue.ID))
			h.Pages.RemovePage("reopen_issue_dialog")
			h.App.SetFocus(h.IssueList)
			time.AfterFunc(500*time.Millisecond, func() {
				h.RefreshIssues(issueID)
			})
		}
	})
	form.AddButton("Cancel", func() {
		h.Pages.RemovePage("reopen_issue_dialog")
		h.App.SetFocus(h.IssueList)
	})

	form.SetBorder(true).SetTitle(" Reopen Issue (Enter to submit) ").SetTitleAlign(tview.AlignCenter)
	form.SetCancelFunc(func() {
		h.Pages.RemovePage("reopen_issue_dialog")
		h.App.SetFocus(h.IssueList)
	})

	// Add Enter key handler to reopen
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			issueID := issue.ID
			args := []string{"reopen", issueID}
			if reason != "" {
				args = append(args, "--reason", reason)
			}
			log.Printf("BD COMMAND: Reopening issue (Enter): bd %s", strings.Join(args, " "))
			reopenedIssue, err := execBdJSONIssue(args...)
			if err != nil {
				log.Printf("BD COMMAND ERROR: Reopen failed: %v", err)
				h.StatusBar.SetText(fmt.Sprintf("[%s]Error reopening issue: %v[-]", formatting.GetErrorColor(), err))
			} else {
				log.Printf("BD COMMAND: Issue reopened successfully: %s", reopenedIssue.ID)
				h.StatusBar.SetText(fmt.Sprintf("[%s]✓ Reopened [%s]%s[-][-]", formatting.GetSuccessColor(), formatting.GetAccentColor(), reopenedIssue.ID))
				h.Pages.RemovePage("reopen_issue_dialog")
				h.App.SetFocus(h.IssueList)
				time.AfterFunc(500*time.Millisecond, func() {
					h.RefreshIssues(issueID)
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
			AddItem(form, 0, 2, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)

	h.Pages.AddPage("reopen_issue_dialog", modal, true, true)
	h.App.SetFocus(form)
}

// ShowEditForm displays a dialog for editing all issue fields
func (h *DialogHelpers) ShowEditForm() {
	// Get current issue
	currentIndex := h.IssueList.GetCurrentItem()
	issue, ok := (*h.IndexToIssue)[currentIndex]
	if !ok {
		h.StatusBar.SetText(fmt.Sprintf("[%s]No issue selected[-]", formatting.GetErrorColor()))
		return
	}

	form := tview.NewForm()
	var title, description, design, acceptance, notes string
	var priority int
	var issueType string

	// Initialize with current values
	title = issue.Title
	description = issue.Description
	design = issue.Design
	acceptance = issue.AcceptanceCriteria
	notes = issue.Notes
	priority = issue.Priority
	issueType = string(issue.IssueType)

	form.AddTextView("Editing", issue.ID, 0, 1, false, false)
	form.AddInputField("Title", title, 60, nil, func(text string) {
		title = text
	})
	form.AddTextArea("Description", description, 60, 5, 0, func(text string) {
		description = text
	})
	form.AddTextArea("Design", design, 60, 5, 0, func(text string) {
		design = text
	})
	form.AddTextArea("Acceptance Criteria", acceptance, 60, 5, 0, func(text string) {
		acceptance = text
	})
	form.AddTextArea("Notes", notes, 60, 5, 0, func(text string) {
		notes = text
	})
	form.AddDropDown("Priority", []string{"P0 (Critical)", "P1 (High)", "P2 (Normal)", "P3 (Low)", "P4 (Lowest)"}, priority, func(option string, index int) {
		priority = index
	})

	// Find index of current type
	typeOptions := []string{"bug", "feature", "task", "epic", "chore"}
	typeIndex := 1 // default to feature
	for i, t := range typeOptions {
		if t == issueType {
			typeIndex = i
			break
		}
	}
	form.AddDropDown("Type", typeOptions, typeIndex, func(option string, index int) {
		issueType = option
	})

	// Save function
	saveChanges := func() {
		issueID := issue.ID // Capture before potential refresh

		// Build update command with all fields
		// Use temp files to avoid shell escaping issues
		titleFile := filepath.Join(os.TempDir(), fmt.Sprintf("beads-tui-title-%s.txt", issueID))
		descFile := filepath.Join(os.TempDir(), fmt.Sprintf("beads-tui-desc-%s.txt", issueID))
		designFile := filepath.Join(os.TempDir(), fmt.Sprintf("beads-tui-design-%s.txt", issueID))
		acceptFile := filepath.Join(os.TempDir(), fmt.Sprintf("beads-tui-accept-%s.txt", issueID))
		notesFile := filepath.Join(os.TempDir(), fmt.Sprintf("beads-tui-notes-%s.txt", issueID))

		defer os.Remove(titleFile)
		defer os.Remove(descFile)
		defer os.Remove(designFile)
		defer os.Remove(acceptFile)
		defer os.Remove(notesFile)

		if err := os.WriteFile(titleFile, []byte(title), 0600); err != nil {
			h.StatusBar.SetText(fmt.Sprintf("[%s]Error: %v[-]", formatting.GetErrorColor(), err))
			return
		}
		if err := os.WriteFile(descFile, []byte(description), 0600); err != nil {
			h.StatusBar.SetText(fmt.Sprintf("[%s]Error: %v[-]", formatting.GetErrorColor(), err))
			return
		}
		if err := os.WriteFile(designFile, []byte(design), 0600); err != nil {
			h.StatusBar.SetText(fmt.Sprintf("[%s]Error: %v[-]", formatting.GetErrorColor(), err))
			return
		}
		if err := os.WriteFile(acceptFile, []byte(acceptance), 0600); err != nil {
			h.StatusBar.SetText(fmt.Sprintf("[%s]Error: %v[-]", formatting.GetErrorColor(), err))
			return
		}
		if err := os.WriteFile(notesFile, []byte(notes), 0600); err != nil {
			h.StatusBar.SetText(fmt.Sprintf("[%s]Error: %v[-]", formatting.GetErrorColor(), err))
			return
		}

		cmd := fmt.Sprintf("bd update %s --title \"$(cat %s)\" --description \"$(cat %s)\" --design \"$(cat %s)\" --acceptance \"$(cat %s)\" --notes \"$(cat %s)\" --priority %d --type %s --json",
			issueID, titleFile, descFile, designFile, acceptFile, notesFile, priority, issueType)

		log.Printf("BD COMMAND: Updating issue: bd update %s ...", issueID)
		output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
		if err != nil {
			log.Printf("BD COMMAND ERROR: Update failed: %v, output: %s", err, string(output))
			h.StatusBar.SetText(fmt.Sprintf("[%s]Error updating issue: %v[-]", formatting.GetErrorColor(), err))
		} else {
			// Parse JSON response to verify success
			result, parseErr := parseBdJSON(output)
			if parseErr != nil {
				log.Printf("BD COMMAND ERROR: Failed to parse response: %v", parseErr)
				h.StatusBar.SetText(fmt.Sprintf("[%s]Error parsing response: %v[-]", formatting.GetErrorColor(), parseErr))
			} else if len(result.Issues) > 0 {
				updatedIssue := result.Issues[0]
				log.Printf("BD COMMAND: Issue updated successfully: %s", updatedIssue.Title)
				h.StatusBar.SetText(fmt.Sprintf("[%s]✓ Updated [%s]%s[-][-]", formatting.GetSuccessColor(), formatting.GetAccentColor(), updatedIssue.ID))
				h.Pages.RemovePage("edit_form")
				h.App.SetFocus(h.IssueList)
				time.AfterFunc(500*time.Millisecond, func() {
					h.RefreshIssues(issueID)
				})
			}
		}
	}

	form.AddButton("Save (Ctrl-S)", saveChanges)
	form.AddButton("Cancel", func() {
		h.Pages.RemovePage("edit_form")
		h.App.SetFocus(h.IssueList)
	})

	form.SetBorder(true).SetTitle(" Edit Issue ").SetTitleAlign(tview.AlignCenter)
	form.SetCancelFunc(func() {
		h.Pages.RemovePage("edit_form")
		h.App.SetFocus(h.IssueList)
	})

	// Add Ctrl-S handler for save
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlS {
			saveChanges()
			return nil
		}
		return event
	})

	// Create modal (centered, larger for editing)
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 0, 4, true).
			AddItem(nil, 0, 1, false), 0, 3, true).
		AddItem(nil, 0, 1, false)

	h.Pages.AddPage("edit_form", modal, true, true)
	h.App.SetFocus(form)
}

// ShowCreateIssueDialog displays a dialog for creating a new issue
func (h *DialogHelpers) ShowCreateIssueDialog() {
	// Helper function to detect priority from text (natural language)
	detectPriority := func(text string) *int {
		lower := strings.ToLower(text)
		// P0 keywords: critical, urgent, blocking, blocker, emergency, outage, down, broken
		if strings.Contains(lower, "critical") || strings.Contains(lower, "urgent") ||
			strings.Contains(lower, "blocking") || strings.Contains(lower, "blocker") ||
			strings.Contains(lower, "emergency") || strings.Contains(lower, "outage") ||
			strings.Contains(lower, "down") || strings.Contains(lower, "broken") {
			p := 0
			return &p
		}
		// P1 keywords: important, high priority, asap, soon, needed
		if strings.Contains(lower, "important") || strings.Contains(lower, "high priority") ||
			strings.Contains(lower, "asap") || strings.Contains(lower, "soon") ||
			strings.Contains(lower, "needed") || strings.Contains(lower, "must have") {
			p := 1
			return &p
		}
		// P3 keywords: low priority, minor, nice to have, eventually, someday
		if strings.Contains(lower, "low priority") || strings.Contains(lower, "minor") ||
			strings.Contains(lower, "nice to have") || strings.Contains(lower, "eventually") ||
			strings.Contains(lower, "someday") || strings.Contains(lower, "polish") {
			p := 3
			return &p
		}
		// P4 keywords: trivial, cosmetic, optional
		if strings.Contains(lower, "trivial") || strings.Contains(lower, "cosmetic") ||
			strings.Contains(lower, "optional") {
			p := 4
			return &p
		}
		return nil // No match, keep default
	}

	// Helper function to detect issue type from text (natural language)
	detectIssueType := func(text string) *string {
		lower := strings.ToLower(text)
		// Bug keywords: bug, error, crash, fix, broken, issue, problem, regression
		if strings.Contains(lower, "bug") || strings.Contains(lower, "error") ||
			strings.Contains(lower, "crash") || strings.Contains(lower, "fix ") ||
			strings.Contains(lower, "broken") || strings.Contains(lower, "problem") ||
			strings.Contains(lower, "regression") {
			t := "bug"
			return &t
		}
		// Epic keywords: epic, project, initiative, milestone (check before task)
		if strings.Contains(lower, "epic") || strings.Contains(lower, "project") ||
			strings.Contains(lower, "initiative") || strings.Contains(lower, "milestone") {
			t := "epic"
			return &t
		}
		// Chore keywords: chore, maintenance, dependency, upgrade, cleanup (check before task)
		if strings.Contains(lower, "chore") || strings.Contains(lower, "maintenance") ||
			strings.Contains(lower, "dependency") || strings.Contains(lower, "upgrade") ||
			strings.Contains(lower, "cleanup") {
			t := "chore"
			return &t
		}
		// Task keywords: task, do, implement, update, change, refactor, clean up
		if strings.Contains(lower, "task") || strings.Contains(lower, "do ") ||
			strings.Contains(lower, "implement") || strings.Contains(lower, "update") ||
			strings.Contains(lower, "change") || strings.Contains(lower, "refactor") ||
			strings.Contains(lower, "clean up") {
			t := "task"
			return &t
		}
		// Feature is default, so only explicitly detect if keywords present
		if strings.Contains(lower, "feature") || strings.Contains(lower, "add ") ||
			strings.Contains(lower, "new ") || strings.Contains(lower, "build") ||
			strings.Contains(lower, "create") {
			t := "feature"
			return &t
		}
		return nil // No match, keep default
	}

	// Create form
	form := tview.NewForm()
	form.SetHorizontal(false) // Vertical layout
	form.SetItemPadding(1) // Add spacing between fields

	// Set field colors to ensure visibility
	currentTheme := theme.Current()
	form.SetFieldBackgroundColor(currentTheme.InputFieldBackground())
	form.SetFieldTextColor(currentTheme.AppForeground())

	var title, description, priority, issueType string
	priority = "2" // Default to P2
	issueType = "feature" // Default to feature
	priorityExplicitlySet := false // Track if user manually changed priority
	typeExplicitlySet := false // Track if user manually changed type

	// Get current issue for potential parent
	var currentIssueID string
	if issue, ok := (*h.IndexToIssue)[h.IssueList.GetCurrentItem()]; ok {
		currentIssueID = issue.ID
	}

	// Create a TextView to show detected keywords
	detectionHintView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	// Helper to update priority/type from text if not explicitly set
	updateFromText := func() {
		combinedText := title + " " + description
		var hints []string

		if !priorityExplicitlySet {
			if detectedP := detectPriority(combinedText); detectedP != nil {
				priority = fmt.Sprintf("%d", *detectedP)
				// Update dropdown to reflect detected priority
				if dropdown := form.GetFormItemByLabel("Priority"); dropdown != nil {
					if dd, ok := dropdown.(*tview.DropDown); ok {
						dd.SetCurrentOption(*detectedP)
					}
				}
				// Add hint
				priorityNames := []string{"P0 (Critical)", "P1 (High)", "P2 (Normal)", "P3 (Low)", "P4 (Lowest)"}
				hints = append(hints, fmt.Sprintf("[%s]Priority:[%s] Auto-detected %s", formatting.GetEmphasisColor(), formatting.GetAccentColor(), priorityNames[*detectedP]))
			}
		}

		if !typeExplicitlySet {
			if detectedT := detectIssueType(combinedText); detectedT != nil {
				issueType = *detectedT
				// Update dropdown to reflect detected type
				if dropdown := form.GetFormItemByLabel("Type"); dropdown != nil {
					if dd, ok := dropdown.(*tview.DropDown); ok {
						typeOptions := []string{"bug", "feature", "task", "epic", "chore"}
						for i, opt := range typeOptions {
							if opt == issueType {
								dd.SetCurrentOption(i)
								break
							}
						}
					}
				}
				// Add hint
				hints = append(hints, fmt.Sprintf("[%s]Type:[%s] Auto-detected %s", formatting.GetEmphasisColor(), formatting.GetAccentColor(), *detectedT))
			}
		}

		// Update hint view
		if len(hints) > 0 {
			detectionHintView.SetText(fmt.Sprintf("[%s]%s[-]", formatting.GetMutedColor(), strings.Join(hints, " | ")))
		} else {
			detectionHintView.SetText("")
		}
	}

	// Add form fields with wide labels that force wrapping
	// This makes inputs appear below labels with full width
	form.AddInputField("Title                                                                                   ", "", 0, nil, func(text string) {
		title = text
		updateFromText()
	})
	form.AddTextArea("Description                                                                             ", "", 0, 5, 0, func(text string) {
		description = text
		updateFromText()
	})
	form.AddDropDown("Priority", []string{"P0 (Critical)", "P1 (High)", "P2 (Normal)", "P3 (Low)", "P4 (Lowest)"}, 2, func(option string, index int) {
		priority = fmt.Sprintf("%d", index)
		priorityExplicitlySet = true
	})
	form.AddDropDown("Type", []string{"bug", "feature", "task", "epic", "chore"}, 1, func(option string, index int) {
		issueType = option
		typeExplicitlySet = true
	})
	if currentIssueID != "" {
		form.AddCheckbox("Add as child of "+currentIssueID, false, nil)
	}

	// Add buttons
	form.AddButton("Create (Ctrl-S)", func() {
		if title == "" {
			h.StatusBar.SetText(fmt.Sprintf("[%s]Error: Title is required[-]", formatting.GetErrorColor()))
			return
		}

		// Build bd create command arguments
		args := []string{"create", title, "-p", priority, "-t", issueType}
		if description != "" {
			args = append(args, "--description", description)
		}

		// Check if we should add parent relationship
		if currentIssueID != "" {
			// Check checkbox state
			formItem := form.GetFormItemByLabel("Add as child of " + currentIssueID)
			if checkbox, ok := formItem.(*tview.Checkbox); ok && checkbox.IsChecked() {
				args = append(args, "--parent", currentIssueID)
			}
		}

		log.Printf("BD COMMAND: Creating issue: bd %s", strings.Join(args, " "))
		createdIssue, err := execBdJSONIssue(args...)
		if err != nil {
			log.Printf("BD COMMAND ERROR: Issue creation failed: %v", err)
			h.StatusBar.SetText(fmt.Sprintf("[%s]Error creating issue: %v[-]", formatting.GetErrorColor(), err))
		} else {
			log.Printf("BD COMMAND: Issue created successfully: %s", createdIssue.ID)
			h.StatusBar.SetText(fmt.Sprintf("[%s]✓ Created [%s]%s[-][-]", formatting.GetSuccessColor(), formatting.GetAccentColor(), createdIssue.ID))

			// Close dialog
			h.Pages.RemovePage("create_issue")
			h.App.SetFocus(h.IssueList)

			// Refresh issues after a short delay
			time.AfterFunc(500*time.Millisecond, func() {
				h.RefreshIssues()
			})
		}
	})
	form.AddButton("Cancel", func() {
		h.Pages.RemovePage("create_issue")
		h.App.SetFocus(h.IssueList)
	})

	form.SetBorder(true).SetTitle(" Create New Issue ").SetTitleAlign(tview.AlignCenter)
	form.SetCancelFunc(func() {
		h.Pages.RemovePage("create_issue")
		h.App.SetFocus(h.IssueList)
	})

	// Add Ctrl-S handler to submit form (Ctrl-Enter is reserved by terminal)
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlS {
			// Ctrl-S pressed - submit form
			if title == "" {
				h.StatusBar.SetText(fmt.Sprintf("[%s]Error: Title is required[-]", formatting.GetErrorColor()))
				return nil
			}

			// Build bd create command arguments
			args := []string{"create", title, "-p", priority, "-t", issueType}
			if description != "" {
				args = append(args, "--description", description)
			}

			// Check if we should add parent relationship
			if currentIssueID != "" {
				formItem := form.GetFormItemByLabel("Add as child of " + currentIssueID)
				if checkbox, ok := formItem.(*tview.Checkbox); ok && checkbox.IsChecked() {
					args = append(args, "--parent", currentIssueID)
				}
			}

			log.Printf("BD COMMAND: Creating issue (Ctrl-S): bd %s", strings.Join(args, " "))
			createdIssue, err := execBdJSONIssue(args...)
			if err != nil {
				log.Printf("BD COMMAND ERROR: Issue creation failed: %v", err)
				h.StatusBar.SetText(fmt.Sprintf("[%s]Error creating issue: %v[-]", formatting.GetErrorColor(), err))
			} else {
				log.Printf("BD COMMAND: Issue created successfully: %s", createdIssue.ID)
				h.StatusBar.SetText(fmt.Sprintf("[%s]✓ Created [%s]%s[-][-]", formatting.GetSuccessColor(), formatting.GetAccentColor(), createdIssue.ID))
				h.Pages.RemovePage("create_issue")
				h.App.SetFocus(h.IssueList)
				time.AfterFunc(500*time.Millisecond, func() {
					h.RefreshIssues()
				})
			}
			return nil
		}
		return event
	})

	// Create modal with hint view (centered)
	formWithHints := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(detectionHintView, 1, 0, false)

	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(formWithHints, 0, 3, true).
			AddItem(nil, 0, 1, false), 0, 3, true).
		AddItem(nil, 0, 1, false)

	h.Pages.AddPage("create_issue", modal, true, true)
	h.App.SetFocus(form)
}
