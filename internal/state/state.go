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
	viewMode         ViewMode
	treeNodes        []*TreeNode
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
