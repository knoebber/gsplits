package main

import (
	"time"

	"github.com/gdamore/tcell"
	"github.com/knoebber/gsplits/route"
	"github.com/rivo/tview"
)

const refreshInterval = 60 / time.Second

var (
	view *tview.Box
	app  *tview.Application
)

func drawTime(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
	timeStr := time.Now().Format("Current time is 15:04:05")
	tview.Print(screen, timeStr, x, height/2, width, tview.AlignCenter, tcell.ColorLime)
	return 0, 0, 0, 0
}

func refresh() {
	tick := time.NewTicker(refreshInterval)
	for {
		select {
		case <-tick.C:
			app.Draw()
		}
	}
}

func safeDurationString(lst []time.Duration, i int) string {
	if i > len(lst)-1 {
		return "N/A"
	}
	return lst[i].String()
}

func newText(text string) tview.Primitive {
	return tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(text)
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

func showPreview(d *route.Data) error {
	table := tview.NewTable().SetBorders(true)
	setTableCell := func(row, column int, value string, color tcell.Color) {
		table.SetCell(row, column,
			tview.NewTableCell(value).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))

	}
	// rows := make([][]string, len(d.SplitNames)+1)

	for col, value := range []string{"Name", "Split Time", "Split Duration", "Gold", "Possible Save"} {
		setTableCell(0, col, value, tcell.ColorYellow)
	}

	for i := range d.SplitNames {
		for j, value := range []string{
			d.SplitNames[i].Name,
			safeDurationString(d.Comparison, i),
			safeDurationString(d.RouteBests, i),
			safeDurationString(d.Golds, i),
			safeDurationString(d.TimeSaves, i),
		} {
			setTableCell(i+1, j, value, tcell.ColorWhite)
		}
	}

	table.SetFixed(1, 1)
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
		AddItem(table, 0, 9, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow), 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(newButton("Start"), 10, 1, false).
			AddItem(nil, 7, 2, false).
			AddItem(newButton("(Q)uit"), 10, 1, false),
			0, 1, true)

	if err := app.SetRoot(flex, true).SetFocus(table).Run(); err != nil {
		return err
	}
	// 	p := widgets.NewParagraph()
	// 	p.Title = fmt.Sprintf("%s: %s", d.Category.Name, d.RouteName)
	// 	if d.Category.Best != nil {
	// 		p.Text = fmt.Sprintf("%s Best: %s", d.Category.Name, *d.Category.Best)
	// 		if d.RouteBestTime != nil && *d.Category.Best < *d.RouteBestTime {
	// 			// Print the route best time only if its slower than the categories best.
	// 			p.Text += fmt.Sprintf("\n%s Best: %s", d.RouteName, *d.RouteBestTime)
	// 		}
	// 	} else {
	// 		p.Text = "No runs yet"
	// 	}
	// 	p.Text += "\nStart or q to quit"

	// 	dataTable := widgets.NewTable()
	// 	dataTable.Rows = rows
	// 	dataTable.TextStyle = ui.NewStyle(ui.ColorWhite)
	// button := 	// button.SetBorder(true).SetRect(0, 0, 22, 3)
	// if err := app.SetRoot(button, false).SetFocus(button).Run(); err != nil {
	// 	panic(err)
	// }
	// return nil

	// table.Select(0, 0).SetFixed(1, 1).SetDoneFunc(func(key tcell.Key) {
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
	return nil
}

func startSplits(d *route.Data) error {
	app = tview.NewApplication()
	view = tview.NewBox().SetDrawFunc(drawTime)
	return showPreview(d)
}

// func main() {
// 	menu := newPrimitive("Menu")
// 	main := newPrimitive("Main content")
// 	sideBar := newPrimitive("Side Bar")

// 	grid := tview.NewGrid().
// 		SetRows(3, 0, 3).
// 		SetColumns(30, 0, 30).
// 		SetBorders(true).
// 		AddItem(newPrimitive("Header"), 0, 0, 1, 3, 0, 0, false).
// 		AddItem(newPrimitive("Footer"), 2, 0, 1, 3, 0, 0, false)

// 	// Layout for screens narrower than 100 cells (menu and side bar are hidden).
// 	grid.AddItem(menu, 0, 0, 0, 0, 0, 0, false).
// 		AddItem(main, 1, 0, 1, 3, 0, 0, false).
// 		AddItem(sideBar, 0, 0, 0, 0, 0, 0, false)

// 	// Layout for screens wider than 100 cells.
// 	grid.AddItem(menu, 1, 0, 1, 1, 0, 100, false).
// 		AddItem(main, 1, 1, 1, 1, 0, 100, false).
// 		AddItem(sideBar, 1, 2, 1, 1, 0, 100, false)

// 	if err := tview.NewApplication().SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
// 		panic(err)
// 	}
// }

// 	go refresh()
// 	if err := app.SetRoot(view, true).Run(); err != nil {
// 		return err
// 	}
// 	return nil
// }
