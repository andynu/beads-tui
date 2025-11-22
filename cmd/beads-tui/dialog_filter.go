package main

import (
	"fmt"
	"strings"

	"github.com/andy/beads-tui/internal/formatting"
	"github.com/andy/beads-tui/internal/parser"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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

	// Add Enter and q key handlers
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			applyQuickFilter()
			return nil
		}
		if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
			h.Pages.RemovePage("quick_filter")
			h.App.SetFocus(h.IssueList)
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
