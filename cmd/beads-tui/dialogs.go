package main

import (
	"github.com/andy/beads-tui/internal/parser"
	"github.com/andy/beads-tui/internal/state"
	"github.com/rivo/tview"
)

// DialogHelpers holds references to UI components needed by dialog functions
//
// This struct is shared across all dialog implementations in separate files:
// - dialog_comment.go: ShowCommentDialog
// - dialog_rename.go: ShowRenameDialog
// - dialog_filter.go: ShowQuickFilter
// - dialog_stats.go: ShowStatsOverlay
// - dialog_help.go: ShowHelpScreen
// - dialog_dependencies.go: ShowDependencyDialog
// - dialog_labels.go: ShowLabelDialog
// - dialog_close.go: ShowCloseIssueDialog, ShowReopenIssueDialog
// - dialog_edit.go: ShowEditForm
// - dialog_create.go: ShowCreateIssueDialog
type DialogHelpers struct {
	App             *tview.Application
	Pages           *tview.Pages
	IssueList       *tview.List
	IndexToIssue    *map[int]*parser.Issue
	StatusBar       *tview.TextView
	AppState        *state.State
	RefreshIssues   func(...string)
	ScheduleRefresh func(string)
}
