package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell"
	"github.com/knoebber/gsplits/route"
	"github.com/rivo/tview"
)

func newText(text string) tview.Primitive {
	return tview.NewTextView().SetText(text)
}

func newButton(text string) *tview.Button {
	button := tview.NewButton(text).SetSelectedFunc(
		func() {
			app.Stop()
		},
	)
	button.SetBorder(true)
	return button
}

func safeDurationString(lst []time.Duration, i int) string {
	if i > len(lst)-1 {
		return "N/A"
	}
	return lst[i].String()
}

func showPreview(d *route.Data) (err error) {
	var (
		title string
		best  string
	)

	title = fmt.Sprintf("%s: %s", d.Category.Name, d.RouteName)
	if d.Category.Best != nil {
		best = fmt.Sprintf("%s Best: %s", d.Category.Name, *d.Category.Best)
		if d.RouteBestTime != nil && *d.Category.Best < *d.RouteBestTime {
			// Print the route best time only if its slower than the categories best.
			best += fmt.Sprintf("\n%s Best: %s", d.RouteName, *d.RouteBestTime)
		}
	} else {
		best = "No runs yet"
	}

	table := newTable()

	// rows := make([][]string, len(d.SplitNames)+1)

	onTableFocus := func(focus bool) {
		for col, value := range []string{
			"Name", "Split Time", "Split Duration", "Gold", "Possible Save",
		} {

			if focus {
				setTableCell(table, 0, col, value, tcell.ColorYellow)
			} else {
				setTableCell(table, 0, col, value, tcell.ColorWhite)
			}

		}
	}
	// Table is initially focused.
	onTableFocus(true)

	for i := range d.SplitNames {
		for j, value := range []string{
			d.SplitNames[i].Name,
			safeDurationString(d.Comparison, i),
			safeDurationString(d.RouteBests, i),
			safeDurationString(d.Golds, i),
			safeDurationString(d.TimeSaves, i),
		} {
			setTableCell(table, i+1, j, value, tcell.ColorWhite)
		}
	}

	startButton := newButton("Start").SetSelectedFunc(func() {
		if err = startTimer(); err != nil {
			panic(err)
		}
	})
	quitButton := newButton("Quit").SetSelectedFunc(func() {
		app.Stop()
	})

	startButton.SetBlurFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyTab:
			app.SetFocus(quitButton)
		}
	})
	quitButton.SetBlurFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyTab:
			onTableFocus(true)
			app.SetFocus(table)
		}
	})

	table.SetFixed(1, 1)
	table.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			go refresh()
			if err = startTimer(); err != nil {
				panic(err)
			}
			// if err = app.SetRoot(timer, true).Run(); err != nil {
			//	panic(err)
			//}
		case tcell.KeyTab:
			onTableFocus(false)
			app.SetFocus(startButton)
		case tcell.KeyExit:
			app.Stop()
		}
	})

	// 	if key == tcell.KeyEscape {
	// 		app.Stop()
	// 	}
	// 	if key == tcell.KeyEnter {
	// 		table.SetSelectable(true, true)
	// 	}
	// }).SetSelectedFunc(func(row int, column int) {
	// 	table.GetCell(row, column).SetTextColor(tcell.ColorRed)
	// 	table.SetSelectable(false, false)
	// })

	flex := tview.NewFlex().SetDirection(tview.FlexRow).SetFullScreen(true).
		AddItem(newText(title), 0, 1, false).
		AddItem(newText(best), 0, 1, false).
		AddItem(table, 0, 8, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow), 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(startButton, 10, 1, false).
			AddItem(nil, 7, 2, false).
			AddItem(quitButton, 10, 1, false),
			0, 1, true)

	return app.SetRoot(flex, true).SetFocus(table).Run()
}
