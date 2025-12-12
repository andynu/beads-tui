# Changelog

All notable changes to beads-tui will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2025-12-12

### Added

#### Tree View Enhancements
- **Collapsible nodes** (`o`) - Collapse/expand parent nodes in tree view (vim-style fold)
  - `[+]` indicator for collapsed nodes with child count
  - `[-]` indicator for expanded nodes
  - Leaf nodes maintain visual alignment
- **Smart collapse defaults** - Nodes auto-collapse/expand based on activity:
  - Auto-expand nodes with active work (any child is `in_progress`)
  - Auto-collapse nodes without active work (all children closed or pending)
  - Manual toggles override smart defaults
- **Collapse state persistence** - Remember collapse preferences between sessions
  - Saved to `~/.beads-tui/collapse-<hash>.json` (per-project)
  - Loads on startup, saves on toggle and exit

#### UI Improvements
- **View mode hint in panel title** - Shows current view mode (e.g., "Issues [List] (t:Tree)")
- Cleaner status bar (removed redundant view mode indicator)

## [0.1.0] - 2025-11-25

Initial release of beads-tui, a terminal user interface for the [beads](https://github.com/steveyegge/beads) issue tracker.

### Features

#### Core Functionality
- Real-time monitoring of beads database with automatic UI refresh
- Direct SQLite database reading for instant updates
- Filesystem watching with debounced refresh (200ms)
- Graceful shutdown with signal handling

#### Views
- **List View**: Issues grouped by status (ready/blocked/in-progress)
- **Tree View**: Dependency hierarchy with ASCII tree visualization
- Toggle between views with `t` key
- Stats dashboard overlay with `S` key

#### Navigation (Vim-style)
- `j`/`k` - Move up/down
- `g` `g` / `G` - Jump to top/bottom
- `Ctrl-d`/`Ctrl-u` - Half-page scroll in detail panel
- `Ctrl-b`/`Ctrl-f` - Full-page scroll
- `/` - Search with `n`/`N` for next/previous match
- `Tab`/`Enter` - Focus detail panel, `ESC` to return

#### Issue Management
- **Create issues** (`a`) - Form-based dialog with natural language detection for priority/type
- **Edit issues** (`e`) - Form-based editing of all fields (description, design, acceptance, notes)
- **Quick priority** (`0`-`4`) - Instantly set priority P0-P4
- **Status cycling** (`s`) - Cycle through open → in_progress → blocked → closed
- **Close/reopen** (`x`/`o`) - Quick issue status changes
- **Add comments** (`c`) - Comment dialog with keyboard shortcuts

#### Dependency Management
- **Dependency dialog** (`D`) - Add/remove blocking and parent-child relationships
- Human-readable dependency phrases ("blocked by", "child of" instead of raw types)
- Visual dependency indicators in list and tree views
- Automatic blocked status detection based on open dependencies
- Blocking propagates through parent-child relationships (matches `bd ready` behavior)

#### Labels
- **Label management** (`L`) - Add/remove labels from issues
- Label display with hashtag prefix in issue list
- Filter by label in quick filter

#### Filtering
- **Quick filter** (`f`) - Filter by status, priority, type, or label
- **Show closed** (`C`) - Toggle visibility of closed issues
- Persistent filter state during session

#### Themes
- 10+ built-in color themes (Gruvbox Dark default)
- High-contrast and colorblind accessibility themes
- TOML-based theme files with embed.FS
- Theme configuration via `~/.config/beads-tui/config.toml`

#### Clipboard
- `y` - Yank issue ID to clipboard
- `Y` - Yank issue ID with title
- Click issue ID in detail panel to copy

#### Other Features
- Help screen (`?`) with all keyboard shortcuts
- Mouse mode toggle (`m`) for terminal text selection
- Debug logging (`--debug`) for troubleshooting
- Type emoji icons in issue list
- Keyboard shortcut hints in dialog titles

### Technical

- Built with [tview](https://github.com/rivo/tview) terminal UI framework
- Uses [fsnotify](https://github.com/fsnotify/fsnotify) for file watching
- SQLite database access via [go-sqlite3](https://github.com/ncruces/go-sqlite3)
- Cross-platform: macOS (Intel/Apple Silicon) and Linux (amd64/arm64)

[0.2.0]: https://github.com/andynu/beads-tui/releases/tag/v0.2.0
[0.1.0]: https://github.com/andynu/beads-tui/releases/tag/v0.1.0
