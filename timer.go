package main

import (
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// Measure times to the tenth of a second.
const (
	refreshInterval = time.Second / 10
	timeFormat      = "15:04:05.0"
)

var (
	timer *tview.Box
)

func drawTime(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
	timeStr := time.Now().Format(timeFormat)
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

func startTimer() error {
	timer = tview.NewBox().SetDrawFunc(drawTime)
	if err := app.SetRoot(timer, true).Run(); err != nil {
		panic(err)
	}

	go refresh()
	return nil
}
