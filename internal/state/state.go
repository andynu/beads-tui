package state

import (
	"github.com/andy/beads-tui/internal/parser"
)

// State manages the application state and issue data
type State struct {
	issues           []*parser.Issue
	issuesByID       map[string]*parser.Issue
	readyIssues      []*parser.Issue
	blockedIssues    []*parser.Issue
	inProgressIssues []*parser.Issue
	closedIssues     []*parser.Issue
	selectedIndex    int
	selectedIssue    *parser.Issue
	filterMode       FilterMode
	searchQuery      string
}

// FilterMode represents different filtering options
type FilterMode int

const (
	FilterAll FilterMode = iota
	FilterReady
	FilterBlocked
	FilterInProgress
	FilterByPriority
	FilterByType
)

// New creates a new application state
func New() *State {
	return &State{
		issuesByID: make(map[string]*parser.Issue),
		filterMode: FilterAll,
	}
}

// LoadIssues updates the state with a new set of issues
func (s *State) LoadIssues(issues []*parser.Issue) {
	s.issues = issues
	s.issuesByID = make(map[string]*parser.Issue)

	// Clear categorized lists
	s.readyIssues = nil
	s.blockedIssues = nil
	s.inProgressIssues = nil
	s.closedIssues = nil

	// Index issues by ID
	for _, issue := range issues {
		s.issuesByID[issue.ID] = issue
	}

	// Categorize issues
	s.categorizeIssues()
}

// categorizeIssues separates issues into ready, blocked, in_progress, and closed
func (s *State) categorizeIssues() {
	// Build a map of issues that are blocked by open dependencies
	blockedByIssueIDs := make(map[string]bool)

	for _, issue := range s.issues {
		// Check if this issue blocks any other issues
		for _, dep := range issue.Dependencies {
			// If this is a "blocks" dependency and the blocking issue is not closed
			if dep.Type == parser.DepBlocks && issue.Status != parser.StatusClosed {
				// The issue that depends on this one (dep.DependsOnID) is blocked
				// Note: in beads JSONL, dependencies are stored on the issue that has them
				// dep.IssueID is the current issue, dep.DependsOnID is what it depends on
				// So we need to check if current issue blocks others by looking at reverse deps
			}
		}
	}

	// Build reverse dependency map (who blocks whom)
	// For each issue, find all issues it blocks
	for _, issue := range s.issues {
		for _, dep := range issue.Dependencies {
			if dep.Type == parser.DepBlocks {
				// issue depends on dep.DependsOnID
				// So dep.DependsOnID blocks issue
				targetIssue := s.issuesByID[dep.DependsOnID]
				if targetIssue != nil && targetIssue.Status != parser.StatusClosed {
					// This issue is blocked by an open dependency
					blockedByIssueIDs[issue.ID] = true
				}
			}
		}
	}

	// Categorize each issue
	for _, issue := range s.issues {
		switch issue.Status {
		case parser.StatusClosed:
			s.closedIssues = append(s.closedIssues, issue)
		case parser.StatusInProgress:
			s.inProgressIssues = append(s.inProgressIssues, issue)
		case parser.StatusBlocked:
			s.blockedIssues = append(s.blockedIssues, issue)
		case parser.StatusOpen:
			// Check if actually blocked by dependencies
			if blockedByIssueIDs[issue.ID] {
				s.blockedIssues = append(s.blockedIssues, issue)
			} else {
				s.readyIssues = append(s.readyIssues, issue)
			}
		}
	}
}

// GetReadyIssues returns issues that are ready to work on
func (s *State) GetReadyIssues() []*parser.Issue {
	return s.readyIssues
}

// GetBlockedIssues returns issues that are blocked
func (s *State) GetBlockedIssues() []*parser.Issue {
	return s.blockedIssues
}

// GetInProgressIssues returns issues that are in progress
func (s *State) GetInProgressIssues() []*parser.Issue {
	return s.inProgressIssues
}

// GetClosedIssues returns closed issues
func (s *State) GetClosedIssues() []*parser.Issue {
	return s.closedIssues
}

// GetAllIssues returns all issues
func (s *State) GetAllIssues() []*parser.Issue {
	return s.issues
}

// GetIssueByID returns an issue by its ID
func (s *State) GetIssueByID(id string) *parser.Issue {
	return s.issuesByID[id]
}

// SetSelectedIssue sets the currently selected issue
func (s *State) SetSelectedIssue(issue *parser.Issue) {
	s.selectedIssue = issue
}

// GetSelectedIssue returns the currently selected issue
func (s *State) GetSelectedIssue() *parser.Issue {
	return s.selectedIssue
}
