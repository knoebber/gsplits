package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const (
	// Accuracy of time measurment.
	refreshInterval = time.Second / 10

	// The minimum size a duration string will be.
	// Prevents containers from resizing as the duration size changes sizes.
	minDurationLength = 10
)

func newText(text string) *tview.TextView {
	return tview.NewTextView().SetText(text)
}

func newButton(text string) *tview.Button {
	button := tview.NewButton(text)
	button.SetBorder(true)
	return button
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

func durationStr(d time.Duration) string {
	// Show only as many digits at the refresh interval.
	d = d - (d % refreshInterval)
	return fmt.Sprintf("%*s", minDurationLength, d)
}
