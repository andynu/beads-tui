package formatting

import (
	"github.com/andy/beads-tui/internal/parser"
	"github.com/andy/beads-tui/internal/theme"
)

// GetPriorityColor returns a tview color code for the given priority level
func GetPriorityColor(priority int) string {
	colors := theme.Current().PriorityColors()
	if priority >= 0 && priority < len(colors) {
		return colors[priority]
	}
	return "white"
}

// GetStatusColor returns a tview color code for the given status
func GetStatusColor(status parser.Status) string {
	t := theme.Current()
	switch status {
	case parser.StatusOpen:
		return t.StatusOpen()
	case parser.StatusInProgress:
		return t.StatusInProgress()
	case parser.StatusBlocked:
		return t.StatusBlocked()
	case parser.StatusClosed:
		return t.StatusClosed()
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
	t := theme.Current()
	switch depType {
	case parser.DepBlocks:
		return t.DepBlocks()
	case parser.DepRelated:
		return t.DepRelated()
	case parser.DepParentChild:
		return t.DepParentChild()
	case parser.DepDiscoveredFrom:
		return t.DepDiscoveredFrom()
	default:
		return "white"
	}
}

// GetSuccessColor returns the theme's success color
func GetSuccessColor() string {
	return theme.Current().Success()
}

// GetErrorColor returns the theme's error color
func GetErrorColor() string {
	return theme.Current().Error()
}

// GetWarningColor returns the theme's warning color
func GetWarningColor() string {
	return theme.Current().Warning()
}

// GetInfoColor returns the theme's info color
func GetInfoColor() string {
	return theme.Current().Info()
}

// GetMutedColor returns the theme's muted color
func GetMutedColor() string {
	return theme.Current().Muted()
}

// GetEmphasisColor returns the theme's emphasis color
func GetEmphasisColor() string {
	return theme.Current().Emphasis()
}

// GetAccentColor returns the theme's accent color
func GetAccentColor() string {
	return theme.Current().Accent()
}
