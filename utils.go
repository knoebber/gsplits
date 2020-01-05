package main

import (
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// Accuracy of time measurment.
const (
	refreshInterval = time.Second / 10
)

func newButton(text string) *tview.Button {
	button := tview.NewButton(text).SetSelectedFunc(
		func() {
			app.Stop()
		},
	)
	button.SetBorder(true)
	return button
}

func newText(text string) *tview.TextView {
	return tview.NewTextView().SetText(text)
}

func newTable() *tview.Table {
	return tview.NewTable().SetBorders(true)
}

func setTableCell(table *tview.Table, row, column int, value string, color tcell.Color) {
	table.SetCell(row, column,
		tview.NewTableCell(value).
			SetTextColor(color).
			SetAlign(tview.AlignCenter))

}

func lstDurationStr(lst []time.Duration, i int) string {
	if i > len(lst)-1 {
		return "N/A"
	}
	return durationStr(lst[i])
}

func safeDurationStr(d *time.Duration) string {
	if d == nil {
		return "N/A"
	}
	return durationStr(*d)
}

func elapsedStr(start time.Time) string {
	return durationStr(time.Since(start))
}

// TODO make better formatter. Not showing 0 is breaking things.
func durationStr(d time.Duration) string {
	d = d - (d % refreshInterval)
	return d.String()
}
