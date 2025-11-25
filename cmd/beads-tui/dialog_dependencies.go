package main

import (
	"fmt"
	"log"

	"github.com/andy/beads-tui/internal/formatting"
	"github.com/andy/beads-tui/internal/parser"
	"github.com/rivo/tview"
)

// depTypeToPhrase converts a dependency type to a human-readable phrase
// from the perspective of the issue that HAS the dependency
func depTypeToPhrase(depType parser.DependencyType) string {
	switch depType {
	case parser.DepBlocks:
		return "blocked by"
	case parser.DepParentChild:
		return "child of"
	case parser.DepRelated:
		return "related to"
	case parser.DepDiscoveredFrom:
		return "discovered from"
	default:
		return string(depType)
	}
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

	// Show current dependencies with human-readable phrases
	if len(issue.Dependencies) > 0 {
		depText := "Current Dependencies:\n"
		for _, dep := range issue.Dependencies {
			phrase := depTypeToPhrase(dep.Type)
			depText += fmt.Sprintf("  %s %s\n", phrase, dep.DependsOnID)
		}
		form.AddTextView("", depText, 0, len(issue.Dependencies)+1, false, false)
	} else {
		form.AddTextView("", "No dependencies", 0, 1, false, false)
	}

	// Add new dependency fields with descriptive labels
	// The dropdown shows what relationship this issue will have TO the target
	var targetID, depType string
	form.AddInputField("Target Issue ID", "", 20, nil, func(text string) {
		targetID = text
	})
	// Use descriptive labels that explain the relationship from this issue's perspective
	depOptions := []string{
		"blocked by (this issue waits for target)",
		"child of (this issue belongs to target)",
		"related to (informational link)",
		"discovered from (provenance)",
	}
	// Map display options back to bd command values
	depTypeValues := []string{"blocks", "parent-child", "related", "discovered-from"}
	form.AddDropDown("Relationship", depOptions, 0, func(option string, index int) {
		depType = depTypeValues[index]
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
			// Show human-readable phrase in success message
			phrase := depTypeToPhrase(parser.DependencyType(depType))
			log.Printf("BD COMMAND: Dependency added successfully to %s", updatedIssue.ID)
			h.StatusBar.SetText(fmt.Sprintf("[%s]✓ Now [%s]%s[-] [%s]%s[-][-]", formatting.GetSuccessColor(), formatting.GetEmphasisColor(), phrase, formatting.GetAccentColor(), targetID))
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
			phrase := depTypeToPhrase(depToRemove.Type)
			buttonLabel := fmt.Sprintf("Remove: %s %s", phrase, depToRemove.DependsOnID)
			form.AddButton(buttonLabel, func() {
				issueID := issue.ID
				log.Printf("BD COMMAND: Removing dependency: bd dep remove %s %s --type %s", issueID, depToRemove.DependsOnID, depToRemove.Type)
				updatedIssue, err := execBdJSONIssue("dep", "remove", issueID, depToRemove.DependsOnID, "--type", string(depToRemove.Type))
				if err != nil {
					log.Printf("BD COMMAND ERROR: Dependency remove failed: %v", err)
					h.StatusBar.SetText(fmt.Sprintf("[%s]Error removing dependency: %v[-]", formatting.GetErrorColor(), err))
				} else {
					removePhrase := depTypeToPhrase(depToRemove.Type)
					log.Printf("BD COMMAND: Dependency removed successfully from %s", updatedIssue.ID)
					h.StatusBar.SetText(fmt.Sprintf("[%s]✓ No longer [%s]%s[-] [%s]%s[-][-]", formatting.GetSuccessColor(), formatting.GetEmphasisColor(), removePhrase, formatting.GetAccentColor(), depToRemove.DependsOnID))
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
