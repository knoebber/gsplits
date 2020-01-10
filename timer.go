package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell"
	"github.com/knoebber/gsplits/route"
	"github.com/rivo/tview"
)

func promptSaveRun(routeID int64, segments []time.Duration, totalDuration time.Duration) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Total Time: %s\nSave Run?", durationStr(totalDuration))).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				_, err := saveRun(routeID, segments, totalDuration)
				if err != nil {
					panic(err)
				}
			}
			app.Stop()
		})

	app.SetRoot(modal, false).SetFocus(modal)
}

func previousSplit(state *timerState) {
	if state.splitIndex == 0 {
		return
	}

	state.segments[state.splitIndex] = 0

	// Make the current row inactive
	state.setTableCell(0, state.routeData.GetSplitName(state.splitIndex), tcell.ColorDefault)
	state.setTableCell(1, placeholder, tcell.ColorDefault)

	// Make the previous row active again.
	state.splitIndex--
	comparisonSegment := state.routeData.GetComparisonSegment(state.splitIndex)
	comparisonSplit := state.routeData.GetComparisonSplit(state.splitIndex)
	state.setTableCell(1, placeholder, tcell.ColorDefault)
	state.setTableCell(2, durationStr(comparisonSegment), tcell.ColorDefault)
	state.setTableCell(3, durationStr(comparisonSplit), tcell.ColorDefault)

	// Reset the segment time.
	now := time.Now()
	lastSegment := state.segments[state.splitIndex]
	revert := now.Sub(state.segmentStart)
	state.segmentStart = now.Add((lastSegment + revert) * -1)
}

func nextSplit(state *timerState) {
	splitTime := time.Since(state.runStart)
	segmentTime := time.Since(state.segmentStart)

	plusMinus, plusMinusColor, _ := state.getPlusMinus(splitTime)

	state.setTableCell(0, state.routeData.GetSplitName(state.splitIndex), tcell.ColorDefault)
	state.setTableCell(1, plusMinus, plusMinusColor)
	state.setTableCell(2, durationStr(segmentTime), tcell.ColorDefault)
	state.setTableCell(3, durationStr(splitTime), tcell.ColorDefault)

	state.segments[state.splitIndex] = segmentTime
	gold := state.routeData.GetGold(state.splitIndex)

	if segmentTime < gold {
		timeSave := gold - segmentTime

		if state.sumOfGold != nil {
			*state.sumOfGold -= timeSave
		}
	}

	if state.splitIndex < state.routeData.Length-1 {
		state.segmentStart = time.Now()
		state.splitIndex++
	}
}

func getInputHandler(state *timerState) func(event *tcell.EventKey) *tcell.EventKey {
	handleNextSplit := func() {
		if state.isDone() {
			promptSaveRun(state.routeData.RouteID, state.segments, state.totalDuration)
		} else {
			nextSplit(state)
		}

		if state.splitIndex == state.routeData.Length-1 {
			state.setTotalDuration()
		}
	}

	resetRun := func() {
		// Need to start the goroutine if the splits were already finished.
		startThread := state.isDone()

		state.reset()

		if startThread {
			go refresh(state)
		}
	}

	// Returning nil stops the input from propagating.
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'r':
			resetRun()

		case ' ':
			handleNextSplit()
			return nil
		}

		switch event.Key() {
		case tcell.KeyCtrlSpace:
			previousSplit(state)
		}
		return event
	}
}

func refresh(state *timerState) {
	tick := time.NewTicker(refreshInterval)
	drawFunc := state.getDrawFunc()

	for !state.isDone() {
		select {
		case <-tick.C:
			app.QueueUpdateDraw(drawFunc)
		}
	}
}

func startTimer(routeData *route.Data) {
	state := newTimerState(routeData)
	container := state.createLayout()
	go refresh(state)
	app.SetRoot(container, true).SetInputCapture(getInputHandler(state))
}
