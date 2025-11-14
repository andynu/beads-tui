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
