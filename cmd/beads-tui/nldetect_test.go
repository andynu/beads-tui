package main

import (
	"strings"
	"testing"
)

// detectPriority detects priority from text using natural language keywords
func detectPriority(text string) *int {
	lower := strings.ToLower(text)
	// P0 keywords: critical, urgent, blocking, blocker, emergency, outage, down, broken
	if strings.Contains(lower, "critical") || strings.Contains(lower, "urgent") ||
		strings.Contains(lower, "blocking") || strings.Contains(lower, "blocker") ||
		strings.Contains(lower, "emergency") || strings.Contains(lower, "outage") ||
		strings.Contains(lower, "down") || strings.Contains(lower, "broken") {
		p := 0
		return &p
	}
	// P1 keywords: important, high priority, asap, soon, needed
	if strings.Contains(lower, "important") || strings.Contains(lower, "high priority") ||
		strings.Contains(lower, "asap") || strings.Contains(lower, "soon") ||
		strings.Contains(lower, "needed") || strings.Contains(lower, "must have") {
		p := 1
		return &p
	}
	// P3 keywords: low priority, minor, nice to have, eventually, someday
	if strings.Contains(lower, "low priority") || strings.Contains(lower, "minor") ||
		strings.Contains(lower, "nice to have") || strings.Contains(lower, "eventually") ||
		strings.Contains(lower, "someday") || strings.Contains(lower, "polish") {
		p := 3
		return &p
	}
	// P4 keywords: trivial, cosmetic, optional
	if strings.Contains(lower, "trivial") || strings.Contains(lower, "cosmetic") ||
		strings.Contains(lower, "optional") {
		p := 4
		return &p
	}
	return nil // No match, keep default
}

// detectIssueType detects issue type from text using natural language keywords
func detectIssueType(text string) *string {
	lower := strings.ToLower(text)
	// Bug keywords: bug, error, crash, fix, broken, issue, problem, regression
	if strings.Contains(lower, "bug") || strings.Contains(lower, "error") ||
		strings.Contains(lower, "crash") || strings.Contains(lower, "fix ") ||
		strings.Contains(lower, "broken") || strings.Contains(lower, "problem") ||
		strings.Contains(lower, "regression") {
		t := "bug"
		return &t
	}
	// Epic keywords: epic, project, initiative, milestone (check before task)
	if strings.Contains(lower, "epic") || strings.Contains(lower, "project") ||
		strings.Contains(lower, "initiative") || strings.Contains(lower, "milestone") {
		t := "epic"
		return &t
	}
	// Chore keywords: chore, maintenance, dependency, upgrade, cleanup (check before task)
	if strings.Contains(lower, "chore") || strings.Contains(lower, "maintenance") ||
		strings.Contains(lower, "dependency") || strings.Contains(lower, "upgrade") ||
		strings.Contains(lower, "cleanup") {
		t := "chore"
		return &t
	}
	// Task keywords: task, do, implement, update, change, refactor, clean up
	if strings.Contains(lower, "task") || strings.Contains(lower, "do ") ||
		strings.Contains(lower, "implement") || strings.Contains(lower, "update") ||
		strings.Contains(lower, "change") || strings.Contains(lower, "refactor") ||
		strings.Contains(lower, "clean up") {
		t := "task"
		return &t
	}
	// Feature is default, so only explicitly detect if keywords present
	if strings.Contains(lower, "feature") || strings.Contains(lower, "add ") ||
		strings.Contains(lower, "new ") || strings.Contains(lower, "build") ||
		strings.Contains(lower, "create") {
		t := "feature"
		return &t
	}
	return nil // No match, keep default
}

func TestDetectPriority(t *testing.T) {
	tests := []struct {
		text     string
		expected *int
	}{
		{"Critical bug in production", intPtr(0)},
		{"Urgent fix needed", intPtr(0)},
		{"This is blocking the release", intPtr(0)},
		{"Emergency outage", intPtr(0)},
		{"System is down", intPtr(0)},
		{"Important feature", intPtr(1)},
		{"High priority task", intPtr(1)},
		{"Need this ASAP", intPtr(1)},
		{"Must have feature", intPtr(1)},
		{"Minor issue", intPtr(3)},
		{"Low priority enhancement", intPtr(3)},
		{"Nice to have feature", intPtr(3)},
		{"Polish the UI", intPtr(3)},
		{"Trivial change", intPtr(4)},
		{"Cosmetic fix", intPtr(4)},
		{"Optional enhancement", intPtr(4)},
		{"Normal feature request", nil}, // No keywords, should return nil
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			result := detectPriority(tt.text)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %d", *result)
				}
			} else {
				if result == nil {
					t.Errorf("Expected %d, got nil", *tt.expected)
				} else if *result != *tt.expected {
					t.Errorf("Expected %d, got %d", *tt.expected, *result)
				}
			}
		})
	}
}

func TestDetectIssueType(t *testing.T) {
	tests := []struct {
		text     string
		expected *string
	}{
		{"Fix the bug in login", strPtr("bug")},
		{"Error in database connection", strPtr("bug")},
		{"Crash when clicking button", strPtr("bug")},
		{"Problem with API", strPtr("bug")},
		{"Regression in authentication", strPtr("bug")},
		{"Implement user authentication", strPtr("task")},
		{"Update dependencies", strPtr("task")},
		{"Refactor database layer", strPtr("task")},
		{"Clean up old code", strPtr("task")},
		{"Build a new dashboard epic", strPtr("epic")},
		{"Project planning for Q4", strPtr("epic")},
		{"Milestone: MVP launch", strPtr("epic")},
		{"Chore: update CI config", strPtr("chore")},
		{"Maintenance work on servers", strPtr("chore")},
		{"Upgrade dependencies", strPtr("chore")},
		{"Add new feature for users", strPtr("feature")},
		{"Build login system", strPtr("feature")},
		{"Create user profile page", strPtr("feature")},
		{"Something else entirely", nil}, // No keywords, should return nil
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			result := detectIssueType(tt.text)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %s", *result)
				}
			} else {
				if result == nil {
					t.Errorf("Expected %s, got nil", *tt.expected)
				} else if *result != *tt.expected {
					t.Errorf("Expected %s, got %s", *tt.expected, *result)
				}
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}

func strPtr(s string) *string {
	return &s
}
