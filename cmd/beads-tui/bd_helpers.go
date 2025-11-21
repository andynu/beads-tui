package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/andy/beads-tui/internal/parser"
)

// BdCommandResult represents the result of executing a bd command with --json
type BdCommandResult struct {
	Issues   []parser.Issue  `json:"issues,omitempty"`
	Comments []parser.Comment `json:"comments,omitempty"`
	Error    string          `json:"error,omitempty"`
}

// execBdJSON executes a bd command with --json flag and parses the response.
// It handles both single object and array responses from bd commands.
//
// Example usage:
//   result, err := execBdJSON("update", "tui-123", "--priority", "1")
//   if err != nil { ... }
//   if len(result.Issues) > 0 {
//     updatedIssue := result.Issues[0]
//   }
func execBdJSON(args ...string) (*BdCommandResult, error) {
	// Add --json flag if not already present
	hasJSON := false
	for _, arg := range args {
		if arg == "--json" {
			hasJSON = true
			break
		}
	}
	if !hasJSON {
		args = append(args, "--json")
	}

	// Create context with timeout to prevent hanging indefinitely
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Execute command with timeout
	cmd := exec.CommandContext(ctx, "bd", args...)
	output, err := cmd.CombinedOutput()

	// Check for timeout error specifically
	if ctx.Err() == context.DeadlineExceeded {
		cmdStr := "bd " + strings.Join(args, " ")
		return nil, fmt.Errorf("bd command timed out after 10s: %s", cmdStr)
	}

	if err != nil {
		// Try to parse error from JSON output first
		var result BdCommandResult
		if jsonErr := json.Unmarshal(output, &result); jsonErr == nil && result.Error != "" {
			return nil, fmt.Errorf("bd %s failed: %s", args[0], result.Error)
		}
		// Fall back to original error with output
		outputStr := strings.TrimSpace(string(output))
		if outputStr == "" {
			return nil, fmt.Errorf("bd %s command failed: %v", args[0], err)
		}
		return nil, fmt.Errorf("bd %s failed: %s", args[0], outputStr)
	}

	// Parse JSON response
	result, parseErr := parseBdJSON(output)
	if parseErr != nil {
		// Provide helpful error with snippet of output
		outputPreview := string(output)
		if len(outputPreview) > 200 {
			outputPreview = outputPreview[:200] + "..."
		}
		return nil, fmt.Errorf("failed to parse JSON from bd %s: %v (output: %s)", args[0], parseErr, outputPreview)
	}

	return result, nil
}

// parseBdJSON parses bd command JSON output, handling multiple response formats:
// - Array of issues: [{"id":"tui-123",...}]
// - Single issue: {"id":"tui-123",...}
// - Single comment: {"id":15,"issue_id":"tui-123",...}
func parseBdJSON(data []byte) (*BdCommandResult, error) {
	result := &BdCommandResult{}

	// Try parsing as array of issues first (most common)
	var issues []parser.Issue
	if err := json.Unmarshal(data, &issues); err == nil {
		result.Issues = issues
		return result, nil
	}

	// Try parsing as single issue
	var issue parser.Issue
	if err := json.Unmarshal(data, &issue); err == nil {
		// Check if this looks like an issue (has ID and Title fields)
		if issue.ID != "" && issue.Title != "" {
			result.Issues = []parser.Issue{issue}
			return result, nil
		}
	}

	// Try parsing as single comment
	var comment parser.Comment
	if err := json.Unmarshal(data, &comment); err == nil {
		// Check if this looks like a comment (has ID and IssueID)
		if comment.ID > 0 && comment.IssueID != "" {
			result.Comments = []parser.Comment{comment}
			return result, nil
		}
	}

	return nil, fmt.Errorf("unable to parse JSON as issue array, issue, or comment")
}

// execBdJSONIssue is a convenience wrapper that executes a bd command and returns
// the first issue from the result, or an error if no issues were returned.
func execBdJSONIssue(args ...string) (*parser.Issue, error) {
	result, err := execBdJSON(args...)
	if err != nil {
		return nil, err
	}

	if len(result.Issues) == 0 {
		cmdName := "unknown"
		if len(args) > 0 {
			cmdName = args[0]
		}
		return nil, fmt.Errorf("bd %s returned no issues (expected an issue in response)", cmdName)
	}

	return &result.Issues[0], nil
}

// execBdJSONComment is a convenience wrapper that executes a bd command and returns
// the first comment from the result, or an error if no comments were returned.
func execBdJSONComment(args ...string) (*parser.Comment, error) {
	result, err := execBdJSON(args...)
	if err != nil {
		return nil, err
	}

	if len(result.Comments) == 0 {
		cmdName := "unknown"
		if len(args) > 0 {
			cmdName = args[0]
		}
		return nil, fmt.Errorf("bd %s returned no comments (expected a comment in response)", cmdName)
	}

	return &result.Comments[0], nil
}
