package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/andy/beads-tui/internal/formatting"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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
			h.ScheduleRefresh(issueID)
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
					h.ScheduleRefresh(issueID)
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

	// Add q key handler
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
			h.Pages.RemovePage("label_dialog")
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
			AddItem(form, 0, 3, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)

	h.Pages.AddPage("label_dialog", modal, true, true)
	h.App.SetFocus(form)
}
