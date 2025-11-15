package formatting

import (
	"fmt"

	"github.com/andy/beads-tui/internal/parser"
)

// FormatIssueDetails formats full issue metadata for display in the detail panel
func FormatIssueDetails(issue *parser.Issue) string {
	var result string

	// Header
	priorityColor := GetPriorityColor(issue.Priority)
	statusColor := GetStatusColor(issue.Status)
	typeIcon := GetTypeIcon(issue.IssueType)

	mutedColor := GetMutedColor()
	accentColor := GetAccentColor()
	emphasisColor := GetEmphasisColor()

	result += fmt.Sprintf("[::b]%s %s[-::-]\n", typeIcon, issue.Title)
	result += fmt.Sprintf("[%s]ID:[-] %s [%s](click to copy)[-]  ", mutedColor, issue.ID, accentColor)
	result += fmt.Sprintf("[%s]P%d[-]  ", priorityColor, issue.Priority)
	result += fmt.Sprintf("[%s]%s[-]\n\n", statusColor, issue.Status)

	// Description
	if issue.Description != "" {
		result += fmt.Sprintf("[%s::b]Description:[-::-]\n", emphasisColor)
		result += issue.Description + "\n\n"
	}

	// Design notes
	if issue.Design != "" {
		result += fmt.Sprintf("[%s::b]Design:[-::-]\n", emphasisColor)
		result += issue.Design + "\n\n"
	}

	// Acceptance criteria
	if issue.AcceptanceCriteria != "" {
		result += fmt.Sprintf("[%s::b]Acceptance Criteria:[-::-]\n", emphasisColor)
		result += issue.AcceptanceCriteria + "\n\n"
	}

	// Notes
	if issue.Notes != "" {
		result += fmt.Sprintf("[%s::b]Notes:[-::-]\n", emphasisColor)
		result += issue.Notes + "\n\n"
	}

	// Dependencies
	if len(issue.Dependencies) > 0 {
		result += fmt.Sprintf("[%s::b]Dependencies:[-::-]\n", emphasisColor)
		for _, dep := range issue.Dependencies {
			result += fmt.Sprintf("  • [%s]%s[-] %s\n",
				GetDependencyColor(dep.Type), dep.Type, dep.DependsOnID)
		}
		result += "\n"
	}

	// Labels
	if len(issue.Labels) > 0 {
		result += fmt.Sprintf("[%s::b]Labels:[-::-] ", emphasisColor)
		for i, label := range issue.Labels {
			if i > 0 {
				result += ", "
			}
			result += fmt.Sprintf("[%s]%s[-]", accentColor, label)
		}
		result += "\n\n"
	}

	// Metadata
	result += fmt.Sprintf("[%s::b]Metadata:[-::-]\n", emphasisColor)
	result += fmt.Sprintf("  Created: %s\n", issue.CreatedAt.Format("2006-01-02 15:04"))
	result += fmt.Sprintf("  Updated: %s\n", issue.UpdatedAt.Format("2006-01-02 15:04"))

	if issue.ClosedAt != nil {
		result += fmt.Sprintf("  Closed: %s\n", issue.ClosedAt.Format("2006-01-02 15:04"))
	}

	if issue.Assignee != "" {
		result += fmt.Sprintf("  Assignee: %s\n", issue.Assignee)
	}

	if issue.EstimatedMinutes != nil {
		hours := *issue.EstimatedMinutes / 60
		mins := *issue.EstimatedMinutes % 60
		result += fmt.Sprintf("  Estimated: %dh %dm\n", hours, mins)
	}

	if issue.ExternalRef != nil {
		result += fmt.Sprintf("  External Ref: %s\n", *issue.ExternalRef)
	}

	// Comments
	if len(issue.Comments) > 0 {
		result += fmt.Sprintf("\n[%s::b]Comments:[-::-]\n", emphasisColor)
		for _, comment := range issue.Comments {
			result += fmt.Sprintf("  [%s]%s[-] (%s):\n", accentColor, comment.Author, comment.CreatedAt.Format("2006-01-02 15:04"))
			result += fmt.Sprintf("    %s\n", comment.Text)
		}
	}

	return result
}
