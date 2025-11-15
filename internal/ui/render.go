package ui

import (
	"fmt"

	"github.com/andy/beads-tui/internal/formatting"
	"github.com/andy/beads-tui/internal/parser"
	"github.com/andy/beads-tui/internal/state"
	"github.com/rivo/tview"
)

// PopulateIssueList clears and rebuilds the issue list from state
// Returns the updated indexToIssue map
func PopulateIssueList(
	issueList *tview.List,
	appState *state.State,
	showClosedIssues bool,
) map[int]*parser.Issue {
	// Clear and rebuild issue list
	issueList.Clear()
	indexToIssue := make(map[int]*parser.Issue)
	currentIndex := 0

	// Check view mode
	if appState.GetViewMode() == state.ViewTree {
		// Tree view
		issueList.AddItem("[cyan::b]DEPENDENCY TREE[-::-]", "", 0, nil)
		currentIndex++

		treeNodes := appState.GetTreeNodes()
		for i, node := range treeNodes {
			isLast := i == len(treeNodes)-1
			renderTreeNode(issueList, node, "", isLast, &currentIndex, indexToIssue)
		}
	} else {
		// List view (original behavior)
		// Add ready issues
		readyIssues := appState.GetReadyIssues()
		if len(readyIssues) > 0 {
			issueList.AddItem(fmt.Sprintf("[limegreen::b]⬤ READY (%d)[-::-]", len(readyIssues)), "", 0, nil)
			currentIndex++

			for _, issue := range readyIssues {
				text := formatIssueListItem(issue, "●")
				issueList.AddItem(text, "", 0, nil)
				indexToIssue[currentIndex] = issue
				currentIndex++
			}
		}

		// Add blocked issues
		blockedIssues := appState.GetBlockedIssues()
		if len(blockedIssues) > 0 {
			issueList.AddItem(fmt.Sprintf("\n[gold::b]⬤ BLOCKED (%d)[-::-]", len(blockedIssues)), "", 0, nil)
			currentIndex++

			for _, issue := range blockedIssues {
				text := formatIssueListItem(issue, "○")
				issueList.AddItem(text, "", 0, nil)
				indexToIssue[currentIndex] = issue
				currentIndex++
			}
		}

		// Add in-progress issues
		inProgressIssues := appState.GetInProgressIssues()
		if len(inProgressIssues) > 0 {
			issueList.AddItem(fmt.Sprintf("\n[deepskyblue::b]⬤ IN PROGRESS (%d)[-::-]", len(inProgressIssues)), "", 0, nil)
			currentIndex++

			for _, issue := range inProgressIssues {
				text := formatIssueListItem(issue, "◆")
				issueList.AddItem(text, "", 0, nil)
				indexToIssue[currentIndex] = issue
				currentIndex++
			}
		}

		// Add closed issues (only if showClosedIssues is enabled)
		if showClosedIssues {
			closedIssues := appState.GetClosedIssues()
			if len(closedIssues) > 0 {
				issueList.AddItem(fmt.Sprintf("\n[gray::b]⬤ CLOSED (%d)[-::-]", len(closedIssues)), "", 0, nil)
				currentIndex++

				for _, issue := range closedIssues {
					text := formatIssueListItem(issue, "✓")
					issueList.AddItem(text, "", 0, nil)
					indexToIssue[currentIndex] = issue
					currentIndex++
				}
			}
		}
	}

	return indexToIssue
}

// formatIssueListItem formats a single issue for the list view
func formatIssueListItem(issue *parser.Issue, statusIcon string) string {
	priorityColor := formatting.GetPriorityColor(issue.Priority)
	typeIcon := formatting.GetTypeIcon(issue.IssueType)
	text := fmt.Sprintf("  [%s]%s[-] %s %s [P%d] %s",
		priorityColor, statusIcon, typeIcon, issue.ID, issue.Priority, issue.Title)

	// Add labels if present
	if len(issue.Labels) > 0 {
		text += " [gray]"
		for i, label := range issue.Labels {
			if i > 0 {
				text += " "
			}
			text += "#" + label
		}
		text += "[-]"
	}

	return text
}

// renderTreeNode recursively renders a tree node and its children
func renderTreeNode(
	issueList *tview.List,
	node *state.TreeNode,
	prefix string,
	isLast bool,
	currentIndex *int,
	indexToIssue map[int]*parser.Issue,
) {
	issue := node.Issue

	// Determine branch characters
	var branch, continuation string
	if node.Depth == 0 {
		branch = ""
		continuation = ""
	} else {
		if isLast {
			branch = "└── "
			continuation = "    "
		} else {
			branch = "├── "
			continuation = "│   "
		}
	}

	// Get status indicator
	var statusIcon string
	switch issue.Status {
	case parser.StatusOpen:
		statusIcon = "●"
	case parser.StatusBlocked:
		statusIcon = "○"
	case parser.StatusInProgress:
		statusIcon = "◆"
	default:
		statusIcon = "·"
	}

	// Format issue line
	priorityColor := formatting.GetPriorityColor(issue.Priority)
	statusColor := formatting.GetStatusColor(issue.Status)
	text := fmt.Sprintf("%s%s[%s]%s[-] [%s]%s[-] [P%d] %s",
		prefix, branch, statusColor, statusIcon, priorityColor, issue.ID, issue.Priority, issue.Title)

	// Add labels if present
	if len(issue.Labels) > 0 {
		text += " [gray]"
		for i, label := range issue.Labels {
			if i > 0 {
				text += " "
			}
			text += "#" + label
		}
		text += "[-]"
	}

	issueList.AddItem(text, "", 0, nil)
	indexToIssue[*currentIndex] = issue
	*currentIndex++

	// Render children
	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		newPrefix := prefix + continuation
		renderTreeNode(issueList, child, newPrefix, isLastChild, currentIndex, indexToIssue)
	}
}

// UpdatePanelFocus updates the visual indicators for which panel is focused
func UpdatePanelFocus(
	issueList *tview.List,
	detailPanel *tview.TextView,
	detailPanelFocused bool,
) {
	if detailPanelFocused {
		issueList.SetBorderColor(tview.Styles.PrimaryTextColor)
		issueList.SetTitle("Issues")
		detailPanel.SetBorderColor(tview.Styles.SecondaryTextColor) // Yellow
		detailPanel.SetTitle("Details [ESC to return]")
	} else {
		issueList.SetBorderColor(tview.Styles.SecondaryTextColor) // Yellow
		issueList.SetTitle("Issues [Navigate]")
		detailPanel.SetBorderColor(tview.Styles.PrimaryTextColor)
		detailPanel.SetTitle("Details")
	}
}
