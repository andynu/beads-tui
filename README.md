# beads-tui

A terminal user interface for the [beads](https://github.com/steveyegge/beads) issue tracker.

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

## Development

This project uses [beads](https://github.com/steveyegge/beads) for issue tracking. View the current work:

```bash
bd ready        # Show ready issues
bd list         # Show all issues
bd dep tree     # Show dependency tree
```

## License

MIT
