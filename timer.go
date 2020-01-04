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

var (
	runStart   time.Time
	splitStart time.Time
)

func elapsedStr(start time.Time) string {
	elapsed := time.Since(start)
	elapsed = elapsed - (elapsed % refreshInterval)
	return elapsed.String()
}

func drawRunTime(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
	tview.Print(screen, elapsedStr(runStart), x, height/2, width, tview.AlignCenter, tcell.ColorLime)
	return 0, 0, 0, 0
}

func drawSplitTime(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
	tview.Print(screen, elapsedStr(splitStart), x, height/2, width, tview.AlignCenter, tcell.ColorLime)
	return 0, 0, 0, 0
}

func startSplit() {
	splitStart = time.Now()
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

// Returning nil stops the input from propagating.
func inputHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case ' ':
		startSplit()
		return nil
	}

	switch event.Key() {
	case tcell.KeyEnter:
		startSplit()
		return nil
	}
	return event
}

/*
Intended layout:
Layout:
| BoB 1    | <comparison segment> | <comparison split> |
| PSS 2    |                   +1 |               2:44 |
| Whomps 9 |                   -3 |               8:33 |
( highlight current row)
---
- Total Time :: xx:xx
- Segment :: xx
- Gold :: xx
- Time Save :: xx
- Best Possible Time :: xx
- Sum of Gold :: xx
*/
func startTimer() error {
	var (
	// totalTimeBox  *tview.Box
	// totalDuration time.Duration
	// splitDuration time.Duration
	)
	// Goal:
	// Implement two timers:
	// One for total time
	// One that resets everytime enter is pressed.

	now := time.Now()
	runStart = now
	splitStart = now

	totalDurationBox := tview.NewBox().SetDrawFunc(drawRunTime)
	splitDurationBox := tview.NewBox().SetDrawFunc(drawSplitTime)

	timerContainer := tview.NewFlex().SetFullScreen(true).
		AddItem(totalDurationBox, 0, 1, false).
		AddItem(splitDurationBox, 0, 1, false)

	go refresh()
	if err := app.SetRoot(timerContainer, true).SetInputCapture(inputHandler).Run(); err != nil {
		return err
	}

	return nil
}
