package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/andy/beads-tui/internal/formatting"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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
				h.ScheduleRefresh(issueID)
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
