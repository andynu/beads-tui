package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/andy/beads-tui/internal/parser"
	"github.com/andy/beads-tui/internal/state"
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
		h.StatusBar.SetText("[red]No issue selected[-]")
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
			h.StatusBar.SetText("[red]Error: Comment cannot be empty[-]")
			return
		}

		// Execute bd comment command
		cmd := fmt.Sprintf("bd comment %s %q", issue.ID, commentText)
		log.Printf("BD COMMAND: Adding comment: %s", cmd)
		output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
		if err != nil {
			log.Printf("BD COMMAND ERROR: Comment failed: %v, output: %s", err, string(output))
			h.StatusBar.SetText(fmt.Sprintf("[red]Error adding comment: %v[-]", err))
		} else {
			log.Printf("BD COMMAND: Comment added successfully: %s", string(output))
			h.StatusBar.SetText("[limegreen]✓ Comment added successfully[-]")

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
				h.StatusBar.SetText("[red]Error: Comment cannot be empty[-]")
				return nil
			}

			cmd := fmt.Sprintf("bd comment %s %q", issue.ID, commentText)
			log.Printf("BD COMMAND: Adding comment: %s", cmd)
			output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
			if err != nil {
				log.Printf("BD COMMAND ERROR: Comment failed: %v, output: %s", err, string(output))
				h.StatusBar.SetText(fmt.Sprintf("[red]Error adding comment: %v[-]", err))
			} else {
				log.Printf("BD COMMAND: Comment added successfully: %s", string(output))
				h.StatusBar.SetText("[limegreen]✓ Comment added successfully[-]")
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

// TODO: Add remaining dialog functions:
// - ShowRenameDialog
// - ShowQuickFilter
// - ShowStatsOverlay
// - ShowHelpScreen
// - ShowDependencyDialog
// - ShowLabelDialog
// - ShowCloseIssueDialog
// - ShowReopenIssueDialog
// - ShowEditForm
// - ShowCreateIssueDialog
