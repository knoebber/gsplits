package main

import (
	"time"

	"github.com/gdamore/tcell"
	"github.com/knoebber/gsplits/route"
	"github.com/rivo/tview"
)

// Start showing time save within this many nano seconds
const plusMinusThreshold = (10 * 1e9) * -1

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
	var plusMinusColor tcell.Color

	diff := total - routeData.Comparison[splitIndex]
	plusMinus := durationStr(diff)
	if diff <= 0 {
		plusMinusColor = tcell.ColorGreen
	} else {
		plusMinusColor = tcell.ColorRed
	}
	showPlusMinus := diff > plusMinusThreshold

	return plusMinus, plusMinusColor, showPlusMinus
}

func nextSplit(routeData *route.Data) {
	splitTime := time.Since(runStart)
	plusMinus, plusMinusColor, _ := getPlusMinus(routeData, splitTime)

	setTableCell(splitsTable, splitIndex, 3, durationStr(splitTime), tcell.ColorWhite)

	setTableCell(splitsTable, splitIndex, 0, routeData.SplitNames[splitIndex].Name, tcell.ColorWhite)
	setTableCell(splitsTable, splitIndex, 1, plusMinus, plusMinusColor)

	segmentStart = time.Now()
	splitIndex++
}

func getInputHandler(routeData *route.Data) func(event *tcell.EventKey) *tcell.EventKey {
	// Returning nil stops the input from propagating.
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case ' ':
			nextSplit(routeData)
			return nil
		}

		switch event.Key() {
		case tcell.KeyEnter:
			nextSplit(routeData)
			return nil
		}
		return event
	}
}

func setSplitsTable(routeData *route.Data) {
	splitsTable = newTable()
	for i := range routeData.SplitNames {
		for j, value := range []string{
			routeData.SplitNames[i].Name,
			"___",
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

func getDrawFunc(routeData *route.Data) func() {
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

func refresh(routeData *route.Data) {
	tick := time.NewTicker(refreshInterval)
	drawFunc := getDrawFunc(routeData)

	for {
		select {
		case <-tick.C:
			app.QueueUpdateDraw(drawFunc)
		}
	}
}

func startTimer(routeData *route.Data) {
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

	go refresh(routeData)
	app.SetRoot(container, true).SetInputCapture(getInputHandler(routeData))
}
