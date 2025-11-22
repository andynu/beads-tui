package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/andy/beads-tui/internal/formatting"
	"github.com/andy/beads-tui/internal/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/term"
)

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

	// Calculate field width based on terminal size
	fieldWidth := 45 // default fallback

	// Try to get terminal width from OS
	if termWidth, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && termWidth > 0 {
		dialogWidth := (termWidth * 4) / 5
		fieldWidth = (dialogWidth * 70) / 100
		// Subtract label width (approximately 15 chars for "Description")
		fieldWidth -= 15
		// Clamp to reasonable bounds
		if fieldWidth < 30 {
			fieldWidth = 30
		}
		if fieldWidth > 80 {
			fieldWidth = 80
		}
		log.Printf("DIALOG: termWidth=%d, dialogWidth=%d, fieldWidth=%d", termWidth, dialogWidth, fieldWidth)
	} else {
		log.Printf("DIALOG: Failed to get terminal size, using default fieldWidth=%d, err=%v", fieldWidth, err)
	}

	// Create form
	form := tview.NewForm()
	form.SetItemPadding(1) // Add spacing between fields

	// Set field colors - use selection colors which we know work
	currentTheme := theme.Current()
	form.SetFieldBackgroundColor(currentTheme.SelectionBg())
	form.SetFieldTextColor(currentTheme.SelectionFg())
	form.SetButtonBackgroundColor(currentTheme.SelectionBg())
	form.SetButtonTextColor(currentTheme.SelectionFg())

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

	// Add form fields with dynamic width
	form.AddInputField("Title", "", fieldWidth, nil, func(text string) {
		title = text
		updateFromText()
	})
	form.AddTextArea("Description", "", fieldWidth, 5, 0, func(text string) {
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
			h.ScheduleRefresh("")
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
				h.ScheduleRefresh("")
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
			AddItem(nil, 0, 1, false), 0, 4, true).
		AddItem(nil, 0, 1, false)

	h.Pages.AddPage("create_issue", modal, true, true)
	h.App.SetFocus(form)
}
