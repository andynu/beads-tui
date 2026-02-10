package ui

import (
	"fmt"

	"github.com/andy/beads-tui/internal/formatting"
	"github.com/andy/beads-tui/internal/parser"
	"github.com/andy/beads-tui/internal/state"
	"github.com/rivo/tview"
)

// PopulateIssueList clears and rebuilds the issue list from state
// Updates the provided indexToIssue map in place to avoid stale pointer issues
func PopulateIssueList(
	issueList *tview.List,
	appState *state.State,
	showClosedIssues bool,
	showPrefix bool,
	indexToIssue map[int]*parser.Issue,
) {
	// Clear and rebuild issue list
	issueList.Clear()

	// Clear the map in place (don't create a new one)
	for k := range indexToIssue {
		delete(indexToIssue, k)
	}
	currentIndex := 0

	// Show filter indicator when filters are active
	if appState.HasActiveFilters() {
		warningColor := formatting.GetWarningColor()
		emphasisColor := formatting.GetEmphasisColor()
		issueList.AddItem(fmt.Sprintf("[%s::b]⊘ FILTERED[-::-] [%s]%s[-] — press f to modify",
			warningColor, emphasisColor, appState.GetActiveFilters()), "", 0, nil)
		currentIndex++
	}

	// Check view mode
	if appState.GetViewMode() == state.ViewTree {
		// Tree view
		accentColor := formatting.GetAccentColor()
		issueList.AddItem(fmt.Sprintf("[%s::b]DEPENDENCY TREE[-::-]", accentColor), "", 0, nil)
		currentIndex++

		treeNodes := appState.GetTreeNodes()
		for i, node := range treeNodes {
			isLast := i == len(treeNodes)-1
			renderTreeNode(issueList, appState, node, "", isLast, showPrefix, &currentIndex, indexToIssue)
		}
	} else {
		// List view (original behavior)
		// Add in-progress issues first (most important)
		inProgressIssues := appState.GetInProgressIssues()
		if len(inProgressIssues) > 0 {
			inProgressColor := formatting.GetStatusColor(parser.StatusInProgress)
			issueList.AddItem(fmt.Sprintf("[%s::b]⬤ IN PROGRESS (%d)[-::-]", inProgressColor, len(inProgressIssues)), "", 0, nil)
			currentIndex++

			for _, issue := range inProgressIssues {
				text := formatIssueListItem(issue, "◆", showPrefix)
				issueList.AddItem(text, "", 0, nil)
				indexToIssue[currentIndex] = issue
				currentIndex++
			}
		}

		// Add ready issues
		readyIssues := appState.GetReadyIssues()
		if len(readyIssues) > 0 {
			openColor := formatting.GetStatusColor(parser.StatusOpen)
			issueList.AddItem(fmt.Sprintf("\n[%s::b]⬤ READY (%d)[-::-]", openColor, len(readyIssues)), "", 0, nil)
			currentIndex++

			for _, issue := range readyIssues {
				text := formatIssueListItem(issue, "●", showPrefix)
				issueList.AddItem(text, "", 0, nil)
				indexToIssue[currentIndex] = issue
				currentIndex++
			}
		}

		// Add blocked issues
		blockedIssues := appState.GetBlockedIssues()
		if len(blockedIssues) > 0 {
			blockedColor := formatting.GetStatusColor(parser.StatusBlocked)
			issueList.AddItem(fmt.Sprintf("\n[%s::b]⬤ BLOCKED (%d)[-::-]", blockedColor, len(blockedIssues)), "", 0, nil)
			currentIndex++

			for _, issue := range blockedIssues {
				text := formatIssueListItem(issue, "○", showPrefix)
				issueList.AddItem(text, "", 0, nil)
				indexToIssue[currentIndex] = issue
				currentIndex++
			}
		}

		// Add closed issues (only if showClosedIssues is enabled)
		if showClosedIssues {
			closedIssues := appState.GetClosedIssues()
			if len(closedIssues) > 0 {
				closedColor := formatting.GetStatusColor(parser.StatusClosed)
				issueList.AddItem(fmt.Sprintf("\n[%s::b]⬤ CLOSED (%d)[-::-]", closedColor, len(closedIssues)), "", 0, nil)
				currentIndex++

				for _, issue := range closedIssues {
					text := formatIssueListItem(issue, "✓", showPrefix)
					issueList.AddItem(text, "", 0, nil)
					indexToIssue[currentIndex] = issue
					currentIndex++
				}
			}
		}
	}

	// Show helpful message when no issues are visible
	if len(indexToIssue) == 0 {
		mutedColor := formatting.GetMutedColor()
		emphasisColor := formatting.GetEmphasisColor()
		if appState.HasActiveFilters() {
			issueList.AddItem(fmt.Sprintf("\n  [%s]No issues match current filters[-]", mutedColor), "", 0, nil)
			issueList.AddItem(fmt.Sprintf("  [%s]Press 'f' to modify filters[-]", emphasisColor), "", 0, nil)
		} else {
			issueList.AddItem(fmt.Sprintf("\n  [%s]No issues found[-]", mutedColor), "", 0, nil)
			issueList.AddItem(fmt.Sprintf("  [%s]Press 'a' to create an issue[-]", emphasisColor), "", 0, nil)
		}
	}
}

