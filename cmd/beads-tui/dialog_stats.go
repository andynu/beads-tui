package main

import (
	"fmt"
	"strings"

	"github.com/andy/beads-tui/internal/formatting"
	"github.com/andy/beads-tui/internal/parser"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ShowStatsOverlay displays a statistics dashboard
func (h *DialogHelpers) ShowStatsOverlay() {
	allIssues := h.AppState.GetAllIssues()

	// Calculate statistics
	stats := struct {
		total        int
		byStatus     map[parser.Status]int
		byPriority   map[int]int
		byType       map[parser.IssueType]int
		totalDeps    int
		avgDepsPerIssue float64
	}{
		byStatus:   make(map[parser.Status]int),
		byPriority: make(map[int]int),
		byType:     make(map[parser.IssueType]int),
	}

	stats.total = len(allIssues)
	totalDeps := 0

	for _, issue := range allIssues {
		stats.byStatus[issue.Status]++
		stats.byPriority[issue.Priority]++
		stats.byType[issue.IssueType]++
		totalDeps += len(issue.Dependencies)
	}

	stats.totalDeps = totalDeps
	if stats.total > 0 {
		stats.avgDepsPerIssue = float64(totalDeps) / float64(stats.total)
	}

	// Build stats text
	var sb strings.Builder
	emphasisColor := formatting.GetEmphasisColor()
	accentColor := formatting.GetAccentColor()
	mutedColor := formatting.GetMutedColor()
	priorityColors := [5]string{
		formatting.GetPriorityColor(0),
		formatting.GetPriorityColor(1),
		formatting.GetPriorityColor(2),
		formatting.GetPriorityColor(3),
		formatting.GetPriorityColor(4),
	}

	sb.WriteString(fmt.Sprintf("[%s::b]Issue Statistics Dashboard[-::-]\n\n", emphasisColor))

	// Overall stats
	sb.WriteString(fmt.Sprintf("[%s::b]Total Issues:[-::-] %d\n\n", accentColor, stats.total))

	// By Status
	sb.WriteString(fmt.Sprintf("[%s::b]By Status:[-::-]\n", accentColor))
	sb.WriteString(fmt.Sprintf("  [%s]Ready[-]:        %3d  (%.1f%%)\n",
		formatting.GetStatusColor(parser.StatusOpen),
		stats.byStatus[parser.StatusOpen],
		float64(stats.byStatus[parser.StatusOpen])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  [%s]In Progress[-]: %3d  (%.1f%%)\n",
		formatting.GetStatusColor(parser.StatusInProgress),
		stats.byStatus[parser.StatusInProgress],
		float64(stats.byStatus[parser.StatusInProgress])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  [%s]Blocked[-]:     %3d  (%.1f%%)\n",
		formatting.GetStatusColor(parser.StatusBlocked),
		stats.byStatus[parser.StatusBlocked],
		float64(stats.byStatus[parser.StatusBlocked])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  [%s]Closed[-]:      %3d  (%.1f%%)\n\n",
		formatting.GetStatusColor(parser.StatusClosed),
		stats.byStatus[parser.StatusClosed],
		float64(stats.byStatus[parser.StatusClosed])/float64(stats.total)*100))

	// By Priority
	sb.WriteString(fmt.Sprintf("[%s::b]By Priority:[-::-]\n", accentColor))
	sb.WriteString(fmt.Sprintf("  [%s]P0 (Critical)[-]: %3d  (%.1f%%)\n",
		priorityColors[0],
		stats.byPriority[0],
		float64(stats.byPriority[0])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  [%s]P1 (High)[-]:     %3d  (%.1f%%)\n",
		priorityColors[1],
		stats.byPriority[1],
		float64(stats.byPriority[1])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  [%s]P2 (Normal)[-]:   %3d  (%.1f%%)\n",
		priorityColors[2],
		stats.byPriority[2],
		float64(stats.byPriority[2])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  [%s]P3 (Low)[-]:      %3d  (%.1f%%)\n",
		priorityColors[3],
		stats.byPriority[3],
		float64(stats.byPriority[3])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  [%s]P4 (Lowest)[-]:   %3d  (%.1f%%)\n\n",
		priorityColors[4],
		stats.byPriority[4],
		float64(stats.byPriority[4])/float64(stats.total)*100))

	// By Type
	sb.WriteString(fmt.Sprintf("[%s::b]By Type:[-::-]\n", accentColor))
	sb.WriteString(fmt.Sprintf("  Bug:      %3d  (%.1f%%)\n",
		stats.byType[parser.TypeBug],
		float64(stats.byType[parser.TypeBug])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  Feature:  %3d  (%.1f%%)\n",
		stats.byType[parser.TypeFeature],
		float64(stats.byType[parser.TypeFeature])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  Task:     %3d  (%.1f%%)\n",
		stats.byType[parser.TypeTask],
		float64(stats.byType[parser.TypeTask])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  Epic:     %3d  (%.1f%%)\n",
		stats.byType[parser.TypeEpic],
		float64(stats.byType[parser.TypeEpic])/float64(stats.total)*100))
	sb.WriteString(fmt.Sprintf("  Chore:    %3d  (%.1f%%)\n\n",
		stats.byType[parser.TypeChore],
		float64(stats.byType[parser.TypeChore])/float64(stats.total)*100))

	// Dependencies
	sb.WriteString(fmt.Sprintf("[%s::b]Dependencies:[-::-]\n", accentColor))
	sb.WriteString(fmt.Sprintf("  Total:           %d\n", stats.totalDeps))
	sb.WriteString(fmt.Sprintf("  Avg per issue:   %.2f\n", stats.avgDepsPerIssue))

	sb.WriteString(fmt.Sprintf("\n[%s]━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━[-]\n", mutedColor))
	sb.WriteString(fmt.Sprintf("[%s]Press ESC or S to close[-]", emphasisColor))

	// Create stats text view
	statsTextView := tview.NewTextView().
		SetDynamicColors(true).
		SetText(sb.String()).
		SetTextAlign(tview.AlignLeft)
	statsTextView.SetBorder(true).
		SetTitle(" Statistics Dashboard ").
		SetTitleAlign(tview.AlignCenter)

	// Create modal (centered, slightly smaller than help)
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(statsTextView, 0, 2, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)

	// Add input capture to close on ESC or S
	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || (event.Key() == tcell.KeyRune && (event.Rune() == 'S' || event.Rune() == 's')) {
			h.Pages.RemovePage("stats")
			h.App.SetFocus(h.IssueList)
			return nil
		}
		return event
	})

	// Show modal
	h.Pages.AddPage("stats", modal, true, true)
	h.App.SetFocus(modal)
}
