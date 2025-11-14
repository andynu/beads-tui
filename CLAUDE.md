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

# Show issue details (metadata only)
bd show <issue-id>

# Show issue comments (separate command!)
bd comments <issue-id>

# IMPORTANT: Always check BOTH when examining an issue
# bd show does NOT include comments - you must run bd comments separately

# Create new issue
bd create "Issue title" -p 1 -t feature

# Add a comment to an issue
bd comment <issue-id> "Your comment text"
# Or: bd comments add <issue-id> "Your comment text"

# Update issue status
bd update <issue-id> --status in_progress

# Force export to JSONL (if TUI shows no issues)
bd export -o .beads/issues.jsonl
```

**Note:** The beads daemon uses a 30-second debounce before auto-exporting. If the TUI shows no issues immediately after creating them, either wait ~30 seconds or run `bd export` manually.

### Debug Logging

The TUI includes comprehensive diagnostic logging to help diagnose hangs and performance issues:

```bash
# Run with debug logging enabled
./beads-tui --debug

# Log file location
~/.beads-tui/debug-YYYY-MM-DD-HH-MM-SS.log
```

**What gets logged:**
- All keyboard input events (key, rune, modifiers, current mode)
- Issue refresh operations (start, database load, UI update, completion)
- bd command executions (priority/status changes)
- File watcher events (changes detected, refresh triggers)
- Application lifecycle (startup, shutdown, errors)
- Timestamps with microsecond precision
- Source file and line numbers

**Use cases:**
- **TUI hangs**: Examine log to see last keyboard event and any stuck operations
- **Performance issues**: Check refresh timing and database load duration
- **Command failures**: See exact bd commands executed and error messages
- **Watcher problems**: Verify file changes are detected

**Example log output:**
```
2025/11/14 09:45:23.123456 main.go:42: === beads-tui started in debug mode ===
2025/11/14 09:45:23.234567 main.go:53: Finding .beads directory
2025/11/14 09:45:23.345678 main.go:301: Setting up file watcher on: /path/.beads/beads.db
2025/11/14 09:45:23.456789 main.go:315: WATCHER: File watcher started successfully
2025/11/14 09:45:30.123456 main.go:412: KEY EVENT: key=Rune rune='j' mod=0 searchMode=false detailFocus=false
2025/11/14 09:45:31.234567 main.go:412: KEY EVENT: key=Rune rune='1' mod=0 searchMode=false detailFocus=false
2025/11/14 09:45:31.345678 main.go:577: BD COMMAND: Executing priority update: bd update tui-123 --priority 1
2025/11/14 09:45:31.456789 main.go:583: BD COMMAND: Priority update successful for tui-123 -> P1
```

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

**Panel focus system:** The TUI has two focusable panels (issue list and detail panel):
- **Default:** Issue list is focused (for navigation)
- **Tab or Enter:** Focus detail panel to enable keyboard scrolling
- **ESC:** Return focus to issue list
- **Visual indicators:** Focused panel has yellow border and updated title text
- **Status bar:** Shows current focus (List or Details)

**Quick issue updates:** Issues can be updated directly from the TUI without leaving the interface:
- **Priority changes:** Press `0`-`4` to instantly set priority (P0=critical, P1=high, P2=normal, P3=low, P4=lowest)
- **Status cycling:** Press `s` to cycle through statuses (open → in_progress → blocked → closed → open)
- **Immediate feedback:** Confirmation message shown in status bar with green checkmark
- **Auto-refresh:** Issue list automatically refreshes 500ms after update to show changes
- **No dialogs:** Updates are instant with single keypress, no confirmation needed

**Issue creation dialog:** Create new issues directly from the TUI with the `a` key (tui-ywv.6):
- **Modal form:** Centered dialog with title, description, priority, type fields
- **Smart defaults:** Pre-selects P2 priority and feature type
- **Parent relationship:** Optional checkbox to add as child of currently selected issue
- **Form navigation:** Tab between fields, Enter/Ctrl-S to submit, ESC to cancel
- **Validation:** Requires title before allowing creation
- **Command execution:** Executes `bd create` command with form data
- **Auto-refresh:** Issue list refreshes 500ms after creation to show new issue

## Editing UX Design (tui-qxy.1)

This section documents the interaction patterns for editing issues in the TUI. The design follows vim-style conventions and reuses patterns from existing features.

### Design Principles

1. **Consistency with existing patterns:** Follow vim-style keybindings (j/k, gg/G, etc.)
2. **Minimal context switching:** Use $EDITOR for text fields, modal dialogs for structured data
3. **Safe defaults:** Require explicit confirmation for destructive operations
4. **Immediate feedback:** Show status bar messages for all operations
5. **Integration via bd CLI:** Use `bd` commands rather than direct DB writes (maintains single source of truth)

### Keybinding Allocation

**Current keybindings (reserved, DO NOT conflict):**
- Navigation: `j`, `k`, `g`, `G`, `Tab`, `Enter`, `ESC`, arrows
- Search: `/`, `n`, `N`
- Quick updates: `0`-`4` (priority), `s` (status)
- View controls: `t` (tree/list), `f` (filters), `m` (mouse), `r` (refresh)
- Creation: `a` (add issue), `c` (add comment)
- Help: `?`
- Quit: `q`
- Detail panel scrolling: `Ctrl-d`, `Ctrl-u`, `Ctrl-e`, `Ctrl-y`, `PageDown`, `PageUp`, `Home`, `End`

**New keybindings for editing:**
- `e` - **Edit menu** (shows modal with options: Description / Design / Acceptance / Notes / Cancel)
- `D` - **Manage dependencies** (modal dialog for adding/removing blocks and parent-child relationships)
- `L` - **Manage labels** (modal dialog for adding/removing labels)
- `y` - **Yank issue ID to clipboard** (vim-style, copies current issue ID)
- `Y` - **Yank issue ID with title** (copies "tui-xyz - Issue title" format)

**Existing but enhanced:**
- `c` - Add comment (already implemented, enhancing to support $EDITOR with template)

### Text Field Editing with $EDITOR

**Keybinding:** `e`

**Behavior:**
1. Press `e` on a selected issue
2. Show modal menu with options:
   - Edit Description
   - Edit Design
   - Edit Acceptance Criteria
   - Edit Notes
   - Cancel
3. On selection:
   - Write current field value to temp file (`/tmp/beads-tui-<field>-<issue-id>.md`)
   - Pause TUI with `app.Suspend()`
   - Spawn `$EDITOR` (fall back to `vim` if unset)
   - Wait for editor to close
   - Read temp file content
   - Call `bd update <id> --description <content>` (or `--design`, `--acceptance`, `--notes`)
   - Clean up temp file
   - Resume TUI with `app.Draw()`
   - Show confirmation in status bar: `[green]✓ Updated description for <issue-id>[-]`
   - Trigger refresh after 500ms (preserves selection)

**Editor command:** `$EDITOR /tmp/beads-tui-description-tui-xyz.md` (or use `vim` if `$EDITOR` unset)

**Error handling:**
- If `$EDITOR` executable not found: Show error in status bar, fall back to `vim`
- If editor exits with non-zero code: Show error, ask "Save changes anyway? (y/n)"
- If content is empty after editing: Ask "Clear this field? (y/n)"
- If `bd update` fails: Show error message with bd output

**Template format for empty fields:**
```markdown
# Edit <field-name> for <issue-id>
# Lines starting with # are ignored
# Save and exit to update, or exit without saving to cancel

