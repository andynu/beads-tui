package formatting

import (
	"fmt"

	"github.com/andy/beads-tui/internal/state"
)

// GetStatusBarText generates the status bar text with view mode, issue count, and filters
func GetStatusBarText(
	beadsDir string,
	appState *state.State,
	viewMode state.ViewMode,
	mouseEnabled bool,
	detailPanelFocused bool,
	showClosedIssues bool,
) string {
	viewModeStr := "List"
	if viewMode == state.ViewTree {
		viewModeStr = "Tree"
	}

	mouseStr := "OFF"
	if mouseEnabled {
		mouseStr = "ON"
	}

	focusStr := "List"
	if detailPanelFocused {
		focusStr = "Details"
	}

	// Count visible issues after filtering
	visibleCount := len(appState.GetReadyIssues()) + len(appState.GetBlockedIssues()) + len(appState.GetInProgressIssues())
	if showClosedIssues {
		visibleCount += len(appState.GetClosedIssues())
	}

	filterText := ""
	if appState.HasActiveFilters() {
		filterText = fmt.Sprintf(" [Filters: %s]", appState.GetActiveFilters())
	}

	closedText := ""
	if showClosedIssues {
		closedText = " [Showing Closed]"
	}

	emphasisColor := GetEmphasisColor()
	return fmt.Sprintf("[%s]Beads TUI[-] - %s (%d issues)%s%s [SQLite] [%s View] [Mouse: %s] [Focus: %s] [Press ? for help, f for quick filter]",
		emphasisColor, beadsDir, visibleCount, filterText, closedText, viewModeStr, mouseStr, focusStr)
}