// formatIssueListItem formats a single issue for the list view
func formatIssueListItem(issue *parser.Issue, statusIcon string, showPrefix bool) string {
	priorityColor := formatting.GetPriorityColor(issue.Priority)
	typeIcon := formatting.GetTypeIcon(issue.IssueType)
	displayID := formatting.FormatIssueID(issue.ID, showPrefix)
	text := fmt.Sprintf("  [%s]%s[-] %s %s [P%d] %s",
		priorityColor, statusIcon, typeIcon, displayID, issue.Priority, issue.Title)

	// Add labels if present
	if len(issue.Labels) > 0 {
		mutedColor := formatting.GetMutedColor()
		text += fmt.Sprintf(" [%s]", mutedColor)
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
	appState *state.State,
	node *state.TreeNode,
	prefix string,
	isLast bool,
	showPrefix bool,
	currentIndex *int,
	indexToIssue map[int]*parser.Issue,
) {
	issue := node.Issue
	hasChildren := len(node.Children) > 0
	isCollapsed := appState.IsCollapsed(issue.ID)

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

	// Get status indicator - use effective blocking status for consistent display
	// This ensures issues blocked by dependencies show as blocked even if their
	// explicit status is "open"
	var statusIcon string
	var statusColor string
	switch {
	case issue.Status == parser.StatusClosed:
		statusIcon = "✓"
		statusColor = formatting.GetStatusColor(parser.StatusClosed)
	case issue.Status == parser.StatusInProgress:
		statusIcon = "◆"
		statusColor = formatting.GetStatusColor(parser.StatusInProgress)
	case appState.IsEffectivelyBlocked(issue.ID):
		// Blocked by explicit status OR by dependency
		statusIcon = "○"
		statusColor = formatting.GetStatusColor(parser.StatusBlocked)
	default:
		// Ready (open and not blocked)
		statusIcon = "●"
		statusColor = formatting.GetStatusColor(parser.StatusOpen)
	}

	// Add collapse indicator for parent nodes
	collapseIndicator := ""
	if hasChildren {
		if isCollapsed {
			collapseIndicator = "▶ " // Collapsed - can expand
		} else {
			collapseIndicator = "▼ " // Expanded - can collapse
		}
	} else {
		collapseIndicator = "  " // Leaf node - no indicator (maintain alignment)
	}

	// Format issue line
	priorityColor := formatting.GetPriorityColor(issue.Priority)
	typeIcon := formatting.GetTypeIcon(issue.IssueType)
	displayID := formatting.FormatIssueID(issue.ID, showPrefix)
	text := fmt.Sprintf("%s%s%s[%s]%s[-] %s [%s]%s[-] [P%d] %s",
		prefix, branch, collapseIndicator, statusColor, statusIcon, typeIcon, priorityColor, displayID, issue.Priority, issue.Title)

	// Add child count for collapsed nodes
	if hasChildren && isCollapsed {
		mutedColor := formatting.GetMutedColor()
		text += fmt.Sprintf(" [%s](%d children)[-]", mutedColor, len(node.Children))
	}

	// Add labels if present
	if len(issue.Labels) > 0 {
		mutedColor := formatting.GetMutedColor()
		text += fmt.Sprintf(" [%s]", mutedColor)
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

	// Render children only if not collapsed
	if !isCollapsed {
		for i, child := range node.Children {
			isLastChild := i == len(node.Children)-1
			newPrefix := prefix + continuation
			renderTreeNode(issueList, appState, child, newPrefix, isLastChild, showPrefix, currentIndex, indexToIssue)
		}
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
