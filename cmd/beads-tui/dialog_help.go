package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ShowHelpScreen displays the keyboard shortcuts help screen
func (h *DialogHelpers) ShowHelpScreen() {
	// Note: This help screen uses hardcoded colors for documentation purposes
	// showing the current theme's colors as examples
	helpText := `[yellow::b]beads-tui Keyboard Shortcuts[-::-]

[cyan::b]Navigation[-::-]
  j / ↓       Move down
  k / ↑       Move up
  gg          Jump to top
  G           Jump to bottom
  Tab         Focus detail panel for scrolling
  Enter       Focus detail panel (when on issue)
  ESC         Return focus to issue list

[cyan::b]Search[-::-]
  /           Start search mode
  n           Next search result
  N           Previous search result
  ESC         Exit search mode

[cyan::b]Quick Actions[-::-]
  0-4         Set priority (P0=critical, P1=high, P2=normal, P3=low, P4=lowest)
  s           Cycle status (open → in_progress → blocked → closed → open)
  R           Rename issue (edit title)
  a           Create new issue (vim-style "add")
  c           Add comment to selected issue
  e           Edit issue (title, description, design, acceptance, notes, priority, type)
  x           Close issue with optional reason
  X           Reopen closed issue with optional reason
  D           Manage dependencies (add/remove blocks, parent-child, related)
  L           Manage labels (add/remove labels)
  y           Yank (copy) issue ID to clipboard
  Y           Yank (copy) issue ID with title to clipboard
  B           Copy git branch name to clipboard

[cyan::b]Two-Character Shortcuts[-::-]
  So          Set status to open
  Si          Set status to in_progress
  Sb          Set status to blocked
  Sc          Set status to closed

[cyan::b]View Controls[-::-]
  t           Toggle between list and tree view
  T           Cycle to next theme (live theme switching)
  C           Toggle showing closed issues in list view
  f           Quick filter (type: p1 bug, feature, etc.)
  S           Show statistics dashboard
  m           Toggle mouse mode on/off
  r           Manual refresh

[cyan::b]Detail Panel Scrolling (when focused)[-::-]
  Ctrl-d      Scroll down half page
  Ctrl-u      Scroll up half page
  Ctrl-f      Scroll down full page (vim)
  Ctrl-b      Scroll up full page (vim)
  Ctrl-e      Scroll down one line
  Ctrl-y      Scroll up one line
  PageDown    Scroll down full page
  PageUp      Scroll up full page
  Home        Jump to top of details
  End         Jump to bottom of details

[cyan::b]General[-::-]
  ?           Show this help screen
  q           Quit

[cyan::b]Command Line Options[-::-]
  --theme <name>      Set color theme
    beads-tui --theme gruvbox-dark

  --view <mode>       Start in list or tree view
    beads-tui --view tree

  --issue <id>        Show only a specific issue
    beads-tui --issue tui-abc

  --debug             Enable debug logging

[cyan::b]Themes[-::-]
  Available themes: default, gruvbox-dark, gruvbox-light, nord,
  solarized-dark, solarized-light, dracula, tokyo-night,
  catppuccin-mocha, catppuccin-latte

  Set via environment variable:
    export BEADS_THEME=gruvbox-dark

[cyan::b]Status Icons[-::-]
  ●           Open/Ready
  ○           Blocked
  ◆           In Progress
  ·           Other

[cyan::b]Priority Colors[-::-]
  [red]P0[-]          Critical
  [orangered]P1[-]          High
  [lightskyblue]P2[-]          Normal
  [gray]P3[-]          Low
  [gray]P4[-]          Lowest

[cyan::b]Status Colors[-::-]
  [limegreen]●[-]           Ready
  [gold]○[-]           Blocked
  [deepskyblue]◆[-]           In Progress
  [gray]·[-]           Closed

[gray]━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━[-]
[yellow]Press ESC or ? to close this help[-]`

	// Create help text view
	helpTextView := tview.NewTextView().
		SetDynamicColors(true).
		SetText(helpText).
		SetTextAlign(tview.AlignLeft)
	helpTextView.SetBorder(true).
		SetTitle(" Help - Keyboard Shortcuts ").
		SetTitleAlign(tview.AlignCenter)

	// Create modal (centered)
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(helpTextView, 0, 3, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)

	// Add input capture to close on ESC or ?
	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || (event.Key() == tcell.KeyRune && event.Rune() == '?') {
			h.Pages.RemovePage("help")
			h.App.SetFocus(h.IssueList)
			return nil
		}
		return event
	})

	h.Pages.AddPage("help", modal, true, true)
	h.App.SetFocus(modal)
}