<existing content here, or empty>
```

**Implementation notes:**
- Use `os/exec` to spawn editor
- Strip comment lines (starting with `#`) before sending to bd
- Properly escape content for shell (use heredoc or write to temp file and pass path)
- Preserve cursor position in detail panel after refresh

### Comment Creation with $EDITOR

**Keybinding:** `c` (already implemented, enhancing)

**Current behavior:** Shows modal dialog with textarea

**Enhanced behavior (optional improvement for consistency):**
1. Press `c` on a selected issue
2. Spawn `$EDITOR` with template file
3. Read comment after editor closes
4. Call `bd comment <id> "<comment-text>"`

**Template format:**
```markdown
# Add comment to <issue-id> - <title>
# Lines starting with # are ignored
# Save and exit to post comment, or exit without saving to cancel

Your comment here...
```

**Note:** This is optional - the current modal dialog implementation works well. Consider this if users prefer $EDITOR workflow.

### Dependency Management Dialog

**Keybinding:** `D`

**Behavior:**
1. Press `D` on a selected issue
2. Show modal dialog with two sections:
   - **Current dependencies** (list with remove buttons)
   - **Add new dependency** (input field + type selector)

**Dialog layout:**
```
┌─────────────────────────────────────────────────┐
│         Manage Dependencies for tui-xyz         │
├─────────────────────────────────────────────────┤
│ Current Dependencies:                           │
│   [x] blocks     tui-abc  [Remove]              │
│   [x] parent-child tui-def [Remove]             │
│                                                  │
│ Add New Dependency:                             │
│   Issue ID: [____________]                      │
│   Type: [blocks ▼] [parent-child] [related]     │
│                    [discovered-from]            │
│                                                  │
│   [Add]  [Close]                                │
└─────────────────────────────────────────────────┘
```

