package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/andy/beads-tui/internal/parser"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// SQLiteReader reads issues directly from .beads/beads.db
type SQLiteReader struct {
	db *sql.DB
}

// NewSQLiteReader creates a new SQLite reader for the given database path
// Opens in read-only mode to avoid any write locking
func NewSQLiteReader(dbPath string) (*SQLiteReader, error) {
	// Open in read-only mode using file: URI scheme
	// ncruces/go-sqlite3 requires file: prefix for proper WAL support
	db, err := sql.Open("sqlite3", "file:"+dbPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool limits to prevent resource exhaustion
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Verify the database has the expected schema
	var tableCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='issues'").Scan(&tableCount)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to verify schema: %w", err)
	}
	if tableCount == 0 {
		db.Close()
		return nil, fmt.Errorf("database does not contain issues table - has beads been initialized?")
	}

	return &SQLiteReader{db: db}, nil
}

// LoadIssues reads all issues from the database with dependencies, labels, and comments
// Uses read-only transaction to ensure consistent snapshot
func (r *SQLiteReader) LoadIssues(ctx context.Context) ([]*parser.Issue, error) {
	// Begin read-only transaction for consistent snapshot
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Safe to call even after commit

	// Query all issues
	rows, err := tx.QueryContext(ctx, `
		SELECT id, title, description, design, acceptance_criteria, notes,
		       status, priority, issue_type, assignee, estimated_minutes,
		       created_at, updated_at, closed_at, external_ref
		FROM issues
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query issues: %w", err)
	}
	defer rows.Close()

	var issues []*parser.Issue
	for rows.Next() {
		var issue parser.Issue
		var closedAt sql.NullTime
		var estimatedMinutes sql.NullInt64
		var assignee sql.NullString
		var externalRef sql.NullString

		err := rows.Scan(
			&issue.ID, &issue.Title, &issue.Description, &issue.Design,
			&issue.AcceptanceCriteria, &issue.Notes, &issue.Status,
			&issue.Priority, &issue.IssueType, &assignee, &estimatedMinutes,
			&issue.CreatedAt, &issue.UpdatedAt, &closedAt, &externalRef,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan issue: %w", err)
		}

		// Handle nullable fields
		if closedAt.Valid {
			issue.ClosedAt = &closedAt.Time
		}
		if estimatedMinutes.Valid {
			mins := int(estimatedMinutes.Int64)
			issue.EstimatedMinutes = &mins
		}
		if assignee.Valid {
			issue.Assignee = assignee.String
		}
		if externalRef.Valid {
			issue.ExternalRef = &externalRef.String
		}

		issues = append(issues, &issue)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating issues: %w", err)
	}

	// Load dependencies for all issues (within same transaction)
	deps, err := r.loadAllDependenciesTx(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to load dependencies: %w", err)
	}

	// Load labels for all issues (within same transaction)
	labels, err := r.loadAllLabelsTx(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to load labels: %w", err)
	}

	// Load comments for all issues (within same transaction)
	comments, err := r.loadAllCommentsTx(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to load comments: %w", err)
	}

	// Attach dependencies, labels, and comments to issues
	for _, issue := range issues {
		if issueDeps, ok := deps[issue.ID]; ok {
			issue.Dependencies = issueDeps
		}
		if issueLabels, ok := labels[issue.ID]; ok {
			issue.Labels = issueLabels
		}
		if issueComments, ok := comments[issue.ID]; ok {
			issue.Comments = issueComments
		}
	}

	// Read-only transaction can just be rolled back (no changes to commit)
	// Rollback is safe and releases locks

	return issues, nil
}

// loadAllDependenciesTx loads all dependencies indexed by issue ID within a transaction
func (r *SQLiteReader) loadAllDependenciesTx(ctx context.Context, tx *sql.Tx) (map[string][]*parser.Dependency, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT issue_id, depends_on_id, type
		FROM dependencies
		ORDER BY issue_id, depends_on_id
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query dependencies: %w", err)
	}
	defer rows.Close()

	deps := make(map[string][]*parser.Dependency)
	for rows.Next() {
		var issueID, dependsOnID string
		var depType parser.DependencyType

		if err := rows.Scan(&issueID, &dependsOnID, &depType); err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}

		deps[issueID] = append(deps[issueID], &parser.Dependency{
			DependsOnID: dependsOnID,
			Type:        depType,
		})
	}

	return deps, rows.Err()
}

// loadAllLabelsTx loads all labels indexed by issue ID within a transaction
func (r *SQLiteReader) loadAllLabelsTx(ctx context.Context, tx *sql.Tx) (map[string][]string, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT issue_id, label
		FROM labels
		ORDER BY issue_id, label
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query labels: %w", err)
	}
	defer rows.Close()

	labels := make(map[string][]string)
	for rows.Next() {
		var issueID, label string

		if err := rows.Scan(&issueID, &label); err != nil {
			return nil, fmt.Errorf("failed to scan label: %w", err)
		}

		labels[issueID] = append(labels[issueID], label)
	}

	return labels, rows.Err()
}

// loadAllCommentsTx loads all comments indexed by issue ID within a transaction
func (r *SQLiteReader) loadAllCommentsTx(ctx context.Context, tx *sql.Tx) (map[string][]*parser.Comment, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT issue_id, author, text, created_at
		FROM comments
		ORDER BY issue_id, created_at
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	comments := make(map[string][]*parser.Comment)
	for rows.Next() {
		var issueID, author, text string
		var createdAt time.Time

		if err := rows.Scan(&issueID, &author, &text, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}

		comments[issueID] = append(comments[issueID], &parser.Comment{
			IssueID:   issueID,
			Author:    author,
			Text:      text,
			CreatedAt: createdAt,
		})
	}

	return comments, rows.Err()
}

// Close closes the database connection
func (r *SQLiteReader) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
