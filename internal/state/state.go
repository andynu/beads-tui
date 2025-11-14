package state

import (
	"fmt"
	"strings"

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
	viewMode         ViewMode
	treeNodes        []*TreeNode

	// Filter state
	priorityFilter map[int]bool           // nil = no filter, otherwise only show these priorities
	typeFilter     map[parser.IssueType]bool // nil = no filter, otherwise only show these types
	statusFilter   map[parser.Status]bool    // nil = no filter, otherwise only show these statuses
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

// ViewMode represents different view layouts
type ViewMode int

const (
	ViewList ViewMode = iota
	ViewTree
)

// TreeNode represents a node in the dependency tree
type TreeNode struct {
	Issue    *parser.Issue
	Children []*TreeNode
	Depth    int
}

// New creates a new application state
func New() *State {
	return &State{
		issuesByID: make(map[string]*parser.Issue),
		filterMode: FilterAll,
		viewMode:   ViewList,
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

// applyFilters filters a list of issues based on active filters
func (s *State) applyFilters(issues []*parser.Issue) []*parser.Issue {
	if s.priorityFilter == nil && s.typeFilter == nil && s.statusFilter == nil {
		return issues
	}

	var filtered []*parser.Issue
	for _, issue := range issues {
		// Check priority filter
		if s.priorityFilter != nil && !s.priorityFilter[issue.Priority] {
			continue
		}

		// Check type filter
		if s.typeFilter != nil && !s.typeFilter[issue.IssueType] {
			continue
		}

		// Check status filter
		if s.statusFilter != nil && !s.statusFilter[issue.Status] {
			continue
		}

		filtered = append(filtered, issue)
	}
	return filtered
}

// GetReadyIssues returns issues that are ready to work on
func (s *State) GetReadyIssues() []*parser.Issue {
	return s.applyFilters(s.readyIssues)
}

// GetBlockedIssues returns issues that are blocked
func (s *State) GetBlockedIssues() []*parser.Issue {
	return s.applyFilters(s.blockedIssues)
}

// GetInProgressIssues returns issues that are in progress
func (s *State) GetInProgressIssues() []*parser.Issue {
	return s.applyFilters(s.inProgressIssues)
}

// GetClosedIssues returns closed issues
func (s *State) GetClosedIssues() []*parser.Issue {
	return s.applyFilters(s.closedIssues)
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

// SetViewMode sets the current view mode
func (s *State) SetViewMode(mode ViewMode) {
	s.viewMode = mode
	if mode == ViewTree {
		s.buildDependencyTree()
	}
}

// GetViewMode returns the current view mode
func (s *State) GetViewMode() ViewMode {
	return s.viewMode
}

// ToggleViewMode switches between list and tree view
func (s *State) ToggleViewMode() ViewMode {
	if s.viewMode == ViewList {
		s.SetViewMode(ViewTree)
	} else {
		s.SetViewMode(ViewList)
	}
	return s.viewMode
}

// GetTreeNodes returns the tree structure for tree view
func (s *State) GetTreeNodes() []*TreeNode {
	return s.treeNodes
}

// buildDependencyTree constructs a tree structure from issue dependencies
func (s *State) buildDependencyTree() {
	s.treeNodes = nil

	// Build maps for parent-child and blocks relationships
	childrenMap := make(map[string][]*parser.Issue)       // parent ID -> children
	blockedByMap := make(map[string][]*parser.Issue)      // blocker ID -> blocked issues
	hasIncomingDep := make(map[string]bool)               // issues that have parents or blockers

	// First pass: build relationship maps
	for _, issue := range s.issues {
		// Skip closed issues in tree view
		if issue.Status == parser.StatusClosed {
			continue
		}

		for _, dep := range issue.Dependencies {
			switch dep.Type {
			case parser.DepParentChild:
				// issue is a child of dep.DependsOnID
				parent := s.issuesByID[dep.DependsOnID]
				if parent != nil && parent.Status != parser.StatusClosed {
					childrenMap[dep.DependsOnID] = append(childrenMap[dep.DependsOnID], issue)
					hasIncomingDep[issue.ID] = true
				}
			case parser.DepBlocks:
				// issue depends on (is blocked by) dep.DependsOnID
				blocker := s.issuesByID[dep.DependsOnID]
				if blocker != nil && blocker.Status != parser.StatusClosed {
					blockedByMap[dep.DependsOnID] = append(blockedByMap[dep.DependsOnID], issue)
					hasIncomingDep[issue.ID] = true
				}
			}
		}
	}

	// Second pass: find root nodes (issues with no incoming dependencies)
	var rootIssues []*parser.Issue
	for _, issue := range s.issues {
		if issue.Status != parser.StatusClosed && !hasIncomingDep[issue.ID] {
			rootIssues = append(rootIssues, issue)
		}
	}

	// Build tree recursively from roots
	visited := make(map[string]bool)
	for _, root := range rootIssues {
		node := s.buildTreeNode(root, 0, childrenMap, blockedByMap, visited)
		if node != nil {
			s.treeNodes = append(s.treeNodes, node)
		}
	}
}

// buildTreeNode recursively builds a tree node and its children
func (s *State) buildTreeNode(issue *parser.Issue, depth int, childrenMap map[string][]*parser.Issue, blockedByMap map[string][]*parser.Issue, visited map[string]bool) *TreeNode {
	// Prevent cycles
	if visited[issue.ID] {
		return nil
	}
	visited[issue.ID] = true

	node := &TreeNode{
		Issue:    issue,
		Children: nil,
		Depth:    depth,
	}

	// Add children (from parent-child relationships)
	if children, ok := childrenMap[issue.ID]; ok {
		for _, child := range children {
			if childNode := s.buildTreeNode(child, depth+1, childrenMap, blockedByMap, visited); childNode != nil {
				node.Children = append(node.Children, childNode)
			}
		}
	}

	// Add blocked issues (from blocks relationships)
	if blocked, ok := blockedByMap[issue.ID]; ok {
		for _, blockedIssue := range blocked {
			if blockedNode := s.buildTreeNode(blockedIssue, depth+1, childrenMap, blockedByMap, visited); blockedNode != nil {
				node.Children = append(node.Children, blockedNode)
			}
		}
	}

	return node
}

// TogglePriorityFilter toggles a priority in the filter
func (s *State) TogglePriorityFilter(priority int) {
	if s.priorityFilter == nil {
		s.priorityFilter = make(map[int]bool)
	}

	if s.priorityFilter[priority] {
		delete(s.priorityFilter, priority)
		// If empty, set to nil to disable filtering
		if len(s.priorityFilter) == 0 {
			s.priorityFilter = nil
		}
	} else {
		s.priorityFilter[priority] = true
	}
}

// ToggleTypeFilter toggles an issue type in the filter
func (s *State) ToggleTypeFilter(issueType parser.IssueType) {
	if s.typeFilter == nil {
		s.typeFilter = make(map[parser.IssueType]bool)
	}

	if s.typeFilter[issueType] {
		delete(s.typeFilter, issueType)
		if len(s.typeFilter) == 0 {
			s.typeFilter = nil
		}
	} else {
		s.typeFilter[issueType] = true
	}
}

// ToggleStatusFilter toggles a status in the filter
func (s *State) ToggleStatusFilter(status parser.Status) {
	if s.statusFilter == nil {
		s.statusFilter = make(map[parser.Status]bool)
	}

	if s.statusFilter[status] {
		delete(s.statusFilter, status)
		if len(s.statusFilter) == 0 {
			s.statusFilter = nil
		}
	} else {
		s.statusFilter[status] = true
	}
}

// ClearAllFilters removes all active filters
func (s *State) ClearAllFilters() {
	s.priorityFilter = nil
	s.typeFilter = nil
	s.statusFilter = nil
}

// IsPriorityFiltered returns true if the given priority is in the active filter
func (s *State) IsPriorityFiltered(priority int) bool {
	return s.priorityFilter != nil && s.priorityFilter[priority]
}

// IsTypeFiltered returns true if the given type is in the active filter
func (s *State) IsTypeFiltered(issueType parser.IssueType) bool {
	return s.typeFilter != nil && s.typeFilter[issueType]
}

// IsStatusFiltered returns true if the given status is in the active filter
func (s *State) IsStatusFiltered(status parser.Status) bool {
	return s.statusFilter != nil && s.statusFilter[status]
}

// HasActiveFilters returns true if any filters are active
func (s *State) HasActiveFilters() bool {
	return s.priorityFilter != nil || s.typeFilter != nil || s.statusFilter != nil
}

// GetActiveFilters returns a human-readable description of active filters
func (s *State) GetActiveFilters() string {
	if !s.HasActiveFilters() {
		return ""
	}

	var filters []string

	// Priority filters
	if s.priorityFilter != nil {
		var priorities []string
		for p := 0; p <= 4; p++ {
			if s.priorityFilter[p] {
				priorities = append(priorities, fmt.Sprintf("P%d", p))
			}
		}
		if len(priorities) > 0 {
			filters = append(filters, "Priority: "+strings.Join(priorities, ","))
		}
	}

	// Type filters
	if s.typeFilter != nil {
		var types []string
		for _, t := range []parser.IssueType{parser.TypeBug, parser.TypeFeature, parser.TypeTask, parser.TypeEpic, parser.TypeChore} {
			if s.typeFilter[t] {
				types = append(types, string(t))
			}
		}
		if len(types) > 0 {
			filters = append(filters, "Type: "+strings.Join(types, ","))
		}
	}

	// Status filters
	if s.statusFilter != nil {
		var statuses []string
		for _, st := range []parser.Status{parser.StatusOpen, parser.StatusInProgress, parser.StatusBlocked, parser.StatusClosed} {
			if s.statusFilter[st] {
				statuses = append(statuses, string(st))
			}
		}
		if len(statuses) > 0 {
			filters = append(filters, "Status: "+strings.Join(statuses, ","))
		}
	}

	return strings.Join(filters, " | ")
}
