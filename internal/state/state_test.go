package state

import (
	"testing"
	"time"

	"github.com/andy/beads-tui/internal/parser"
)

func TestStateLoadIssues(t *testing.T) {
	state := New()

	issues := []*parser.Issue{
		{
			ID:          "test-1",
			Title:       "Ready Issue",
			Status:      parser.StatusOpen,
			Priority:    1,
			IssueType:   parser.TypeTask,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Dependencies: nil,
		},
		{
			ID:        "test-2",
			Title:     "In Progress Issue",
			Status:    parser.StatusInProgress,
			Priority:  0,
			IssueType: parser.TypeBug,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "test-3",
			Title:     "Closed Issue",
			Status:    parser.StatusClosed,
			Priority:  2,
			IssueType: parser.TypeFeature,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	state.LoadIssues(issues)

	// Verify all issues are loaded
	if len(state.GetAllIssues()) != 3 {
		t.Errorf("Expected 3 issues, got %d", len(state.GetAllIssues()))
	}

	// Verify categorization
	if len(state.GetReadyIssues()) != 1 {
		t.Errorf("Expected 1 ready issue, got %d", len(state.GetReadyIssues()))
	}
	if len(state.GetInProgressIssues()) != 1 {
		t.Errorf("Expected 1 in-progress issue, got %d", len(state.GetInProgressIssues()))
	}
	if len(state.GetClosedIssues()) != 1 {
		t.Errorf("Expected 1 closed issue, got %d", len(state.GetClosedIssues()))
	}
}

func TestStateBlockedIssues(t *testing.T) {
	state := New()

	now := time.Now()
	issues := []*parser.Issue{
		{
			ID:        "test-1",
			Title:     "Blocking Issue (open)",
			Status:    parser.StatusOpen,
			Priority:  1,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "test-2",
			Title:     "Blocked Issue",
			Status:    parser.StatusOpen,
			Priority:  0,
			IssueType: parser.TypeBug,
			CreatedAt: now,
			UpdatedAt: now,
			Dependencies: []*parser.Dependency{
				{
					IssueID:     "test-2",
					DependsOnID: "test-1",
					Type:        parser.DepBlocks,
					CreatedAt:   now,
					CreatedBy:   "test",
				},
			},
		},
		{
			ID:        "test-3",
			Title:     "Another Ready Issue",
			Status:    parser.StatusOpen,
			Priority:  2,
			IssueType: parser.TypeFeature,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	state.LoadIssues(issues)

	// test-2 should be blocked because it depends on test-1 which is open
	// test-1 and test-3 should be ready
	readyIssues := state.GetReadyIssues()
	blockedIssues := state.GetBlockedIssues()

	if len(blockedIssues) != 1 {
		t.Errorf("Expected 1 blocked issue, got %d", len(blockedIssues))
		for _, issue := range blockedIssues {
			t.Logf("  Blocked: %s - %s", issue.ID, issue.Title)
		}
	}

	if blockedIssues[0].ID != "test-2" {
		t.Errorf("Expected test-2 to be blocked, got %s", blockedIssues[0].ID)
	}

	if len(readyIssues) != 2 {
		t.Errorf("Expected 2 ready issues, got %d", len(readyIssues))
		for _, issue := range readyIssues {
			t.Logf("  Ready: %s - %s", issue.ID, issue.Title)
		}
	}
}

func TestStateBlockedByClosedIssue(t *testing.T) {
	state := New()

	now := time.Now()
	closedAt := now.Add(-1 * time.Hour)

	issues := []*parser.Issue{
		{
			ID:        "test-1",
			Title:     "Blocking Issue (closed)",
			Status:    parser.StatusClosed,
			Priority:  1,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
			ClosedAt:  &closedAt,
		},
		{
			ID:        "test-2",
			Title:     "No Longer Blocked Issue",
			Status:    parser.StatusOpen,
			Priority:  0,
			IssueType: parser.TypeBug,
			CreatedAt: now,
			UpdatedAt: now,
			Dependencies: []*parser.Dependency{
				{
					IssueID:     "test-2",
					DependsOnID: "test-1",
					Type:        parser.DepBlocks,
					CreatedAt:   now,
					CreatedBy:   "test",
				},
			},
		},
	}

	state.LoadIssues(issues)

	// test-2 should be ready because test-1 is closed
	readyIssues := state.GetReadyIssues()
	blockedIssues := state.GetBlockedIssues()

	if len(blockedIssues) != 0 {
		t.Errorf("Expected 0 blocked issues, got %d", len(blockedIssues))
	}

	if len(readyIssues) != 1 {
		t.Errorf("Expected 1 ready issue, got %d", len(readyIssues))
	}

	if readyIssues[0].ID != "test-2" {
		t.Errorf("Expected test-2 to be ready, got %s", readyIssues[0].ID)
	}
}

func TestGetIssueByID(t *testing.T) {
	state := New()

	issues := []*parser.Issue{
		{
			ID:        "test-1",
			Title:     "Issue 1",
			Status:    parser.StatusOpen,
			Priority:  1,
			IssueType: parser.TypeTask,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "test-2",
			Title:     "Issue 2",
			Status:    parser.StatusOpen,
			Priority:  2,
			IssueType: parser.TypeTask,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	state.LoadIssues(issues)

	// Test getting existing issue
	issue := state.GetIssueByID("test-1")
	if issue == nil {
		t.Fatal("Expected to find test-1")
	}
	if issue.Title != "Issue 1" {
		t.Errorf("Expected title 'Issue 1', got '%s'", issue.Title)
	}

	// Test getting non-existent issue
	issue = state.GetIssueByID("test-999")
	if issue != nil {
		t.Error("Expected nil for non-existent issue")
	}
}

func TestTreeViewMode(t *testing.T) {
	state := New()

	// Initially should be in list view
	if state.GetViewMode() != ViewList {
		t.Errorf("Expected ViewList mode, got %v", state.GetViewMode())
	}

	// Toggle to tree view
	mode := state.ToggleViewMode()
	if mode != ViewTree {
		t.Errorf("Expected ViewTree after toggle, got %v", mode)
	}
	if state.GetViewMode() != ViewTree {
		t.Errorf("Expected ViewTree mode, got %v", state.GetViewMode())
	}

	// Toggle back to list view
	mode = state.ToggleViewMode()
	if mode != ViewList {
		t.Errorf("Expected ViewList after second toggle, got %v", mode)
	}
}

func TestBuildDependencyTree(t *testing.T) {
	state := New()
	now := time.Now()

	// Create a simple parent-child tree:
	// parent (test-1)
	//   ├── child1 (test-2)
	//   └── child2 (test-3)
	//       └── grandchild (test-4)
	issues := []*parser.Issue{
		{
			ID:        "test-1",
			Title:     "Parent Issue",
			Status:    parser.StatusOpen,
			Priority:  1,
			IssueType: parser.TypeEpic,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "test-2",
			Title:     "Child Issue 1",
			Status:    parser.StatusOpen,
			Priority:  1,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
			Dependencies: []*parser.Dependency{
				{
					IssueID:     "test-2",
					DependsOnID: "test-1",
					Type:        parser.DepParentChild,
					CreatedAt:   now,
					CreatedBy:   "test",
				},
			},
		},
		{
			ID:        "test-3",
			Title:     "Child Issue 2",
			Status:    parser.StatusOpen,
			Priority:  1,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
			Dependencies: []*parser.Dependency{
				{
					IssueID:     "test-3",
					DependsOnID: "test-1",
					Type:        parser.DepParentChild,
					CreatedAt:   now,
					CreatedBy:   "test",
				},
			},
		},
		{
			ID:        "test-4",
			Title:     "Grandchild Issue",
			Status:    parser.StatusOpen,
			Priority:  2,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
			Dependencies: []*parser.Dependency{
				{
					IssueID:     "test-4",
					DependsOnID: "test-3",
					Type:        parser.DepParentChild,
					CreatedAt:   now,
					CreatedBy:   "test",
				},
			},
		},
	}

	state.LoadIssues(issues)
	state.SetViewMode(ViewTree)

	// Should have 1 root node (test-1)
	treeNodes := state.GetTreeNodes()
	if len(treeNodes) != 1 {
		t.Errorf("Expected 1 root node, got %d", len(treeNodes))
	}

	if treeNodes[0].Issue.ID != "test-1" {
		t.Errorf("Expected root to be test-1, got %s", treeNodes[0].Issue.ID)
	}

	// Root should have 2 children
	if len(treeNodes[0].Children) != 2 {
		t.Errorf("Expected root to have 2 children, got %d", len(treeNodes[0].Children))
	}

	// Find test-3 among children
	var test3Node *TreeNode
	for _, child := range treeNodes[0].Children {
		if child.Issue.ID == "test-3" {
			test3Node = child
			break
		}
	}

	if test3Node == nil {
		t.Fatal("Could not find test-3 in children")
	}

	// test-3 should have 1 child (test-4)
	if len(test3Node.Children) != 1 {
		t.Errorf("Expected test-3 to have 1 child, got %d", len(test3Node.Children))
	}

	if test3Node.Children[0].Issue.ID != "test-4" {
		t.Errorf("Expected grandchild to be test-4, got %s", test3Node.Children[0].Issue.ID)
	}
}

func TestTreeViewWithBlockedIssues(t *testing.T) {
	state := New()
	now := time.Now()

	// Create a tree with blocking dependencies:
	// blocker (test-1)
	//   └── blocked (test-2) [blocked by test-1]
	issues := []*parser.Issue{
		{
			ID:        "test-1",
			Title:     "Blocker Issue",
			Status:    parser.StatusOpen,
			Priority:  1,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "test-2",
			Title:     "Blocked Issue",
			Status:    parser.StatusOpen,
			Priority:  1,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
			Dependencies: []*parser.Dependency{
				{
					IssueID:     "test-2",
					DependsOnID: "test-1",
					Type:        parser.DepBlocks,
					CreatedAt:   now,
					CreatedBy:   "test",
				},
			},
		},
	}

	state.LoadIssues(issues)
	state.SetViewMode(ViewTree)

	treeNodes := state.GetTreeNodes()
	if len(treeNodes) != 1 {
		t.Errorf("Expected 1 root node, got %d", len(treeNodes))
	}

	if treeNodes[0].Issue.ID != "test-1" {
		t.Errorf("Expected root to be test-1, got %s", treeNodes[0].Issue.ID)
	}

	// Blocker should have the blocked issue as child
	if len(treeNodes[0].Children) != 1 {
		t.Errorf("Expected blocker to have 1 child, got %d", len(treeNodes[0].Children))
	}

	if treeNodes[0].Children[0].Issue.ID != "test-2" {
		t.Errorf("Expected child to be test-2, got %s", treeNodes[0].Children[0].Issue.ID)
	}
}

func TestTreeViewExcludesClosedIssues(t *testing.T) {
	state := New()
	now := time.Now()
	closedAt := now.Add(-1 * time.Hour)

	issues := []*parser.Issue{
		{
			ID:        "test-1",
			Title:     "Open Issue",
			Status:    parser.StatusOpen,
			Priority:  1,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "test-2",
			Title:     "Closed Issue",
			Status:    parser.StatusClosed,
			Priority:  1,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
			ClosedAt:  &closedAt,
		},
	}

	state.LoadIssues(issues)
	state.SetViewMode(ViewTree)

	treeNodes := state.GetTreeNodes()
	// Only test-1 should appear (test-2 is closed)
	if len(treeNodes) != 1 {
		t.Errorf("Expected 1 root node (closed excluded), got %d", len(treeNodes))
	}

	if treeNodes[0].Issue.ID != "test-1" {
		t.Errorf("Expected root to be test-1, got %s", treeNodes[0].Issue.ID)
	}
}

func TestFilterByPriority(t *testing.T) {
	state := New()

	issues := []*parser.Issue{
		{ID: "test-1", Title: "P0 Issue", Status: parser.StatusOpen, Priority: 0, IssueType: parser.TypeBug, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "test-2", Title: "P1 Issue", Status: parser.StatusOpen, Priority: 1, IssueType: parser.TypeTask, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "test-3", Title: "P2 Issue", Status: parser.StatusOpen, Priority: 2, IssueType: parser.TypeFeature, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	state.LoadIssues(issues)

	// Initially no filters, should get all 3
	if len(state.GetReadyIssues()) != 3 {
		t.Errorf("Expected 3 ready issues with no filter, got %d", len(state.GetReadyIssues()))
	}

	// Filter for P0 only
	state.TogglePriorityFilter(0)
	readyIssues := state.GetReadyIssues()
	if len(readyIssues) != 1 {
		t.Errorf("Expected 1 issue with P0 filter, got %d", len(readyIssues))
	}
	if readyIssues[0].Priority != 0 {
		t.Errorf("Expected priority 0, got %d", readyIssues[0].Priority)
	}

	// Add P1 to filter
	state.TogglePriorityFilter(1)
	readyIssues = state.GetReadyIssues()
	if len(readyIssues) != 2 {
		t.Errorf("Expected 2 issues with P0,P1 filter, got %d", len(readyIssues))
	}

	// Toggle P0 off
	state.TogglePriorityFilter(0)
	readyIssues = state.GetReadyIssues()
	if len(readyIssues) != 1 {
		t.Errorf("Expected 1 issue with P1 filter, got %d", len(readyIssues))
	}
	if readyIssues[0].Priority != 1 {
		t.Errorf("Expected priority 1, got %d", readyIssues[0].Priority)
	}

	// Clear all filters
	state.ClearAllFilters()
	if len(state.GetReadyIssues()) != 3 {
		t.Errorf("Expected 3 ready issues after clearing filters, got %d", len(state.GetReadyIssues()))
	}
}

func TestFilterByType(t *testing.T) {
	state := New()

	issues := []*parser.Issue{
		{ID: "test-1", Title: "Bug", Status: parser.StatusOpen, Priority: 1, IssueType: parser.TypeBug, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "test-2", Title: "Feature", Status: parser.StatusOpen, Priority: 1, IssueType: parser.TypeFeature, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "test-3", Title: "Task", Status: parser.StatusOpen, Priority: 1, IssueType: parser.TypeTask, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	state.LoadIssues(issues)

	// Filter for bugs only
	state.ToggleTypeFilter(parser.TypeBug)
	readyIssues := state.GetReadyIssues()
	if len(readyIssues) != 1 {
		t.Errorf("Expected 1 bug, got %d", len(readyIssues))
	}
	if readyIssues[0].IssueType != parser.TypeBug {
		t.Errorf("Expected type bug, got %s", readyIssues[0].IssueType)
	}

	// Add features
	state.ToggleTypeFilter(parser.TypeFeature)
	readyIssues = state.GetReadyIssues()
	if len(readyIssues) != 2 {
		t.Errorf("Expected 2 issues (bug+feature), got %d", len(readyIssues))
	}
}

func TestFilterByStatus(t *testing.T) {
	state := New()

	issues := []*parser.Issue{
		{ID: "test-1", Title: "Open", Status: parser.StatusOpen, Priority: 1, IssueType: parser.TypeTask, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "test-2", Title: "In Progress", Status: parser.StatusInProgress, Priority: 1, IssueType: parser.TypeTask, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "test-3", Title: "Closed", Status: parser.StatusClosed, Priority: 1, IssueType: parser.TypeTask, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	state.LoadIssues(issues)

	// Filter for in_progress only
	state.ToggleStatusFilter(parser.StatusInProgress)
	inProgressIssues := state.GetInProgressIssues()
	if len(inProgressIssues) != 1 {
		t.Errorf("Expected 1 in_progress issue, got %d", len(inProgressIssues))
	}

	// Filter should exclude open and closed
	if len(state.GetReadyIssues()) != 0 {
		t.Errorf("Expected 0 ready issues with in_progress filter, got %d", len(state.GetReadyIssues()))
	}
	if len(state.GetClosedIssues()) != 0 {
		t.Errorf("Expected 0 closed issues with in_progress filter, got %d", len(state.GetClosedIssues()))
	}
}

func TestCombinedFilters(t *testing.T) {
	state := New()

	issues := []*parser.Issue{
		{ID: "test-1", Title: "P0 Bug", Status: parser.StatusOpen, Priority: 0, IssueType: parser.TypeBug, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "test-2", Title: "P0 Feature", Status: parser.StatusOpen, Priority: 0, IssueType: parser.TypeFeature, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "test-3", Title: "P1 Bug", Status: parser.StatusOpen, Priority: 1, IssueType: parser.TypeBug, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "test-4", Title: "P1 Feature", Status: parser.StatusInProgress, Priority: 1, IssueType: parser.TypeFeature, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	state.LoadIssues(issues)

	// Filter for P0 bugs
	state.TogglePriorityFilter(0)
	state.ToggleTypeFilter(parser.TypeBug)

	readyIssues := state.GetReadyIssues()
	if len(readyIssues) != 1 {
		t.Errorf("Expected 1 P0 bug, got %d", len(readyIssues))
	}
	if readyIssues[0].ID != "test-1" {
		t.Errorf("Expected test-1, got %s", readyIssues[0].ID)
	}

	// Add P1 to priority filter - should get both bugs
	state.TogglePriorityFilter(1)
	readyIssues = state.GetReadyIssues()
	if len(readyIssues) != 2 {
		t.Errorf("Expected 2 bugs (P0+P1), got %d", len(readyIssues))
	}

	// Add status filter for in_progress - should get nothing from ready (test-4 is in progress, but it's a feature)
	state.ToggleStatusFilter(parser.StatusInProgress)
	readyIssues = state.GetReadyIssues()
	if len(readyIssues) != 0 {
		t.Errorf("Expected 0 ready issues with in_progress status filter, got %d", len(readyIssues))
	}
}

func TestFilterHelpers(t *testing.T) {
	state := New()

	// Initially no filters
	if state.HasActiveFilters() {
		t.Error("Expected no active filters initially")
	}

	if state.IsPriorityFiltered(1) {
		t.Error("Expected priority 1 not filtered initially")
	}

	if state.IsTypeFiltered(parser.TypeBug) {
		t.Error("Expected bug type not filtered initially")
	}

	if state.IsStatusFiltered(parser.StatusOpen) {
		t.Error("Expected open status not filtered initially")
	}

	// Add some filters
	state.TogglePriorityFilter(1)
	state.ToggleTypeFilter(parser.TypeBug)
	state.ToggleStatusFilter(parser.StatusOpen)

	if !state.HasActiveFilters() {
		t.Error("Expected active filters after toggling")
	}

	if !state.IsPriorityFiltered(1) {
		t.Error("Expected priority 1 to be filtered")
	}

	if !state.IsTypeFiltered(parser.TypeBug) {
		t.Error("Expected bug type to be filtered")
	}

	if !state.IsStatusFiltered(parser.StatusOpen) {
		t.Error("Expected open status to be filtered")
	}

	// Clear all
	state.ClearAllFilters()

	if state.HasActiveFilters() {
		t.Error("Expected no active filters after clearing")
	}
}

func TestGetActiveFilters(t *testing.T) {
	state := New()

	// No filters
	filterStr := state.GetActiveFilters()
	if filterStr != "" {
		t.Errorf("Expected empty filter string, got '%s'", filterStr)
	}

	// Add priority filters
	state.TogglePriorityFilter(0)
	state.TogglePriorityFilter(1)
	filterStr = state.GetActiveFilters()
	if filterStr != "Priority: P0,P1" {
		t.Errorf("Expected 'Priority: P0,P1', got '%s'", filterStr)
	}

	// Add type filter
	state.ToggleTypeFilter(parser.TypeBug)
	filterStr = state.GetActiveFilters()
	if filterStr != "Priority: P0,P1 | Type: bug" {
		t.Errorf("Expected 'Priority: P0,P1 | Type: bug', got '%s'", filterStr)
	}

	// Add status filter
	state.ToggleStatusFilter(parser.StatusOpen)
	filterStr = state.GetActiveFilters()
	if filterStr != "Priority: P0,P1 | Type: bug | Status: open" {
		t.Errorf("Expected full filter string, got '%s'", filterStr)
	}
}

func TestSelectedIssue(t *testing.T) {
	state := New()

	issue := &parser.Issue{
		ID:        "test-1",
		Title:     "Test Issue",
		Status:    parser.StatusOpen,
		Priority:  1,
		IssueType: parser.TypeTask,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Initially no selection
	if state.GetSelectedIssue() != nil {
		t.Error("Expected no selected issue initially")
	}

	// Set selection
	state.SetSelectedIssue(issue)
	selected := state.GetSelectedIssue()
	if selected == nil {
		t.Fatal("Expected selected issue to be set")
	}
	if selected.ID != "test-1" {
		t.Errorf("Expected selected issue ID 'test-1', got '%s'", selected.ID)
	}

	// Clear selection
	state.SetSelectedIssue(nil)
	if state.GetSelectedIssue() != nil {
		t.Error("Expected no selected issue after clearing")
	}
}

func TestEpicsAlwaysAtRootLevel(t *testing.T) {
	state := New()
	now := time.Now()

	// Create an epic with a blocking dependency
	// Epic should still appear at root level even though it has an incoming dependency
	issues := []*parser.Issue{
		{
			ID:        "test-blocker",
			Title:     "Blocker Issue",
			Status:    parser.StatusOpen,
			Priority:  1,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "test-epic",
			Title:     "Epic with Blocker",
			Status:    parser.StatusOpen,
			Priority:  0,
			IssueType: parser.TypeEpic,
			CreatedAt: now,
			UpdatedAt: now,
			Dependencies: []*parser.Dependency{
				{
					IssueID:     "test-epic",
					DependsOnID: "test-blocker",
					Type:        parser.DepBlocks,
					CreatedAt:   now,
					CreatedBy:   "test",
				},
			},
		},
		{
			ID:        "test-child",
			Title:     "Child of Epic",
			Status:    parser.StatusOpen,
			Priority:  2,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
			Dependencies: []*parser.Dependency{
				{
					IssueID:     "test-child",
					DependsOnID: "test-epic",
					Type:        parser.DepParentChild,
					CreatedAt:   now,
					CreatedBy:   "test",
				},
			},
		},
	}

	state.LoadIssues(issues)
	state.SetViewMode(ViewTree)

	treeNodes := state.GetTreeNodes()

	// Should have 2 root nodes: the epic (always at root) and the blocker (no dependencies)
	if len(treeNodes) != 2 {
		t.Errorf("Expected 2 root nodes (epic + blocker), got %d", len(treeNodes))
	}

	// Find the epic node
	var epicNode *TreeNode
	for _, node := range treeNodes {
		if node.Issue.IssueType == parser.TypeEpic {
			epicNode = node
			break
		}
	}

	if epicNode == nil {
		t.Fatal("Expected to find epic at root level")
	}

	if epicNode.Issue.ID != "test-epic" {
		t.Errorf("Expected epic to be test-epic, got %s", epicNode.Issue.ID)
	}

	// Epic should have 1 child
	if len(epicNode.Children) != 1 {
		t.Errorf("Expected epic to have 1 child, got %d", len(epicNode.Children))
	}

	if epicNode.Children[0].Issue.ID != "test-child" {
		t.Errorf("Expected epic's child to be test-child, got %s", epicNode.Children[0].Issue.ID)
	}
}

func TestIDBasedParentChildRelationship(t *testing.T) {
	state := New()
	now := time.Now()

	// Test ID-based nesting (beads naming convention: parent.child)
	// e.g., tui-y4h is parent of tui-y4h.1, tui-y4h.2, tui-y4h.3
	issues := []*parser.Issue{
		{
			ID:        "tui-y4h",
			Title:     "Epic: bd upstream PRs",
			Status:    parser.StatusOpen,
			Priority:  2,
			IssueType: parser.TypeEpic,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "tui-y4h.1",
			Title:     "Add --type flag to bd update",
			Status:    parser.StatusOpen,
			Priority:  2,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
			// No explicit dependency - should be nested by ID convention
		},
		{
			ID:        "tui-y4h.2",
			Title:     "Restore type shortcuts",
			Status:    parser.StatusOpen,
			Priority:  2,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "tui-y4h.3",
			Title:     "Add --comments flag",
			Status:    parser.StatusOpen,
			Priority:  2,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "tui-other",
			Title:     "Unrelated issue",
			Status:    parser.StatusOpen,
			Priority:  3,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	state.LoadIssues(issues)
	state.SetViewMode(ViewTree)

	treeNodes := state.GetTreeNodes()

	// Should have 2 root nodes: epic (tui-y4h) and unrelated issue (tui-other)
	if len(treeNodes) != 2 {
		t.Errorf("Expected 2 root nodes, got %d", len(treeNodes))
		for i, node := range treeNodes {
			t.Logf("  Root %d: %s", i, node.Issue.ID)
		}
	}

	// Find the epic node
	var epicNode *TreeNode
	for _, node := range treeNodes {
		if node.Issue.ID == "tui-y4h" {
			epicNode = node
			break
		}
	}

	if epicNode == nil {
		t.Fatal("Expected to find tui-y4h epic at root level")
	}

	// Epic should have 3 children (tui-y4h.1, tui-y4h.2, tui-y4h.3)
	if len(epicNode.Children) != 3 {
		t.Errorf("Expected epic to have 3 children, got %d", len(epicNode.Children))
		for i, child := range epicNode.Children {
			t.Logf("  Child %d: %s", i, child.Issue.ID)
		}
	}

	// Verify children are the expected ones
	childIDs := make(map[string]bool)
	for _, child := range epicNode.Children {
		childIDs[child.Issue.ID] = true
	}

	expectedChildren := []string{"tui-y4h.1", "tui-y4h.2", "tui-y4h.3"}
	for _, expectedID := range expectedChildren {
		if !childIDs[expectedID] {
			t.Errorf("Expected to find child %s under epic", expectedID)
		}
	}
}

// TestBlockingPropagatesThroughParentChild verifies that blocking propagates
// through parent-child relationships, matching bd ready behavior
func TestBlockingPropagatesThroughParentChild(t *testing.T) {
	state := New()
	now := time.Now()

	// Setup:
	// - blocker (open) blocks epic
	// - epic has child-a and child-b via parent-child
	// - child-a has grandchild via parent-child
	// Expected: All of epic, child-a, child-b, grandchild should be blocked
	issues := []*parser.Issue{
		{
			ID:        "blocker",
			Title:     "Blocker Issue",
			Status:    parser.StatusOpen,
			Priority:  1,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "epic",
			Title:     "Epic",
			Status:    parser.StatusOpen,
			Priority:  2,
			IssueType: parser.TypeEpic,
			CreatedAt: now,
			UpdatedAt: now,
			Dependencies: []*parser.Dependency{
				{
					IssueID:     "epic",
					DependsOnID: "blocker",
					Type:        parser.DepBlocks,
					CreatedAt:   now,
					CreatedBy:   "test",
				},
			},
		},
		{
			ID:        "child-a",
			Title:     "Child A",
			Status:    parser.StatusOpen,
			Priority:  2,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
			Dependencies: []*parser.Dependency{
				{
					IssueID:     "child-a",
					DependsOnID: "epic",
					Type:        parser.DepParentChild,
					CreatedAt:   now,
					CreatedBy:   "test",
				},
			},
		},
		{
			ID:        "child-b",
			Title:     "Child B",
			Status:    parser.StatusOpen,
			Priority:  2,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
			Dependencies: []*parser.Dependency{
				{
					IssueID:     "child-b",
					DependsOnID: "epic",
					Type:        parser.DepParentChild,
					CreatedAt:   now,
					CreatedBy:   "test",
				},
			},
		},
		{
			ID:        "grandchild",
			Title:     "Grandchild",
			Status:    parser.StatusOpen,
			Priority:  2,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
			Dependencies: []*parser.Dependency{
				{
					IssueID:     "grandchild",
					DependsOnID: "child-a",
					Type:        parser.DepParentChild,
					CreatedAt:   now,
					CreatedBy:   "test",
				},
			},
		},
		{
			ID:        "unrelated",
			Title:     "Unrelated Issue",
			Status:    parser.StatusOpen,
			Priority:  3,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	state.LoadIssues(issues)

	// Only blocker and unrelated should be ready
	readyIssues := state.GetReadyIssues()
	if len(readyIssues) != 2 {
		t.Errorf("Expected 2 ready issues, got %d", len(readyIssues))
		for _, issue := range readyIssues {
			t.Logf("  Ready: %s", issue.ID)
		}
	}

	readyIDs := make(map[string]bool)
	for _, issue := range readyIssues {
		readyIDs[issue.ID] = true
	}

	if !readyIDs["blocker"] {
		t.Error("Expected 'blocker' to be ready")
	}
	if !readyIDs["unrelated"] {
		t.Error("Expected 'unrelated' to be ready")
	}

	// epic, child-a, child-b, grandchild should all be blocked
	blockedIssues := state.GetBlockedIssues()
	if len(blockedIssues) != 4 {
		t.Errorf("Expected 4 blocked issues, got %d", len(blockedIssues))
		for _, issue := range blockedIssues {
			t.Logf("  Blocked: %s", issue.ID)
		}
	}

	blockedIDs := make(map[string]bool)
	for _, issue := range blockedIssues {
		blockedIDs[issue.ID] = true
	}

	expectedBlocked := []string{"epic", "child-a", "child-b", "grandchild"}
	for _, id := range expectedBlocked {
		if !blockedIDs[id] {
			t.Errorf("Expected '%s' to be blocked", id)
		}
	}
}

// TestRelatedAndDiscoveredFromDontBlock verifies that related and discovered-from
// dependencies do NOT cause blocking
func TestRelatedAndDiscoveredFromDontBlock(t *testing.T) {
	state := New()
	now := time.Now()

	issues := []*parser.Issue{
		{
			ID:        "blocker",
			Title:     "Blocker Issue",
			Status:    parser.StatusOpen,
			Priority:  1,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "related-issue",
			Title:     "Related Issue",
			Status:    parser.StatusOpen,
			Priority:  2,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
			Dependencies: []*parser.Dependency{
				{
					IssueID:     "related-issue",
					DependsOnID: "blocker",
					Type:        parser.DepRelated,
					CreatedAt:   now,
					CreatedBy:   "test",
				},
			},
		},
		{
			ID:        "discovered-issue",
			Title:     "Discovered Issue",
			Status:    parser.StatusOpen,
			Priority:  2,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
			Dependencies: []*parser.Dependency{
				{
					IssueID:     "discovered-issue",
					DependsOnID: "blocker",
					Type:        parser.DepDiscoveredFrom,
					CreatedAt:   now,
					CreatedBy:   "test",
				},
			},
		},
	}

	state.LoadIssues(issues)

	// All three should be ready (related and discovered-from don't block)
	readyIssues := state.GetReadyIssues()
	if len(readyIssues) != 3 {
		t.Errorf("Expected 3 ready issues, got %d", len(readyIssues))
		for _, issue := range readyIssues {
			t.Logf("  Ready: %s", issue.ID)
		}
	}

	blockedIssues := state.GetBlockedIssues()
	if len(blockedIssues) != 0 {
		t.Errorf("Expected 0 blocked issues, got %d", len(blockedIssues))
		for _, issue := range blockedIssues {
			t.Logf("  Blocked: %s", issue.ID)
		}
	}
}

// TestExplicitBlockedStatusDoesNotPropagate verifies that explicit status:blocked
// does NOT propagate to children (only blocks dependencies propagate)
func TestExplicitBlockedStatusDoesNotPropagate(t *testing.T) {
	state := New()
	now := time.Now()

	issues := []*parser.Issue{
		{
			ID:        "parent",
			Title:     "Parent with status:blocked",
			Status:    parser.StatusBlocked, // Explicitly blocked
			Priority:  2,
			IssueType: parser.TypeEpic,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "child",
			Title:     "Child of blocked parent",
			Status:    parser.StatusOpen,
			Priority:  2,
			IssueType: parser.TypeTask,
			CreatedAt: now,
			UpdatedAt: now,
			Dependencies: []*parser.Dependency{
				{
					IssueID:     "child",
					DependsOnID: "parent",
					Type:        parser.DepParentChild,
					CreatedAt:   now,
					CreatedBy:   "test",
				},
			},
		},
	}

	state.LoadIssues(issues)

	// Child should be ready - explicit status doesn't propagate
	readyIssues := state.GetReadyIssues()
	if len(readyIssues) != 1 {
		t.Errorf("Expected 1 ready issue, got %d", len(readyIssues))
	}
	if len(readyIssues) > 0 && readyIssues[0].ID != "child" {
		t.Errorf("Expected 'child' to be ready, got %s", readyIssues[0].ID)
	}

	// Parent should be blocked (explicit status)
	blockedIssues := state.GetBlockedIssues()
	if len(blockedIssues) != 1 {
		t.Errorf("Expected 1 blocked issue, got %d", len(blockedIssues))
	}
	if len(blockedIssues) > 0 && blockedIssues[0].ID != "parent" {
		t.Errorf("Expected 'parent' to be blocked, got %s", blockedIssues[0].ID)
	}
}
