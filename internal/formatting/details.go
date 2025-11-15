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

	result += fmt.Sprintf("[::b]%s %s[-::-]\n", typeIcon, issue.Title)
	result += fmt.Sprintf("[gray]ID:[-] %s [blue](click to copy)[-]  ", issue.ID)
	result += fmt.Sprintf("[%s]P%d[-]  ", priorityColor, issue.Priority)
	result += fmt.Sprintf("[%s]%s[-]\n\n", statusColor, issue.Status)

	// Description
	if issue.Description != "" {
		result += "[yellow::b]Description:[-::-]\n"
		result += issue.Description + "\n\n"
	}

	// Design notes
	if issue.Design != "" {
		result += "[yellow::b]Design:[-::-]\n"
		result += issue.Design + "\n\n"
	}

	// Acceptance criteria
	if issue.AcceptanceCriteria != "" {
		result += "[yellow::b]Acceptance Criteria:[-::-]\n"
		result += issue.AcceptanceCriteria + "\n\n"
	}

	// Notes
	if issue.Notes != "" {
		result += "[yellow::b]Notes:[-::-]\n"
		result += issue.Notes + "\n\n"
	}

	// Dependencies
	if len(issue.Dependencies) > 0 {
		result += "[yellow::b]Dependencies:[-::-]\n"
		for _, dep := range issue.Dependencies {
			result += fmt.Sprintf("  • [%s]%s[-] %s\n",
				GetDependencyColor(dep.Type), dep.Type, dep.DependsOnID)
		}
		result += "\n"
	}

	// Labels
	if len(issue.Labels) > 0 {
		result += "[yellow::b]Labels:[-::-] "
		for i, label := range issue.Labels {
			if i > 0 {
				result += ", "
			}
			result += fmt.Sprintf("[cyan]%s[-]", label)
		}
		result += "\n\n"
	}

	// Metadata
	result += "[yellow::b]Metadata:[-::-]\n"
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
		result += "\n[yellow::b]Comments:[-::-]\n"
		for _, comment := range issue.Comments {
			result += fmt.Sprintf("  [cyan]%s[-] (%s):\n", comment.Author, comment.CreatedAt.Format("2006-01-02 15:04"))
			result += fmt.Sprintf("    %s\n", comment.Text)
		}
	}

	return result
}