**Operations:**
- **Add:** Call `bd dep add <issue-id> <target-id> --type <type>`
- **Remove:** Call `bd dep remove <issue-id> <target-id> --type <type>`
- Validate that target issue exists (search in current state)
- Show error if target not found: `[red]Issue <id> not found[-]`
- Trigger refresh after each add/remove operation

**Implementation notes:**
- Use `tview.Form` for the dialog
- List dependencies from `issue.Dependencies`
- Use checkboxes or list items with remove buttons
- Add autocomplete for issue IDs (optional enhancement)
- Show issue titles next to IDs for clarity

### Label Management Dialog

**Keybinding:** `L`

**Behavior:**
1. Press `L` on a selected issue
2. Show modal dialog with current labels and add/remove interface

**Dialog layout:**
```
┌─────────────────────────────────────────────────┐
│         Manage Labels for tui-xyz               │
├─────────────────────────────────────────────────┤
│ Current Labels:                                 │
│   [ui] [x]    [bug] [x]    [urgent] [x]        │
│                                                  │
│ Add Label:                                       │
│   [____________]  [Add]                         │
│                                                  │
│   [Close]                                        │
└─────────────────────────────────────────────────┘
```

**Operations:**
- **Add:** Call `bd label <issue-id> <label-name>`
- **Remove:** Call `bd label <issue-id> --remove <label-name>`
- Trim whitespace from label names
- Prevent duplicate labels (check before adding)
- Show available labels from other issues (optional autocomplete)

**Implementation notes:**
- Use `tview.Form` for the dialog
- Display labels as chips/tags with [x] buttons
- Input field for new label name
- Consider color-coding labels (enhancement)

### Clipboard Integration

**Keybindings:** `y` (yank ID), `Y` (yank ID with title)

**Behavior:**
- Press `y` on selected issue: Copy issue ID to clipboard (e.g., `tui-xyz`)
- Press `Y` on selected issue: Copy ID with title (e.g., `tui-xyz - Build beads-tui`)
- Show confirmation: `[green]✓ Copied tui-xyz to clipboard[-]`
- Clear message after 2 seconds

**Use cases:**
- Quick reference for adding dependencies
- Pasting into commit messages
- Sharing issue IDs in chat/docs

**Implementation:**
- Use `github.com/atotto/clipboard` library (already imported)
- Call `clipboard.WriteAll(issue.ID)` or `clipboard.WriteAll(fmt.Sprintf("%s - %s", issue.ID, issue.Title))`
- Handle errors gracefully (show error in status bar if clipboard unavailable)
- Cross-platform support (Linux/xclip, macOS/pbcopy, Windows/clip)

**Note:** Clipboard already works for clicking issue ID in detail panel (line 384-415), extend this to keybindings

### Implementation Order

1. **tui-qxy.1** (this task) - Document UX design ✓
2. **tui-qxy.2** - Implement text field editor with $EDITOR (highest priority)
3. **tui-qxy.3** - Enhance comment creation with $EDITOR (optional)
4. **tui-qxy.4** - Add dependency management dialog
5. **tui-qxy.5** - Add label management dialog
6. **tui-qxy.7** - Add clipboard yank keybindings
7. **tui-qxy.8** - Write comprehensive tests

### Success Criteria

- Users can edit all text fields without leaving TUI
- $EDITOR integration works with common editors (vim, nano, emacs, vscode --wait)
- TUI properly suspends/resumes without corruption
- All bd commands execute successfully with proper error handling
- Status bar provides clear feedback for all operations
- Selection is preserved after updates
- No keybinding conflicts with existing features

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

**Issue List Mode:**
- `q`: Quit
- `r`: Manual refresh
- `j`: Down (simulates arrow key)
- `k`: Up (simulates arrow key)
- `t`: Toggle between list and tree view
- `m`: Toggle mouse mode on/off
- `a`: Open issue creation dialog (vim-style "add")
- `0`-`4`: Quick priority change (set current issue to P0-P4)
- `s`: Toggle status (cycles: open → in_progress → blocked → closed → open)
- `g` + `g`: Jump to top
- `G`: Jump to bottom
- `/`: Start search mode
- `n`: Next search result
- `N`: Previous search result
- `Tab`: Focus detail panel for scrolling
- `Enter`: Focus detail panel (when on an issue)

**Detail Panel Mode (when focused):**
- `ESC`: Return focus to issue list
- `Ctrl-d`: Scroll down half page (vim style)
- `Ctrl-u`: Scroll up half page (vim style)
- `Ctrl-e`: Scroll down one line (vim style)
- `Ctrl-y`: Scroll up one line (vim style)
- `PageDown`: Scroll down full page
- `PageUp`: Scroll up full page
- `Home`: Jump to top of details
- `End`: Jump to bottom of details

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
