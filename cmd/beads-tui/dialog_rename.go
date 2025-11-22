package main

import (
	"fmt"
	"log"

	"github.com/andy/beads-tui/internal/formatting"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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
			h.ScheduleRefresh(issueID)
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
				h.ScheduleRefresh(issueID)
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
