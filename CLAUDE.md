# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**beads-tui** is a terminal user interface for the [beads](https://github.com/steveyegge/beads) issue tracker. It provides real-time monitoring of `.beads/issues.jsonl` with vim-style navigation and displays issues segregated by ready/blocked/in-progress status.

This project uses beads for its own issue tracking (dogfooding). The `beads/` subdirectory contains a vendored copy of the full beads project.

## Development Commands

### Build and Run

```bash
# Build the TUI
go build -o beads-tui ./cmd/beads-tui

# Run in current directory (finds .beads automatically)
./beads-tui

# Run tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Working with Beads Issues

```bash
# View ready issues
bd ready

# Show all issues
bd list

# View dependency tree
bd dep tree

# Create new issue
bd create "Issue title" -p 1 -t feature

# Update issue status
bd update <issue-id> --status in_progress

# Force export to JSONL (if TUI shows no issues)
bd export -o .beads/issues.jsonl
```

**Note:** The beads daemon uses a 30-second debounce before auto-exporting. If the TUI shows no issues immediately after creating them, either wait ~30 seconds or run `bd export` manually.

### Testing Considerations

**CRITICAL:** Never pollute the production `.beads/beads.db` with test data. When manually testing, use:

```bash
BEADS_DB=/tmp/test.db ./bd init --quiet --prefix test
BEADS_DB=/tmp/test.db ./bd create "Test issue" -p 1
```

## Architecture

### High-Level Structure

```
beads-tui/
├── cmd/beads-tui/       # Main application entry point
│   └── main.go          # TUI app, layout, key bindings
├── internal/
│   ├── parser/          # JSONL parser for beads issues
│   │   ├── types.go     # Issue, Dependency, Comment types
│   │   └── parser.go    # Line-by-line JSONL reader
│   ├── state/           # Application state management
│   │   └── state.go     # Issue categorization logic
│   └── watcher/         # Filesystem monitoring
│       └── watcher.go   # fsnotify wrapper with debouncing
├── beads/               # Vendored beads project (full)
└── go.mod
```

### Key Concepts

**JSONL as data source:** The TUI reads directly from `.beads/issues.jsonl` (not the SQLite database). This file is the git-committed source of truth maintained by the beads daemon.

**Issue categorization:** The `state` package implements the core logic for segregating issues:
- **Ready:** Open issues with no open blocking dependencies
- **Blocked:** Open issues with unresolved blocking dependencies OR status="blocked"
- **In Progress:** Issues with status="in_progress"
- **Closed:** Issues with status="closed" (not displayed in main view)

**Filesystem watching:** Uses fsnotify with 200ms debouncing to detect JSONL updates and refresh the UI automatically.

**Vim-style navigation:** j/k for up/down, q to quit. Details auto-show on navigation (tui-p62).

**Mouse mode disabled:** Mouse is intentionally disabled to allow terminal text selection (see cmd/beads-tui/main.go:212).

**View modes:** The TUI supports two view modes:
- **List View (default):** Issues grouped by status (ready/blocked/in-progress)
- **Tree View:** Issues displayed as dependency hierarchy with ASCII tree characters
- Toggle between modes with the 't' key

### Package Responsibilities

**`cmd/beads-tui/main.go`** - TUI application
- tview-based UI with status bar, issue list, and detail panel
- Vim keybindings (j/k for navigation)
- Auto-show details on list navigation change
- Layout: 3-pane (status bar + issue list + details)
- Color-coded priority and status indicators

**`internal/parser/`** - JSONL parsing
- `types.go`: Domain types matching beads schema (Issue, Status, IssueType, Dependency, Comment)
- `parser.go`: Line-by-line JSONL reader with error handling
- No business logic - pure parsing only

**`internal/state/`** - State management
- Categorizes issues into ready/blocked/in-progress/closed
- Builds dependency graph to detect blocking relationships
- Provides getters for filtered issue lists
- Manages view mode state (list vs tree)
- Builds tree structure from parent-child and blocks dependencies
- No UI concerns - pure business logic

**`internal/watcher/`** - File monitoring
- fsnotify wrapper with debouncing (default 200ms)
- Triggers callback on file write/create events
- Handles cleanup and shutdown

### Data Flow

**Startup:**
1. Find `.beads` directory (current dir or walk up parent dirs)
2. Parse `.beads/issues.jsonl` via `parser.ParseFile()`
3. Load into `state.State` and categorize issues
4. Build tview UI with populated lists
5. Start fsnotify watcher on JSONL file
6. Display TUI

**Live updates:**
1. User runs `bd` command (e.g., `bd create`, `bd update`)
2. bd daemon writes to SQLite, then exports to JSONL (30s debounce)
3. fsnotify detects JSONL write
4. Watcher debounces (200ms) and triggers `refreshIssues()`
5. Re-parse JSONL, update state, queue UI redraw on main thread
6. TUI updates automatically

**Issue details:**
1. User navigates with j/k
2. `SetChangedFunc` handler fires on list selection change
3. Check if selected item is issue (not header)
4. Format full details (description, design, acceptance criteria, dependencies, comments, metadata)
5. Update detail panel with scrollable text

### Dependency Detection Algorithm

From `internal/state/state.go:62-93`:

```go
// Build reverse dependency map
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
```

An issue is "blocked" if:
- It has status="blocked", OR
- It has a "blocks" dependency where the blocking issue is not closed

### Tree View Algorithm

The tree view visualizes issue dependencies as a hierarchical structure:

**Building the tree** (from `internal/state/state.go:201-288`):
1. Build maps of parent→children (from parent-child dependencies) and blocker→blocked (from blocks dependencies)
2. Find root nodes: issues with no incoming parent-child or blocks dependencies
3. Recursively build tree from roots, adding both children and blocked issues
4. Skip closed issues in tree view
5. Use cycle detection to prevent infinite recursion

**Rendering the tree** (from `cmd/beads-tui/main.go:70-119`):
- Uses ASCII tree characters: `├──`, `└──`, `│`
- Status indicators: `●` (open), `○` (blocked), `◆` (in-progress)
- Color-coded by priority and status
- Hierarchical indentation based on depth

**Example tree output:**
```
DEPENDENCY TREE
[●] tui-ywv [P1] Build beads-tui: Terminal UI for beads issue tracker
├── [◆] tui-bne [P2] Add tree view mode for issue dependency visualization
├── [●] tui-hxu [P2] Add filtering capabilities (priority, type, status)
└── [●] tui-79b [P2] Write tests for core TUI components
```

## Common Development Tasks

### Adding a New Filter or View

1. Update `state.FilterMode` enum in `internal/state/state.go`
2. Add filter logic to `state.categorizeIssues()` or new getter method
3. Update `cmd/beads-tui/main.go` to add keybinding and UI handler
4. Update status bar text to show active filter

### Adding Issue Details

1. Update `parser.Issue` struct in `internal/parser/types.go` if parsing new fields
2. Update `formatIssueDetails()` in `cmd/beads-tui/main.go:259-351`
3. Use tview color markup: `[color]text[-]` for foreground, `[::b]text[-::-]` for bold

### Changing Keybindings

Update `app.SetInputCapture()` in `cmd/beads-tui/main.go`. Current bindings:
- `q`: Quit
- `r`: Manual refresh
- `j`: Down (simulates arrow key)
- `k`: Up (simulates arrow key)
- `t`: Toggle between list and tree view
- `g` + `g`: Jump to top
- `G`: Jump to bottom
- `/`: Start search mode
- `n`: Next search result
- `N`: Previous search result
- `Enter`: Handled by tview list (selection)

### Debugging Watcher Issues

If live updates aren't working:
1. Check daemon is running: `bd daemons list`
2. Check daemon logs: `bd daemons logs . -n 100`
3. Force export: `bd export -o .beads/issues.jsonl`
4. Check fsnotify errors in TUI startup output (stderr)
5. Use manual refresh with `r` key

## Key Files to Know

- **cmd/beads-tui/main.go** - Complete TUI implementation (402 lines)
- **internal/parser/types.go** - Beads domain types (matches JSONL schema)
- **internal/state/state.go** - Issue categorization and filtering
- **internal/watcher/watcher.go** - Debounced filesystem watching
- **beads/CLAUDE.md** - Full beads project documentation

## Integration with Beads

This TUI is a read-only viewer of beads data. For modifications, use the `bd` command:

```bash
# All bd commands work normally
bd create "New issue" -p 1
bd update bd-a1b2 --status in_progress
bd close bd-a1b2

# Changes will auto-refresh in TUI (30s + 200ms latency)
```

The JSONL file format is append-only with "last write wins" semantics. The TUI always reads the latest state.

## Troubleshooting

### No issues displayed

**Cause:** JSONL file hasn't been exported yet (30s debounce).

**Fix:**
```bash
bd export -o .beads/issues.jsonl
```

### File not found error

**Cause:** Not in a directory with `.beads` folder.

**Fix:**
```bash
bd init --quiet
```

### TUI not updating after bd commands

**Cause:** Daemon not running or export debounce delay.

**Fix:**
```bash
# Check daemon
bd daemons list

# Force immediate export
bd sync
```

### Incorrect issue categorization

**Cause:** Dependency logic bug or stale JSONL.

**Debug:**
```bash
# Check what bd thinks
bd ready
bd list --status blocked

# Compare with TUI display
```
