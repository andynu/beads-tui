package storage

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andy/beads-tui/internal/parser"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// setupTestDB creates a temporary database with the beads schema
func setupTestDB(t *testing.T) (string, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "beads-tui-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")

	// Create database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	// Create schema
	schema := `
		CREATE TABLE issues (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT DEFAULT '',
			design TEXT DEFAULT '',
			acceptance_criteria TEXT DEFAULT '',
			notes TEXT DEFAULT '',
			status TEXT DEFAULT 'open',
			priority INTEGER DEFAULT 2,
			issue_type TEXT DEFAULT 'task',
			assignee TEXT,
			estimated_minutes INTEGER,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			closed_at TIMESTAMP,
			external_ref TEXT
		);

		CREATE TABLE dependencies (
			issue_id TEXT NOT NULL,
			depends_on_id TEXT NOT NULL,
			type TEXT NOT NULL,
			PRIMARY KEY (issue_id, depends_on_id, type)
		);

		CREATE TABLE labels (
			issue_id TEXT NOT NULL,
			label TEXT NOT NULL,
			PRIMARY KEY (issue_id, label)
		);

		CREATE TABLE comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			issue_id TEXT NOT NULL,
			author TEXT NOT NULL,
			text TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`

	if _, err := db.Exec(schema); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create schema: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return dbPath, cleanup
}

func TestNewSQLiteReader(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	reader, err := NewSQLiteReader(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteReader failed: %v", err)
	}
	defer reader.Close()

	if reader.db == nil {
		t.Fatal("Expected db to be initialized")
	}
}

func TestNewSQLiteReader_NonexistentDB(t *testing.T) {
	_, err := NewSQLiteReader("/nonexistent/path/db.sqlite")
	if err == nil {
		t.Fatal("Expected error for nonexistent database")
	}
}

