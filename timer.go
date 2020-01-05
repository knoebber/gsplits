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
	splitsTable      *tview.Table
	currentGold      time.Duration
	possibleTimeSave time.Duration

	totalTimeView        *tview.TextView
	segmentTimeView      *tview.TextView
	goldView             *tview.TextView
	possibleTimeSaveView *tview.TextView
	bestPossibleTimeView *tview.TextView

	splitIndex int
)

func nextSplit() {
	segmentStart = time.Now()

	splitTime := time.Since(runStart)
	setTableCell(splitsTable, splitIndex, 3, durationStr(splitTime), tcell.ColorWhite)
	splitIndex++
}

// Returning nil stops the input from propagating.
func inputHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case ' ':
		nextSplit()
		return nil
	}

	switch event.Key() {
	case tcell.KeyEnter:
		nextSplit()
		return nil
	}
	return event
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

// Draws an active row.
func drawCurrentSplitRow(routeData *route.Data, runDuration time.Duration) {
	var (
		plusMinus      string
		plusMinusColor tcell.Color
	)
	diff := runDuration - routeData.Comparison[splitIndex]
	plusMinus = durationStr(diff)
	if diff <= 0 {
		plusMinusColor = tcell.ColorGreen
	} else {
		plusMinusColor = tcell.ColorRed
	}

	setTableCell(splitsTable, splitIndex, 0, routeData.SplitNames[splitIndex].Name, tcell.ColorYellow)
	if diff > plusMinusThreshold {
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

func liveTimeDrawer(routeData *route.Data) func() {
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
	drawFunc := liveTimeDrawer(routeData)

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
	app.SetRoot(container, true).SetInputCapture(inputHandler)
}
