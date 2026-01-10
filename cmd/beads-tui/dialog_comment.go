package main

import (
	"fmt"
	"log"

	"github.com/andy/beads-tui/internal/formatting"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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

	// Define save function to be used by both button and Ctrl-S
	saveComment := func() {
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
			h.ScheduleRefresh(issueID)
		}
	}

	form.AddTextView("Adding comment to", issue.ID+" - "+issue.Title, 0, 2, false, false)
	form.AddTextArea("Comment", "", 60, 8, 0, func(text string) {
		commentText = text
	})

	// Get the TextArea and add Ctrl-S handler directly to it
	// (form's InputCapture doesn't receive events when TextArea has focus)
	if textArea, ok := form.GetFormItemByLabel("Comment").(*tview.TextArea); ok {
		textArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyCtrlS {
				saveComment()
				return nil
			}
			return event
		})
	}

	form.AddButton("Save (Ctrl-S)", saveComment)
	form.AddButton("Cancel", func() {
		h.Pages.RemovePage("comment_dialog")
		h.App.SetFocus(h.IssueList)
	})

	form.SetBorder(true).SetTitle(" Add Comment ").SetTitleAlign(tview.AlignCenter)
	form.SetCancelFunc(func() {
		h.Pages.RemovePage("comment_dialog")
		h.App.SetFocus(h.IssueList)
	})

	// Add Ctrl-S handler for when buttons have focus
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlS {
			saveComment()
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
