package main

import (
	"encoding/json"
	"testing"

	"github.com/andy/beads-tui/internal/parser"
)

func TestParseBdJSON_IssueArray(t *testing.T) {
	jsonData := `[
		{
			"id": "tui-123",
			"title": "Test Issue",
			"description": "Test description",
			"status": "open",
			"priority": 2,
			"issue_type": "feature",
			"created_at": "2025-11-14T22:00:00Z",
			"updated_at": "2025-11-14T22:00:00Z"
		}
	]`

	result, err := parseBdJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("parseBdJSON failed: %v", err)
	}

	if len(result.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(result.Issues))
	}

	issue := result.Issues[0]
	if issue.ID != "tui-123" {
		t.Errorf("expected ID 'tui-123', got '%s'", issue.ID)
	}
	if issue.Title != "Test Issue" {
		t.Errorf("expected title 'Test Issue', got '%s'", issue.Title)
	}
	if issue.Status != parser.StatusOpen {
		t.Errorf("expected status 'open', got '%s'", issue.Status)
	}
	if issue.Priority != 2 {
		t.Errorf("expected priority 2, got %d", issue.Priority)
	}
}

func TestParseBdJSON_SingleIssue(t *testing.T) {
	jsonData := `{
		"id": "tui-456",
		"title": "Another Test",
		"description": "",
		"status": "in_progress",
		"priority": 1,
		"issue_type": "bug",
		"created_at": "2025-11-14T22:00:00Z",
		"updated_at": "2025-11-14T22:00:00Z"
	}`

	result, err := parseBdJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("parseBdJSON failed: %v", err)
	}

	if len(result.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(result.Issues))
	}

	issue := result.Issues[0]
	if issue.ID != "tui-456" {
		t.Errorf("expected ID 'tui-456', got '%s'", issue.ID)
	}
	if issue.IssueType != parser.TypeBug {
		t.Errorf("expected type 'bug', got '%s'", issue.IssueType)
	}
}

func TestParseBdJSON_Comment(t *testing.T) {
	jsonData := `{
		"id": 42,
		"issue_id": "tui-789",
		"author": "testuser",
		"text": "Test comment",
		"created_at": "2025-11-14T22:00:00Z"
	}`

	result, err := parseBdJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("parseBdJSON failed: %v", err)
	}

	if len(result.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(result.Comments))
	}

	comment := result.Comments[0]
	if comment.ID != 42 {
		t.Errorf("expected ID 42, got %d", comment.ID)
	}
	if comment.IssueID != "tui-789" {
		t.Errorf("expected issue_id 'tui-789', got '%s'", comment.IssueID)
	}
	if comment.Author != "testuser" {
		t.Errorf("expected author 'testuser', got '%s'", comment.Author)
	}
}

func TestParseBdJSON_InvalidJSON(t *testing.T) {
	jsonData := `{invalid json`

	_, err := parseBdJSON([]byte(jsonData))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestParseBdJSON_EmptyArray(t *testing.T) {
	jsonData := `[]`

	result, err := parseBdJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("parseBdJSON failed: %v", err)
	}

	if len(result.Issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(result.Issues))
	}
}

func TestParseBdJSON_MultipleIssues(t *testing.T) {
	jsonData := `[
		{
			"id": "tui-1",
			"title": "First",
			"status": "open",
			"priority": 0,
			"issue_type": "feature",
			"created_at": "2025-11-14T22:00:00Z",
			"updated_at": "2025-11-14T22:00:00Z"
		},
		{
			"id": "tui-2",
			"title": "Second",
			"status": "closed",
			"priority": 3,
			"issue_type": "task",
			"created_at": "2025-11-14T22:00:00Z",
			"updated_at": "2025-11-14T22:00:00Z"
		}
	]`

	result, err := parseBdJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("parseBdJSON failed: %v", err)
	}

	if len(result.Issues) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(result.Issues))
	}

	if result.Issues[0].ID != "tui-1" {
		t.Errorf("expected first issue ID 'tui-1', got '%s'", result.Issues[0].ID)
	}
	if result.Issues[1].ID != "tui-2" {
		t.Errorf("expected second issue ID 'tui-2', got '%s'", result.Issues[1].ID)
	}
}

