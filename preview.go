package main

import (
	"fmt"

	"github.com/gdamore/tcell"
	"github.com/knoebber/gsplits/route"
	"github.com/rivo/tview"
)

func showPreview(routeData *route.Data) (err error) {
	var (
		title string
		best  string
	)

	title = fmt.Sprintf("%s: %s", routeData.Category.Name, routeData.RouteName)
	if routeData.Category.Best != nil {
		best = fmt.Sprintf("%s Best: %s", routeData.Category.Name, *routeData.Category.Best)
		if routeData.RouteBestTime != nil && *routeData.Category.Best < *routeData.RouteBestTime {
			// Print the route best time only if its slower than the categories best.
			best += fmt.Sprintf("\n%s Best: %s", routeData.RouteName, *routeData.RouteBestTime)
		}
	} else {
		best = "No runs yet"
	}

	table := newTable()

	onTableFocus := func(focus bool) {
		for col, value := range []string{
			"Name", "Split Time", "Segment Duration", "Gold", "Possible Save",
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

	for i := range routeData.SplitNames {
		for j, value := range []string{
			routeData.GetSplitName(i),
			durationStr(routeData.GetComparisonSplit(i)),
			durationStr(routeData.GetComparisonSegment(i)),
			durationStr(routeData.GetGold(i)),
			durationStr(routeData.GetTimeSave(i)),
		} {
			setTableCell(table, i+1, j, value, tcell.ColorWhite)
		}
	}

	quitButton := newButton("Quit").SetSelectedFunc(func() {
		app.Stop()
	})

	quitButton.SetBlurFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyTab:
			onTableFocus(true)
			app.SetFocus(table)
		}
	})

	startButton := newButton("Start").SetSelectedFunc(func() {
		startTimer(routeData)
	})

	startButton.SetBlurFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyTab:
			app.SetFocus(quitButton)
		}
	})

	table.SetFixed(1, 1)
	table.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			startTimer(routeData)
		case tcell.KeyTab:
			onTableFocus(false)
			app.SetFocus(startButton)
		case tcell.KeyExit:
			app.Stop()
		}
	})

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
