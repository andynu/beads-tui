# beads-tui

A powerful terminal user interface for the [beads](https://github.com/steveyegge/beads) issue tracker.

> **Note:** This project showcases two things:
> 1. **[Beads](https://github.com/steveyegge/beads)** - An exceptional local-first issue tracker that uses SQLite + JSONL for storage. If you're tired of heavyweight issue trackers and want something fast, git-friendly, and developer-focused, check out beads!
> 2. **AI-Assisted Development** - This TUI is developed primarily by guiding [Claude Code](https://code.claude.com), demonstrating how AI pair programming can build complex, maintainable software. The recent refactoring reduced main.go from 2687 to 905 lines (66%) through iterative collaboration.

## Features

### Core Functionality
- **Live monitoring** of `.beads/beads.db` SQLite database with automatic refresh
- **Dual view modes** - List view (grouped by status) and Tree view (dependency hierarchy)
- **Issue segregation** - Separate views for ready, blocked, and in-progress issues
- **Vim-style navigation** - j/k for movement, gg/G for jumps, familiar keybindings
- **Rich detail panel** - Full issue metadata, dependencies, comments, and acceptance criteria
- **Real-time updates** - Automatically refreshes when database changes

### Editing & Management
- **Full CRUD operations** - Create, edit, close, and reopen issues without leaving the TUI
- **Full-field editing** - Edit title, description, design, acceptance criteria, and notes via modal dialog (`e` key)
- **Quick updates** - Instant priority (0-4) and status (s) changes with single keypress
- **Comment system** - Add comments to issues directly from the TUI
- **Dependency management** - Add/remove blocks, parent-child, and related dependencies via dialog
- **Label management** - Add/remove labels through dedicated dialog interface
- **Clipboard integration** - Yank issue IDs (y) or IDs with titles (Y) to clipboard

### Advanced Features
- **Statistics dashboard** - Press S to view issue distribution, priority breakdown, and completion metrics
- **Advanced filtering** - Filter by priority (p0-p4), type (bug, feature, task, epic, chore), status, or labels
- **Search functionality** - Full-text search with n/N navigation through results
- **Panel focus system** - Tab between issue list and detail panel with keyboard scrolling support
- **Mouse mode toggle** - Enable/disable mouse interaction (m key) for terminal text selection
- **Natural language detection** - Automatically detects priority and type keywords when creating issues

### Visual Design
- **Color-coded priorities** - Visual indicators for P0 (critical) through P4 (lowest)
- **Status icons** - ● (ready), ○ (blocked), ◆ (in-progress), · (closed)
- **Type emoji** - 🐛 (bug), ✨ (feature), 📋 (task), 🎯 (epic), 🔧 (chore)
- **Syntax highlighting** - Color-coded dependencies, labels, and metadata
- **Responsive layout** - Adapts to terminal size with graceful degradation

## Installation

Build from source:

```bash
git clone https://github.com/andy/beads-tui  # Update with actual repo URL
cd beads-tui
go build -o beads-tui ./cmd/beads-tui
```

Or run directly:

```bash
go run ./cmd/beads-tui
```

**Note:** This project is not yet published to a package registry. Install from source as shown above.

## Usage

Navigate to a directory containing a `.beads` folder and run:

```bash
./beads-tui
```

The TUI will automatically find the `.beads/beads.db` database in the current or parent directories.

### Debug Mode

Run with comprehensive diagnostic logging:

```bash
./beads-tui --debug
# Logs saved to: ~/.beads-tui/debug-YYYY-MM-DD-HH-MM-SS.log
```

Debug logs include keyboard events, refresh operations, bd command executions, and timing information - useful for diagnosing hangs or performance issues.

## Keyboard Shortcuts

### Navigation
- `j` / `↓` - Move down
- `k` / `↑` - Move up
- `gg` - Jump to top
- `G` - Jump to bottom
- `Tab` - Focus detail panel for scrolling
- `Enter` - Focus detail panel (when on issue)
- `ESC` - Return focus to issue list

### Search
- `/` - Start search mode
- `n` - Next search result
- `N` - Previous search result
- `ESC` - Exit search mode

### Quick Actions
- `0-4` - Set priority (P0=critical, P1=high, P2=normal, P3=low, P4=lowest)
- `s` - Cycle status (open → in_progress → blocked → closed → open)
- `R` - Rename issue (edit title)
- `a` - Create new issue (vim-style "add")
- `c` - Add comment to selected issue
- `e` - Edit issue (title, description, design, acceptance, notes, priority, type)
- `x` - Close issue with optional reason
- `X` - Reopen closed issue with optional reason
- `D` - Manage dependencies (add/remove blocks, parent-child, related)
- `L` - Manage labels (add/remove labels)
- `y` - Yank (copy) issue ID to clipboard
- `Y` - Yank (copy) issue ID with title to clipboard
- `B` - Copy git branch name to clipboard

### Two-Character Shortcuts
- `So` - Set status to open
- `Si` - Set status to in_progress
- `Sb` - Set status to blocked
- `Sc` - Set status to closed

### View Controls
- `t` - Toggle between list and tree view
- `C` - Toggle showing closed issues in list view
- `f` - Quick filter (type: `p1 bug`, `feature`, etc.)
- `S` - Show statistics dashboard
- `m` - Toggle mouse mode on/off
- `r` - Manual refresh

### Detail Panel Scrolling (when focused)
- `Ctrl-d` - Scroll down half page
- `Ctrl-u` - Scroll up half page
- `Ctrl-f` - Scroll down full page (vim)
- `Ctrl-b` - Scroll up full page (vim)
- `Ctrl-e` - Scroll down one line
- `Ctrl-y` - Scroll up one line
- `PageDown` - Scroll down full page
- `PageUp` - Scroll up full page
- `Home` - Jump to top of details
- `End` - Jump to bottom of details

### General
- `?` - Show help screen
- `q` - Quit

## Quick Filter Syntax

The filter dialog (`f` key) supports natural language filtering:

```
p0-p4          Priority (e.g., 'p1' or 'p1,p2')
bug, feature, task, epic, chore    Types
open, in_progress, blocked, closed    Statuses
#label         Label (e.g., '#ui' or '#bug,#urgent')
```

**Examples:**
- `p1 bug` - P1 bugs only
- `feature,task` - Features and tasks
- `p0,p1 open` - High priority open issues
- `#ui #urgent` - Issues with 'ui' or 'urgent' labels

Leave empty to clear all filters.

## Status Indicators

- ● (green) - Ready to work on
- ○ (yellow) - Blocked by dependencies
- ◆ (blue) - In progress
- · (gray) - Closed

## Priority Colors

- `[P0]` - Critical (red)
- `[P1]` - High (orange)
- `[P2]` - Normal (light blue)
- `[P3]` - Low (gray)
- `[P4]` - Lowest (gray)

## Type Emoji

- 🐛 - Bug
- ✨ - Feature
- 📋 - Task
- 🎯 - Epic
- 🔧 - Chore

## Project Structure

```
beads-tui/
├── cmd/beads-tui/       # Main application and dialog components
│   ├── main.go          # TUI app, layout, keybindings, event loop
│   └── dialogs.go       # Modal dialogs for create/edit/dependencies/labels
├── internal/
│   ├── app/             # Application context and initialization
│   ├── formatting/      # Color schemes, status formatting, detail rendering
│   ├── parser/          # JSONL parser for beads issues (legacy support)
│   ├── state/           # Issue categorization and filtering logic
│   ├── storage/         # SQLite database reader (primary data source)
│   ├── ui/              # UI components and rendering helpers
│   └── watcher/         # Filesystem monitoring with debouncing
├── beads/               # Vendored beads project (full)
└── go.mod
```

## Dependencies

- [tview](https://github.com/rivo/tview) - Terminal UI framework
- [tcell](https://github.com/gdamore/tcell) - Low-level terminal control
- [fsnotify](https://github.com/fsnotify/fsnotify) - Filesystem monitoring
- [go-sqlite3](https://github.com/ncruces/go-sqlite3) - SQLite database driver
- [clipboard](https://github.com/atotto/clipboard) - Cross-platform clipboard access

## Troubleshooting

### No issues displayed

The TUI reads directly from the SQLite database (`.beads/beads.db`). If no issues appear:

1. **Check database exists:**
   ```bash
   ls -la .beads/beads.db
   ```

2. **Verify issues exist:**
   ```bash
   bd list
   ```

3. **Force a refresh:**
   - Press `r` in the TUI for manual refresh
   - Check debug logs with `--debug` flag

### TUI not updating after bd commands

The TUI watches the SQLite database file for changes. Updates should appear within ~200ms.

**If updates don't appear:**
1. Check file watcher is running (no errors on startup)
2. Force manual refresh with `r` key
3. Run with `--debug` to check watcher events

### File not found error

Ensure you're in a directory with a `.beads` folder:

```bash
bd init --quiet  # Initialize beads if needed
```

### Editor integration

The TUI uses built-in text areas for editing. Press `e` to open the edit dialog with fields for title, description, design, acceptance criteria, and notes.

> **Note:** External `$EDITOR` integration is planned but not yet implemented. See `tui-qxy.1` in the issue tracker.

## Development

This project uses [beads](https://github.com/steveyegge/beads) for its own issue tracking (dogfooding). View current work:

```bash
bd ready        # Show ready issues
bd list         # Show all issues
bd dep tree     # Show dependency tree
bd show tui-xyz # Show issue details
bd comments tui-xyz # Show issue comments (separate from bd show!)
```

**IMPORTANT:** `bd show` does NOT include comments. Always run `bd comments` separately when examining an issue.

### Running Tests

```bash
go test ./...                          # All tests
go test -coverprofile=coverage.out ./... # With coverage
go tool cover -html=coverage.out       # View coverage report
```

### Development Commands

```bash
# Build
go build -o beads-tui ./cmd/beads-tui

# Run with debug logging
./beads-tui --debug

# Create new issue
bd create "Issue title" -p 1 -t feature

# Update issue
bd update tui-xyz --status in_progress

# Add comment
bd comment tui-xyz "Your comment text"
```

## Architecture

### Data Flow

**Startup:**
1. Find `.beads` directory (current dir or walk up parent dirs)
2. Open SQLite database at `.beads/beads.db`
3. Load issues and categorize (ready/blocked/in-progress/closed)
4. Build tview UI with populated lists
5. Start fsnotify watcher on database file
6. Display TUI

**Live updates:**
1. User runs `bd` command (e.g., `bd create`, `bd update`)
2. bd writes to SQLite database
3. fsnotify detects database write
4. Watcher debounces (200ms) and triggers refresh
5. Re-query database, update state, redraw UI
6. TUI updates automatically

**Issue categorization:**
- **Ready:** Open issues with no open blocking dependencies
- **Blocked:** Open issues with unresolved blocking dependencies OR status="blocked"
- **In Progress:** Issues with status="in_progress"
- **Closed:** Issues with status="closed" (hidden by default, toggle with C)

### Package Responsibilities

**`cmd/beads-tui/`** - Main application
- `main.go`: TUI layout, keybindings, event loop, issue list rendering
- `dialogs.go`: Modal dialogs for create/edit/dependencies/labels/help

**`internal/app/`** - Application context
- Initialization and application-wide state

**`internal/formatting/`** - Presentation logic
- Color schemes for priority/status/type
- Detail panel formatting
- Status icon and emoji rendering

**`internal/parser/`** - JSONL parsing (legacy)
- Domain types matching beads schema
- Line-by-line JSONL reader (kept for backward compatibility)

**`internal/state/`** - Business logic
- Issue categorization (ready/blocked/in-progress/closed)
- Dependency graph building
- Filter and search logic
- Tree view structure building

**`internal/storage/`** - Data access
- SQLite database reading (primary data source)
- Query construction for issues, dependencies, comments

**`internal/ui/`** - UI helpers
- Component builders
- Rendering utilities

**`internal/watcher/`** - File monitoring
- fsnotify wrapper with 200ms debouncing
- Triggers refresh callback on database writes

## Integration with Beads

This TUI provides full CRUD operations for beads issues. All modifications use the `bd` command under the hood:

```bash
# All these work from within the TUI via dialogs/keybindings
bd create "New issue" -p 1
bd update bd-a1b2 --status in_progress
bd comment bd-a1b2 "Added comment"
bd dep add bd-a1b2 bd-xyz3 --type blocks
bd label bd-a1b2 urgent
bd close bd-a1b2

# Changes appear in TUI within ~200ms
```

The TUI is both a viewer and a full-featured editor for beads issues.

## License

MIT
