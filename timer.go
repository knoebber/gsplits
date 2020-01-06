package main

import (
	"time"

	"github.com/gdamore/tcell"
	"github.com/knoebber/gsplits/route"
	"github.com/rivo/tview"
)

const (
	// Start showing time save within this many nano seconds
	plusMinusThreshold = (10 * 1e9) * -1

	placeholder = "___"
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

	splitIndex int
)

func getPlusMinus(routeData *route.Data, total time.Duration) (string, tcell.Color, bool) {
	var (
		lastDiff       time.Duration
		plusMinusColor tcell.Color
	)

	diff := total - routeData.Comparison[splitIndex]
	plusMinus := durationStr(diff)
	if diff <= 0 {
		plusMinusColor = tcell.ColorGreen
	} else {
		plusMinusColor = tcell.ColorRed
	}

	if splitIndex > 0 {
		lastDiff = total - routeData.Comparison[splitIndex-1]
	}

	showPlusMinus := diff > (plusMinusThreshold + (lastDiff * -1))

	return plusMinus, plusMinusColor, showPlusMinus
}

func previousSplit(routeData *route.Data, segments []time.Duration) {
	if splitIndex == 0 {
		return
	}

	// Make the current row inactive
	setTableCell(splitsTable, splitIndex, 0, routeData.SplitNames[splitIndex].Name, tcell.ColorWhite)
	setTableCell(splitsTable, splitIndex, 1, placeholder, tcell.ColorWhite)

	// Make the previous row active again.
	splitIndex--
	setTableCell(splitsTable, splitIndex, 1, placeholder, tcell.ColorWhite)
	setTableCell(splitsTable, splitIndex, 2, durationStr(routeData.RouteBests[splitIndex]), tcell.ColorWhite)
	setTableCell(splitsTable, splitIndex, 3, lstDurationStr(routeData.Comparison, splitIndex), tcell.ColorWhite)

	// Reset the segment time.
	now := time.Now()
	lastSegment := segments[splitIndex]
	revert := now.Sub(segmentStart)
	segmentStart = now.Add((lastSegment + revert) * -1)
	segments[splitIndex] = 0
}

func nextSplit(routeData *route.Data, segments []time.Duration) {
	if splitIndex == len(routeData.SplitNames) {
		return
	}

	splitTime := time.Since(runStart)
	segmentTime := time.Since(segmentStart)

	plusMinus, plusMinusColor, _ := getPlusMinus(routeData, splitTime)

	setTableCell(splitsTable, splitIndex, 0, routeData.SplitNames[splitIndex].Name, tcell.ColorWhite)
	setTableCell(splitsTable, splitIndex, 1, plusMinus, plusMinusColor)
	setTableCell(splitsTable, splitIndex, 2, durationStr(segmentTime), tcell.ColorWhite)
	setTableCell(splitsTable, splitIndex, 3, durationStr(splitTime), tcell.ColorWhite)

	segments[splitIndex] = segmentTime

	segmentStart = time.Now()
	splitIndex++

}

func getInputHandler(routeData *route.Data, segments []time.Duration) func(event *tcell.EventKey) *tcell.EventKey {
	// Returning nil stops the input from propagating.
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case ' ':
			nextSplit(routeData, segments)
			return nil
		}

		switch event.Key() {
		case tcell.KeyEnter:
			nextSplit(routeData, segments)
			return nil
		case tcell.KeyCtrlSpace:
			previousSplit(routeData, segments)
		}
		return event
	}
}

func setSplitsTable(routeData *route.Data) {
	splitsTable = newTable()
	for i := range routeData.SplitNames {
		for j, value := range []string{
			routeData.SplitNames[i].Name,
			placeholder,
			lstDurationStr(routeData.RouteBests, i),
			lstDurationStr(routeData.Comparison, i),
		} {
			setTableCell(splitsTable, i, j, value, tcell.ColorWhite)
		}
	}
}

// Draws the row for the active split.
func drawCurrentSplitRow(routeData *route.Data, runDuration time.Duration) {
	plusMinus, plusMinusColor, showPlusMinus := getPlusMinus(routeData, runDuration)

	setTableCell(splitsTable, splitIndex, 0, routeData.SplitNames[splitIndex].Name, tcell.ColorYellow)
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

	gold := tview.NewFlex().
		AddItem(newText("Gold"), 0, 1, false).
		AddItem(goldView, 0, 1, false).
		AddItem(nil, 0, 1, false)

	possibleTimeSave := tview.NewFlex().
		AddItem(newText("Possible Time Save"), 0, 1, false).
		AddItem(possibleTimeSaveView, 0, 1, false).
		AddItem(nil, 0, 1, false)

	bestPossibleTime := tview.NewFlex().
		AddItem(newText("Best Possible Time"), 0, 1, false).
		AddItem(bestPossibleTimeView, 0, 1, false).
		AddItem(nil, 0, 1, false)

	sumOfGold := tview.NewFlex().
		AddItem(newText("Sum of Gold"), 0, 1, false).
		AddItem(newText(safeDurationStr(routeData.SumOfGold)), 0, 1, false).
		AddItem(nil, 0, 1, false)

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(totalTime, 1, 0, false).
		AddItem(segmentTime, 1, 0, false).
		AddItem(gold, 1, 0, false).
		AddItem(possibleTimeSave, 1, 0, false).
		AddItem(bestPossibleTime, 1, 0, false).
		AddItem(sumOfGold, 1, 0, false)
}

func getDrawFunc(routeData *route.Data, segments []time.Duration) func() {
	return func() {
		runDuration := time.Since(runStart)

		drawCurrentSplitRow(routeData, runDuration)
		totalTimeView.SetText(elapsedStr(runStart))
		segmentTimeView.SetText(elapsedStr(segmentStart))
		goldView.SetText(lstDurationStr(routeData.Golds, splitIndex))
		possibleTimeSaveView.SetText(lstDurationStr(routeData.TimeSaves, splitIndex))
		bestPossibleTimeView.SetText(durationStr(routeData.GetBPT(runDuration, splitIndex)))
	}
}

func refresh(routeData *route.Data, splits []time.Duration) {
	tick := time.NewTicker(refreshInterval)
	drawFunc := getDrawFunc(routeData, splits)

	for splitIndex < len(routeData.SplitNames) {
		select {
		case <-tick.C:
			app.QueueUpdateDraw(drawFunc)
		}
	}
}

// func saveRun(routeID int64, durations []time.Duration, totalDuration time.Duration) (runID int64, err error) {
func startTimer(routeData *route.Data) {
	segments := make([]time.Duration, len(routeData.SplitNames))
	now := time.Now()
	runStart = now
	segmentStart = now

	// Times that are redrawn
	totalTimeView = newText(elapsedStr(now))
	segmentTimeView = newText(elapsedStr(now))
	goldView = newText(lstDurationStr(routeData.Golds, 0))
	possibleTimeSaveView = newText(lstDurationStr(routeData.TimeSaves, 0))
	bestPossibleTimeView = newText(lstDurationStr(routeData.Golds, 0))

	setSplitsTable(routeData)
	container := tview.NewFlex().SetDirection(tview.FlexRow).SetFullScreen(true).
		AddItem(splitsTable, 0, 10, true).
		AddItem(nil, 0, 1, false).
		AddItem(createSplitsFlex(routeData), 0, 4, false)

	go refresh(routeData, segments)
	app.SetRoot(container, true).SetInputCapture(getInputHandler(routeData, segments))
}
