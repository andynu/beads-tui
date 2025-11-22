package main

import (
	"fmt"
	"log"

	"github.com/andy/beads-tui/internal/formatting"
	"github.com/rivo/tview"
)

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
			h.ScheduleRefresh(issueID)
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
					h.ScheduleRefresh(issueID)
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
