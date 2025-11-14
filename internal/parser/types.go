package parser

import "time"

// Issue represents a beads issue from JSONL
type Issue struct {
	ID                 string        `json:"id"`
	ContentHash        string        `json:"content_hash,omitempty"`
	Title              string        `json:"title"`
	Description        string        `json:"description"`
	Design             string        `json:"design,omitempty"`
	AcceptanceCriteria string        `json:"acceptance_criteria,omitempty"`
	Notes              string        `json:"notes,omitempty"`
	Status             Status        `json:"status"`
	Priority           int           `json:"priority"`
	IssueType          IssueType     `json:"issue_type"`
	Assignee           string        `json:"assignee,omitempty"`
	EstimatedMinutes   *int          `json:"estimated_minutes,omitempty"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
	ClosedAt           *time.Time    `json:"closed_at,omitempty"`
	ExternalRef        *string       `json:"external_ref,omitempty"`
	SourceRepo         string        `json:"source_repo,omitempty"`
	Labels             []string      `json:"labels,omitempty"`
	Dependencies       []*Dependency `json:"dependencies,omitempty"`
	Comments           []*Comment    `json:"comments,omitempty"`
}

// Status represents the current state of an issue
type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusBlocked    Status = "blocked"
	StatusClosed     Status = "closed"
)

// IssueType categorizes the kind of work
type IssueType string

const (
	TypeBug     IssueType = "bug"
	TypeFeature IssueType = "feature"
	TypeTask    IssueType = "task"
	TypeEpic    IssueType = "epic"
	TypeChore   IssueType = "chore"
)

// Dependency represents a relationship between issues
type Dependency struct {
	IssueID     string         `json:"issue_id"`
	DependsOnID string         `json:"depends_on_id"`
	Type        DependencyType `json:"type"`
	CreatedAt   time.Time      `json:"created_at"`
	CreatedBy   string         `json:"created_by"`
}

// DependencyType categorizes the relationship
type DependencyType string

const (
	DepBlocks         DependencyType = "blocks"
	DepRelated        DependencyType = "related"
	DepParentChild    DependencyType = "parent-child"
	DepDiscoveredFrom DependencyType = "discovered-from"
)

// Comment represents a comment on an issue
type Comment struct {
	ID        int64     `json:"id"`
	IssueID   string    `json:"issue_id"`
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
