package formatting

import "github.com/andy/beads-tui/internal/parser"

// GetPriorityColor returns a tview color code for the given priority level
func GetPriorityColor(priority int) string {
	switch priority {
	case 0:
		return "red" // Critical - bright red
	case 1:
		return "orangered" // High - orange-red for urgency
	case 2:
		return "lightskyblue" // Normal - calm blue
	case 3:
		return "darkgray" // Low - subdued gray
	case 4:
		return "gray" // Lowest - very subdued
	default:
		return "white"
	}
}

// GetStatusColor returns a tview color code for the given status
func GetStatusColor(status parser.Status) string {
	switch status {
	case parser.StatusOpen:
		return "limegreen" // Open - bright green for ready work
	case parser.StatusInProgress:
		return "deepskyblue" // In Progress - vibrant blue for active work
	case parser.StatusBlocked:
		return "gold" // Blocked - gold/yellow for warning
	case parser.StatusClosed:
		return "darkgray" // Closed - muted gray
	default:
		return "white"
	}
}

// GetTypeIcon returns an emoji icon for the given issue type
func GetTypeIcon(issueType parser.IssueType) string {
	switch issueType {
	case parser.TypeBug:
		return "🐛"
	case parser.TypeFeature:
		return "✨"
	case parser.TypeTask:
		return "📋"
	case parser.TypeEpic:
		return "🎯"
	case parser.TypeChore:
		return "🔧"
	default:
		return "•"
	}
}

// GetDependencyColor returns a tview color code for the given dependency type
func GetDependencyColor(depType parser.DependencyType) string {
	switch depType {
	case parser.DepBlocks:
		return "red"
	case parser.DepRelated:
		return "blue"
	case parser.DepParentChild:
		return "green"
	case parser.DepDiscoveredFrom:
		return "yellow"
	default:
		return "white"
	}
}