func TestNewSQLiteReader_InvalidSchema(t *testing.T) {
	// Create a database without the issues table
	tmpDir, err := os.MkdirTemp("", "beads-tui-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "invalid.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	// Create an empty table (not the issues table)
	_, err = db.Exec("CREATE TABLE dummy (id INTEGER)")
	if err != nil {
		db.Close()
		t.Fatalf("failed to create dummy table: %v", err)
	}
	db.Close()

	// Try to open with SQLiteReader - should fail validation
	_, err = NewSQLiteReader(dbPath)
	if err == nil {
		t.Fatal("Expected error for database without issues table")
	}
	if err.Error() != "database does not contain issues table - has beads been initialized?" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestLoadIssues_Empty(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	reader, err := NewSQLiteReader(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteReader failed: %v", err)
	}
	defer reader.Close()

	ctx := context.Background()
	issues, err := reader.LoadIssues(ctx)
	if err != nil {
		t.Fatalf("LoadIssues failed: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestLoadIssues_BasicIssue(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	// Insert test data
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Truncate(time.Second)
	_, err = db.Exec(`
		INSERT INTO issues (id, title, description, status, priority, issue_type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, "test-123", "Test Issue", "Test description", "open", 1, "feature", now, now)
	if err != nil {
		t.Fatalf("failed to insert test issue: %v", err)
	}

	// Load issues
	reader, err := NewSQLiteReader(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteReader failed: %v", err)
	}
	defer reader.Close()

	ctx := context.Background()
	issues, err := reader.LoadIssues(ctx)
	if err != nil {
		t.Fatalf("LoadIssues failed: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}

	issue := issues[0]
	if issue.ID != "test-123" {
		t.Errorf("Expected ID 'test-123', got '%s'", issue.ID)
	}
	if issue.Title != "Test Issue" {
		t.Errorf("Expected title 'Test Issue', got '%s'", issue.Title)
	}
	if issue.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got '%s'", issue.Description)
	}
	if issue.Status != parser.StatusOpen {
		t.Errorf("Expected status 'open', got '%s'", issue.Status)
	}
	if issue.Priority != 1 {
		t.Errorf("Expected priority 1, got %d", issue.Priority)
	}
	if issue.IssueType != parser.TypeFeature {
		t.Errorf("Expected type 'feature', got '%s'", issue.IssueType)
	}
}

func TestLoadIssues_WithDependencies(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Truncate(time.Second)

	// Insert two issues
	_, err = db.Exec(`
		INSERT INTO issues (id, title, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?), (?, ?, ?, ?, ?)
	`, "test-1", "Issue 1", "open", now, now,
		"test-2", "Issue 2", "open", now, now)
	if err != nil {
		t.Fatalf("failed to insert issues: %v", err)
	}

	// Add dependency: test-1 blocks test-2
	_, err = db.Exec(`
		INSERT INTO dependencies (issue_id, depends_on_id, type)
		VALUES (?, ?, ?)
	`, "test-2", "test-1", "blocks")
	if err != nil {
		t.Fatalf("failed to insert dependency: %v", err)
	}

	// Load issues
	reader, err := NewSQLiteReader(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteReader failed: %v", err)
	}
	defer reader.Close()

	ctx := context.Background()
	issues, err := reader.LoadIssues(ctx)
	if err != nil {
		t.Fatalf("LoadIssues failed: %v", err)
	}

	// Find test-2 (the issue with the dependency)
	var issue2 *parser.Issue
	for _, iss := range issues {
		if iss.ID == "test-2" {
			issue2 = iss
			break
		}
	}

	if issue2 == nil {
		t.Fatal("Could not find test-2")
	}

	if len(issue2.Dependencies) != 1 {
		t.Fatalf("Expected 1 dependency, got %d", len(issue2.Dependencies))
	}

	dep := issue2.Dependencies[0]
	if dep.DependsOnID != "test-1" {
		t.Errorf("Expected dependency on 'test-1', got '%s'", dep.DependsOnID)
	}
	if dep.Type != parser.DepBlocks {
		t.Errorf("Expected type 'blocks', got '%s'", dep.Type)
	}
}

func TestLoadIssues_WithLabels(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Truncate(time.Second)

	// Insert issue
	_, err = db.Exec(`
		INSERT INTO issues (id, title, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, "test-1", "Issue 1", "open", now, now)
	if err != nil {
		t.Fatalf("failed to insert issue: %v", err)
	}

	// Add labels
	_, err = db.Exec(`
		INSERT INTO labels (issue_id, label)
		VALUES (?, ?), (?, ?)
	`, "test-1", "bug", "test-1", "urgent")
	if err != nil {
		t.Fatalf("failed to insert labels: %v", err)
	}

	// Load issues
	reader, err := NewSQLiteReader(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteReader failed: %v", err)
	}
	defer reader.Close()

	ctx := context.Background()
	issues, err := reader.LoadIssues(ctx)
	if err != nil {
		t.Fatalf("LoadIssues failed: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}

	issue := issues[0]
	if len(issue.Labels) != 2 {
		t.Fatalf("Expected 2 labels, got %d", len(issue.Labels))
	}

	// Labels are returned in sorted order
	expectedLabels := []string{"bug", "urgent"}
	for i, label := range issue.Labels {
		if label != expectedLabels[i] {
			t.Errorf("Expected label '%s', got '%s'", expectedLabels[i], label)
		}
	}
}

func TestLoadIssues_WithComments(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Truncate(time.Second)
	commentTime := now.Add(1 * time.Hour)

	// Insert issue
	_, err = db.Exec(`
		INSERT INTO issues (id, title, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, "test-1", "Issue 1", "open", now, now)
	if err != nil {
		t.Fatalf("failed to insert issue: %v", err)
	}

	// Add comment
	_, err = db.Exec(`
		INSERT INTO comments (issue_id, author, text, created_at)
		VALUES (?, ?, ?, ?)
	`, "test-1", "alice", "This is a comment", commentTime)
	if err != nil {
		t.Fatalf("failed to insert comment: %v", err)
	}

	// Load issues
	reader, err := NewSQLiteReader(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteReader failed: %v", err)
	}
	defer reader.Close()

	ctx := context.Background()
	issues, err := reader.LoadIssues(ctx)
	if err != nil {
		t.Fatalf("LoadIssues failed: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}

	issue := issues[0]
	if len(issue.Comments) != 1 {
		t.Fatalf("Expected 1 comment, got %d", len(issue.Comments))
	}

	comment := issue.Comments[0]
	if comment.Author != "alice" {
		t.Errorf("Expected author 'alice', got '%s'", comment.Author)
	}
	if comment.Text != "This is a comment" {
		t.Errorf("Expected text 'This is a comment', got '%s'", comment.Text)
	}
	if !comment.CreatedAt.Equal(commentTime) {
		t.Errorf("Expected created_at %v, got %v", commentTime, comment.CreatedAt)
	}
}

func TestLoadIssues_NullableFields(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Truncate(time.Second)
	closedAt := now.Add(1 * time.Hour)

	// Insert issue with all nullable fields set
	_, err = db.Exec(`
		INSERT INTO issues (id, title, status, assignee, estimated_minutes, closed_at, external_ref, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "test-1", "Issue 1", "closed", "bob", 120, closedAt, "JIRA-123", now, now)
	if err != nil {
		t.Fatalf("failed to insert issue: %v", err)
	}

	// Load issues
	reader, err := NewSQLiteReader(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteReader failed: %v", err)
	}
	defer reader.Close()

	ctx := context.Background()
	issues, err := reader.LoadIssues(ctx)
	if err != nil {
		t.Fatalf("LoadIssues failed: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}

	issue := issues[0]
	if issue.Assignee != "bob" {
		t.Errorf("Expected assignee 'bob', got '%s'", issue.Assignee)
	}
	if issue.EstimatedMinutes == nil || *issue.EstimatedMinutes != 120 {
		t.Errorf("Expected estimated_minutes 120, got %v", issue.EstimatedMinutes)
	}
	if issue.ClosedAt == nil || !issue.ClosedAt.Equal(closedAt) {
		t.Errorf("Expected closed_at %v, got %v", closedAt, issue.ClosedAt)
	}
	if issue.ExternalRef == nil || *issue.ExternalRef != "JIRA-123" {
		t.Errorf("Expected external_ref 'JIRA-123', got %v", issue.ExternalRef)
	}
}

func TestLoadIssues_ContextCancellation(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	reader, err := NewSQLiteReader(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteReader failed: %v", err)
	}
	defer reader.Close()

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = reader.LoadIssues(ctx)
	if err == nil {
		t.Fatal("Expected error with cancelled context")
	}
	// Error should mention context cancellation (wrapped error)
	if err.Error() != "failed to begin transaction: context canceled" {
		t.Errorf("Expected context cancellation error, got: %v", err)
	}
}

func TestClose(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	reader, err := NewSQLiteReader(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteReader failed: %v", err)
	}

	// Close should succeed
	if err := reader.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Close again should be safe
	if err := reader.Close(); err != nil {
		t.Errorf("Second Close failed: %v", err)
	}
}

func TestClose_NilDB(t *testing.T) {
	reader := &SQLiteReader{db: nil}
	if err := reader.Close(); err != nil {
		t.Errorf("Close with nil db failed: %v", err)
	}
}
