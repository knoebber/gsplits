package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell"
	"github.com/knoebber/gsplits/route"
	"github.com/rivo/tview"
)

const (
	// Start showing time save within this many nano seconds
	plusMinusThreshold = (10 * 1e9) * -1
	placeholder        = "___"
)

var (
	runStart         time.Time
	segmentStart     time.Time
	currentGold      time.Duration
	possibleTimeSave time.Duration

	splitsTable          *tview.Table
	totalTimeView        *tview.TextView
	segmentTimeView      *tview.TextView
	goldView             *tview.TextView
	possibleTimeSaveView *tview.TextView
	bestPossibleTimeView *tview.TextView
	sumOfGoldView        *tview.TextView

	splitIndex int
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

func getPlusMinus(routeData *route.Data, total time.Duration) (plusMinus string, color tcell.Color, show bool) {
	var lastDiff time.Duration

	diff := total - routeData.GetComparisonSplit(splitIndex)

	plusMinus = durationStr(diff)
	if diff <= 0 {
		color = tcell.ColorGreen
	} else {
		color = tcell.ColorRed
	}

	if splitIndex > 0 {
		lastSplit := segmentStart.Sub(runStart)
		lastDiff = (lastSplit - routeData.GetComparisonSplit(splitIndex-1))
	}

	if diff < 0 && lastDiff < 0 {
		show = diff > plusMinusThreshold+lastDiff
	} else if diff < 0 {
		show = diff > plusMinusThreshold
	} else {
		show = true
	}

	return
}

func isDone(segments []time.Duration) bool {
	// The run is finished once the last segment time is filled in.
	return segments[len(segments)-1] != 0
}

func previousSplit(routeData *route.Data, segments []time.Duration) {
	if splitIndex == 0 {
		return
	}

	segments[splitIndex] = 0

	// Make the current row inactive
	setTableCell(splitsTable, splitIndex, 0, routeData.GetSplitName(splitIndex), tcell.ColorDefault)
	setTableCell(splitsTable, splitIndex, 1, placeholder, tcell.ColorDefault)

	// Make the previous row active again.
	splitIndex--
	setTableCell(splitsTable, splitIndex, 1, placeholder, tcell.ColorDefault)
	setTableCell(splitsTable, splitIndex, 2, durationStr(routeData.GetComparisonSegment(splitIndex)), tcell.ColorDefault)
	setTableCell(splitsTable, splitIndex, 3, durationStr(routeData.GetComparisonSplit(splitIndex)), tcell.ColorDefault)

	// Reset the segment time.
	now := time.Now()
	lastSegment := segments[splitIndex]
	revert := now.Sub(segmentStart)
	segmentStart = now.Add((lastSegment + revert) * -1)
}

func nextSplit(routeData *route.Data, segments []time.Duration) {
	splitTime := time.Since(runStart)
	segmentTime := time.Since(segmentStart)

	plusMinus, plusMinusColor, _ := getPlusMinus(routeData, splitTime)

	setTableCell(splitsTable, splitIndex, 0, routeData.GetSplitName(splitIndex), tcell.ColorDefault)
	setTableCell(splitsTable, splitIndex, 1, plusMinus, plusMinusColor)
	setTableCell(splitsTable, splitIndex, 2, durationStr(segmentTime), tcell.ColorDefault)
	setTableCell(splitsTable, splitIndex, 3, durationStr(splitTime), tcell.ColorDefault)

	segments[splitIndex] = segmentTime
	gold := routeData.GetGold(splitIndex)

	if segmentTime < gold {
		timeSave := gold - segmentTime

		if routeData.SumOfGold != nil {
			*routeData.SumOfGold -= timeSave
		}
	}

	if splitIndex < routeData.Length-1 {
		segmentStart = time.Now()
		splitIndex++
	}
}

func getInputHandler(routeData *route.Data, segments []time.Duration) func(event *tcell.EventKey) *tcell.EventKey {
	var totalDuration time.Duration

	handleNextSplit := func() {
		if isDone(segments) {
			promptSaveRun(routeData.RouteID, segments, totalDuration)
		} else {
			nextSplit(routeData, segments)
		}

		// splitIndex may have just been incremented by nextSplit()
		// It is global variable so this is hard to read.
		// TODO make a better state system for this.
		if splitIndex == routeData.Length-1 {
			totalDuration = time.Since(runStart)
		}
	}

	resetRun := func() {
		// Need to start the goroutine if the splits were already finished.
		startThread := isDone(segments)

		for i := range segments {
			segments[i] = 0
		}
		splitIndex = 0
		totalDuration = 0
		initRun(routeData)

		if startThread {
			go refresh(routeData, segments)
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
			previousSplit(routeData, segments)
		}
		return event
	}
}

func setSplitsTable(routeData *route.Data) {
	if splitsTable == nil {
		splitsTable = newTable()
	}

	for i := range routeData.SplitNames {
		for j, value := range []string{
			routeData.GetSplitName(i),
			placeholder,
			durationStr(routeData.GetComparisonSegment(i)),
			durationStr(routeData.GetComparisonSplit(i)),
		} {
			setTableCell(splitsTable, i, j, value, tcell.ColorDefault)
		}
	}
}

// Draws the row for the active split.
func drawCurrentSplitRow(routeData *route.Data, runDuration time.Duration) {
	plusMinus, plusMinusColor, showPlusMinus := getPlusMinus(routeData, runDuration)

	setTableCell(splitsTable, splitIndex, 0, routeData.GetSplitName(splitIndex), tcell.ColorYellow)
	if showPlusMinus {
		setTableCell(splitsTable, splitIndex, 1, plusMinus, plusMinusColor)
	}
}

func createSplitsFlex(routeData *route.Data) *tview.Flex {
	totalTime := tview.NewFlex().
		AddItem(newText("Total Time"), 0, 1, false).
		AddItem(totalTimeView, 0, 1, false).
		AddItem(nil, 0, 1, false)

	segmentTime := tview.NewFlex().
		AddItem(newText("Segment Time"), 0, 1, false).
		AddItem(segmentTimeView, 0, 1, false).
		AddItem(nil, 0, 1, false)

	bestPossibleTime := tview.NewFlex().
		AddItem(newText("Best Possible Time"), 0, 1, false).
		AddItem(bestPossibleTimeView, 0, 1, false).
		AddItem(nil, 0, 1, false)

	gold := tview.NewFlex().
		AddItem(newText("Gold"), 0, 1, false).
		AddItem(goldView, 0, 1, false).
		AddItem(nil, 0, 1, false)

	possibleTimeSave := tview.NewFlex().
		AddItem(newText("Possible Time Save"), 0, 1, false).
		AddItem(possibleTimeSaveView, 0, 1, false).
		AddItem(nil, 0, 1, false)

	sumOfGold := tview.NewFlex().
		AddItem(newText("Sum of Gold"), 0, 1, false).
		AddItem(sumOfGoldView, 0, 1, false).
		AddItem(nil, 0, 1, false)

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(totalTime, 1, 0, false).
		AddItem(segmentTime, 1, 0, false).
		AddItem(bestPossibleTime, 1, 0, false).
		AddItem(gold, 1, 0, false).
		AddItem(possibleTimeSave, 1, 0, false).
		AddItem(sumOfGold, 1, 0, false)
}

func getDrawFunc(routeData *route.Data, segments []time.Duration) func() {
	return func() {
		runDuration := time.Since(runStart)
		lastSplit := segmentStart.Sub(runStart)
		diff := runDuration - routeData.GetComparisonSplit(splitIndex)

		drawCurrentSplitRow(routeData, runDuration)
		totalTimeView.SetText(elapsedStr(runStart))
		segmentTimeView.SetText(elapsedStr(segmentStart))
		goldView.SetText(durationStr(routeData.GetGold(splitIndex)))
		possibleTimeSaveView.SetText(durationStr(routeData.GetTimeSave(splitIndex)))
		bestPossibleTimeView.SetText(durationStr(routeData.GetBPT(splitIndex, lastSplit, diff)))
		sumOfGoldView.SetText(safeDurationStr(routeData.SumOfGold))
	}
}

func refresh(routeData *route.Data, segments []time.Duration) {
	tick := time.NewTicker(refreshInterval)
	drawFunc := getDrawFunc(routeData, segments)

	for !isDone(segments) {
		select {
		case <-tick.C:
			app.QueueUpdateDraw(drawFunc)
		}
	}
}

func initRun(routeData *route.Data) {
	now := time.Now()
	runStart = now
	segmentStart = now

	setSplitsTable(routeData)
}

func startTimer(routeData *route.Data) {
	segments := make([]time.Duration, routeData.Length)

	initRun(routeData)

	totalTimeView = newText(elapsedStr(runStart))
	segmentTimeView = newText(elapsedStr(runStart))
	goldView = newText(durationStr(routeData.GetGold(0)))
	possibleTimeSaveView = newText(durationStr(routeData.GetTimeSave(0)))
	bestPossibleTimeView = newText(durationStr(routeData.GetGold(0)))
	sumOfGoldView = newText(safeDurationStr(routeData.SumOfGold))

	container := tview.NewFlex().SetDirection(tview.FlexRow).SetFullScreen(true).
		AddItem(splitsTable, 0, 10, true).
		AddItem(nil, 0, 1, false).
		AddItem(createSplitsFlex(routeData), 0, 4, false)

	go refresh(routeData, segments)
	app.SetRoot(container, true).SetInputCapture(getInputHandler(routeData, segments))
}
