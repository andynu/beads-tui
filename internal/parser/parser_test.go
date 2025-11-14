package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile(t *testing.T) {
	// Create a temporary JSONL file
	tmpDir := t.TempDir()
	jsonlPath := filepath.Join(tmpDir, "test.jsonl")

	content := `{"id":"test-1","title":"Test Issue 1","description":"Description 1","status":"open","priority":1,"issue_type":"task","created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z"}
{"id":"test-2","title":"Test Issue 2","description":"Description 2","status":"in_progress","priority":0,"issue_type":"bug","created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z","dependencies":[{"issue_id":"test-2","depends_on_id":"test-1","type":"blocks","created_at":"2025-01-01T00:00:00Z","created_by":"test"}]}
{"id":"test-3","title":"Test Issue 3","description":"Description 3","status":"closed","priority":2,"issue_type":"feature","created_at":"2025-01-01T00:00:00Z","updated_at":"2025-01-01T00:00:00Z","closed_at":"2025-01-02T00:00:00Z"}
`

	if err := os.WriteFile(jsonlPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the file
	issues, err := ParseFile(jsonlPath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Verify we got 3 issues
	if len(issues) != 3 {
		t.Fatalf("Expected 3 issues, got %d", len(issues))
	}

	// Verify first issue
	if issues[0].ID != "test-1" {
		t.Errorf("Expected ID 'test-1', got '%s'", issues[0].ID)
	}
	if issues[0].Title != "Test Issue 1" {
		t.Errorf("Expected title 'Test Issue 1', got '%s'", issues[0].Title)
	}
	if issues[0].Status != StatusOpen {
		t.Errorf("Expected status 'open', got '%s'", issues[0].Status)
	}
	if issues[0].Priority != 1 {
		t.Errorf("Expected priority 1, got %d", issues[0].Priority)
	}

	// Verify second issue has dependency
	if len(issues[1].Dependencies) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(issues[1].Dependencies))
	}
	if issues[1].Dependencies[0].Type != DepBlocks {
		t.Errorf("Expected dependency type 'blocks', got '%s'", issues[1].Dependencies[0].Type)
	}

	// Verify third issue is closed
	if issues[2].Status != StatusClosed {
		t.Errorf("Expected status 'closed', got '%s'", issues[2].Status)
	}
	if issues[2].ClosedAt == nil {
		t.Error("Expected closed_at to be set")
	}
}

func TestParseEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	jsonlPath := filepath.Join(tmpDir, "empty.jsonl")

	if err := os.WriteFile(jsonlPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	issues, err := ParseFile(jsonlPath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestParseInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	jsonlPath := filepath.Join(tmpDir, "invalid.jsonl")

	content := `{"id":"test-1","title":"Valid"}
{invalid json}
{"id":"test-3","title":"Also valid"}`

	if err := os.WriteFile(jsonlPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := ParseFile(jsonlPath)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}
