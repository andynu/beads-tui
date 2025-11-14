package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// Parser reads and parses JSONL files containing beads issues
type Parser struct {
	path string
}

// New creates a new parser for the given JSONL file path
func New(path string) *Parser {
	return &Parser{path: path}
}

// ParseAll reads all issues from the JSONL file
func (p *Parser) ParseAll() ([]*Issue, error) {
	file, err := os.Open(p.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var issues []*Issue
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		var issue Issue
		if err := json.Unmarshal(line, &issue); err != nil {
			return nil, fmt.Errorf("invalid JSON at line %d: %w", lineNum, err)
		}

		issues = append(issues, &issue)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return issues, nil
}

// ParseFile is a convenience function to parse a JSONL file
func ParseFile(path string) ([]*Issue, error) {
	p := New(path)
	return p.ParseAll()
}