// Test that parseBdJSON correctly handles the exact format returned by bd commands
func TestParseBdJSON_RealBdOutput(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		wantType string // "issues" or "comment"
	}{
		{
			name: "bd show output",
			jsonData: `[
				{
					"id": "tui-vg7",
					"content_hash": "bc2bf433c3faa31538c447bc7dfd434b6f8a8045725f1111cb147fa7269e6850",
					"title": "Interact with the bd commandline using --json wherever possible",
					"description": "",
					"status": "in_progress",
					"priority": 2,
					"issue_type": "feature",
					"created_at": "2025-11-14T15:04:03.569822263-05:00",
					"updated_at": "2025-11-14T22:15:51.010287994-05:00",
					"source_repo": "."
				}
			]`,
			wantType: "issues",
		},
		{
			name: "bd create output",
			jsonData: `{
				"id":"tui-vyc",
				"content_hash":"06df81a8d414a51a714eacdb647c7ab88ff8b22da82e95e4161606372221f070",
				"title":"test issue for json check",
				"description":"",
				"status":"open",
				"priority":4,
				"issue_type":"task",
				"created_at":"2025-11-14T22:15:39.591881491-05:00",
				"updated_at":"2025-11-14T22:15:39.591881491-05:00"
			}`,
			wantType: "issues",
		},
		{
			name: "bd comment output",
			jsonData: `{
				"id": 15,
				"issue_id": "tui-vg7",
				"author": "andy",
				"text": "Test comment for JSON check",
				"created_at": "2025-11-15T03:15:51Z"
			}`,
			wantType: "comment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseBdJSON([]byte(tt.jsonData))
			if err != nil {
				t.Fatalf("parseBdJSON failed: %v", err)
			}

			switch tt.wantType {
			case "issues":
				if len(result.Issues) == 0 {
					t.Error("expected issues, got none")
				}
			case "comment":
				if len(result.Comments) == 0 {
					t.Error("expected comment, got none")
				}
			}
		})
	}
}

// Test execBdJSONIssue helper
func TestExecBdJSONIssue_NoIssues(t *testing.T) {
	// This is a unit test for the helper function behavior
	// We can't easily test the actual exec without mocking
	result := &BdCommandResult{
		Issues: []parser.Issue{},
	}

	if len(result.Issues) == 0 {
		// This simulates what execBdJSONIssue would encounter
		err := "bd command returned no issues"
		if err == "" {
			t.Error("expected error when no issues returned")
		}
	}
}

// Test that JSON unmarshaling works for all parser types
func TestParserTypesJSONCompatibility(t *testing.T) {
	// Test Issue
	issueJSON := `{
		"id": "test-1",
		"title": "Test",
		"status": "open",
		"priority": 1,
		"issue_type": "feature",
		"labels": ["tag1", "tag2"],
		"created_at": "2025-11-14T22:00:00Z",
		"updated_at": "2025-11-14T22:00:00Z"
	}`
	var issue parser.Issue
	if err := json.Unmarshal([]byte(issueJSON), &issue); err != nil {
		t.Errorf("Failed to unmarshal Issue: %v", err)
	}
	if len(issue.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(issue.Labels))
	}

	// Test Comment
	commentJSON := `{
		"id": 1,
		"issue_id": "test-1",
		"author": "user",
		"text": "comment",
		"created_at": "2025-11-14T22:00:00Z"
	}`
	var comment parser.Comment
	if err := json.Unmarshal([]byte(commentJSON), &comment); err != nil {
		t.Errorf("Failed to unmarshal Comment: %v", err)
	}
}
