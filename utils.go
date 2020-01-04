package main

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func newTable() *tview.Table {
	return tview.NewTable().SetBorders(true)
}

func setTableCell(table *tview.Table, row, column int, value string, color tcell.Color) {
	table.SetCell(row, column,
		tview.NewTableCell(value).
			SetTextColor(color).
			SetAlign(tview.AlignCenter))

}
