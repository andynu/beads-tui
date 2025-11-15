package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Components holds all the TUI widgets
type Components struct {
	App         *tview.Application
	StatusBar   *tview.TextView
	IssueList   *tview.List
	DetailPanel *tview.TextView
	Pages       *tview.Pages
	Layout      *tview.Flex
}

// CreateComponents initializes all TUI widgets with default settings
func CreateComponents() *Components {
	app := tview.NewApplication()

	// Status bar
	statusBar := tview.NewTextView().
		SetDynamicColors(true)

	// Issue list
	issueList := tview.NewList().
		ShowSecondaryText(false).
		SetSelectedBackgroundColor(tcell.ColorDarkCyan).
		SetSelectedTextColor(tcell.ColorBlack)
	issueList.SetBorder(true).SetTitle("Issues")

	// Detail panel
	detailPanel := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)
	detailPanel.SetBorder(true).SetTitle("Details")

	// Layout: vertical flex with status bar, horizontal split, pages for modals
	horizontalLayout := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(issueList, 0, 1, true).
		AddItem(detailPanel, 0, 2, false)

	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(statusBar, 1, 0, false).
		AddItem(horizontalLayout, 0, 1, true)

	// Pages for modal dialogs
	pages := tview.NewPages().
		AddPage("main", mainLayout, true, true)

	return &Components{
		App:         app,
		StatusBar:   statusBar,
		IssueList:   issueList,
		DetailPanel: detailPanel,
		Pages:       pages,
		Layout:      mainLayout,
	}
}
