# beads-tui

A terminal user interface for the [beads](https://github.com/steveyegge/beads) issue tracker.

> **Note:** This project showcases two things:
> 1. **[Beads](https://github.com/steveyegge/beads)** - An exceptional local-first issue tracker that uses SQLite + JSONL for storage. If you're tired of heavyweight issue trackers and want something fast, git-friendly, and developer-focused, check out beads!
> 2. **AI-Assisted Development** - This TUI is developed primarily by guiding [Claude Code](https://code.claude.com), demonstrating how AI pair programming can build complex, maintainable software. The recent refactoring reduced main.go from 2687 to 905 lines (66%) through iterative collaboration.

## Features

- **Live monitoring** of `.beads/issues.jsonl` with filesystem watching
- **Issue segregation** - separate views for ready, blocked, and in-progress issues
- **Vim-style navigation** - j/k for movement, familiar keybindings
- **Rich detail panel** - view full issue metadata, dependencies, and comments
- **Color-coded priorities** - visual indicators for P0-P4 issues
- **Real-time updates** - automatically refreshes when JSONL file changes

## Installation

```bash
go install github.com/andy/beads-tui/cmd/beads-tui@latest
```

Or build from source:

```bash
git clone https://github.com/andy/beads-tui
cd beads-tui
go build -o beads-tui ./cmd/beads-tui
```

## Usage

Navigate to a directory containing a `.beads` folder and run:

```bash
beads-tui
```

The TUI will automatically find the `.beads/issues.jsonl` file in the current or parent directories.

## Keyboard Navigation

### Movement
- `j` / `↓` - Move down
- `k` / `↑` - Move up
- `Enter` - Select issue (show details)
- `q` - Quit

### Coming Soon
- `gg` - Jump to top
- `G` - Jump to bottom
- `/` - Search
- `f` - Filter
- `1-4` - Quick filters (ready/blocked/in-progress/all)

## Status Indicators

- `●` (green) - Ready to work on
- `○` (yellow) - Blocked by dependencies
- `◆` (blue) - In progress
- `✓` (gray) - Closed

## Priority Colors

- `[P0]` - Critical (red)
- `[P1]` - High (orange)
- `[P2]` - Medium (white)
- `[P3]` - Low (gray)
- `[P4]` - Backlog (dark gray)

## Project Structure

```
beads-tui/
├── cmd/beads-tui/       # Main application entry point
├── internal/
│   ├── parser/          # JSONL parser for beads issues
│   ├── state/           # Application state management
│   ├── ui/              # TUI components (coming soon)
│   └── watcher/         # Filesystem monitoring (coming soon)
├── go.mod
└── README.md
```

## Dependencies

- [tview](https://github.com/rivo/tview) - Terminal UI framework
- [tcell](https://github.com/gdamore/tcell) - Low-level terminal control
- [fsnotify](https://github.com/fsnotify/fsnotify) - Filesystem monitoring

## Troubleshooting

### No issues displayed

If the TUI shows no issues but `bd ready` shows issues, the JSONL file hasn't been exported yet. The beads daemon uses a 30-second debounce before auto-exporting.

**Quick fix:**
```bash
bd export -o .beads/issues.jsonl
```

Or wait ~30 seconds after creating/modifying issues for auto-export.

### File not found error

Ensure you're in a directory with a `.beads` folder:
```bash
bd init --quiet  # Initialize beads if needed
```

## Development

This project uses [beads](https://github.com/steveyegge/beads) for issue tracking. View the current work:

```bash
bd ready        # Show ready issues
bd list         # Show all issues
bd dep tree     # Show dependency tree
```

## License

MIT
