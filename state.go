package main

import (
	"time"

	"github.com/gdamore/tcell"
	"github.com/knoebber/gsplits/route"
	"github.com/rivo/tview"
)

const placeholder = "___"

type timerState struct {
	routeData *route.Data

	sumOfGold *time.Duration

	splitIndex    int
	runStart      time.Time
	segmentStart  time.Time
	segments      []time.Duration
	totalDuration time.Duration

	splitsTable          *tview.Table
	totalTimeView        *tview.TextView
	segmentTimeView      *tview.TextView
	goldView             *tview.TextView
	possibleTimeSaveView *tview.TextView
	bestPossibleTimeView *tview.TextView
	sumOfGoldView        *tview.TextView
}

// Called when the run is completed.
func (t *timerState) setTotalDuration() {
	if t.totalDuration == 0 {
		t.totalDuration = time.Since(t.runStart)
	}
}

func (t *timerState) reset() {
	for i := range t.segments {
		t.segments[i] = 0
	}

	t.splitIndex = 0
	t.totalDuration = 0
	now := time.Now()
	t.runStart = now
	t.segmentStart = now

	if t.routeData.SumOfGold != nil {
		sob := *t.routeData.SumOfGold
		t.sumOfGold = &sob
	}

	t.setSplitsTable()
}

func (t *timerState) setTableCell(col int, value string, color tcell.Color) {
	setTableCell(t.splitsTable, t.splitIndex, col, value, color)
}

func (t *timerState) getPlusMinus(total time.Duration) (plusMinus string, color tcell.Color, show bool) {
	var lastDiff time.Duration

	diff := total - t.routeData.GetComparisonSplit(t.splitIndex)

	plusMinus = durationStr(diff)
	if diff <= 0 {
		color = tcell.ColorGreen
	} else {
		color = tcell.ColorRed
	}

	if t.splitIndex > 0 {
		lastSplit := t.segmentStart.Sub(t.runStart)
		lastDiff = (lastSplit - t.routeData.GetComparisonSplit(t.splitIndex-1))
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

func (t *timerState) setSplitsTable() {
	if t.splitsTable == nil {
		t.splitsTable = newTable()
	}

	for i := range t.routeData.SplitNames {
		for j, value := range []string{
			t.routeData.GetSplitName(i),
			placeholder,
			durationStr(t.routeData.GetComparisonSegment(i)),
			durationStr(t.routeData.GetComparisonSplit(i)),
		} {
			setTableCell(t.splitsTable, i, j, value, tcell.ColorDefault)
		}
	}
}

func (t *timerState) createLayout() *tview.Flex {
	totalTime := tview.NewFlex().
		AddItem(newText("Total Time"), 0, 1, false).
		AddItem(t.totalTimeView, 0, 1, false).
		AddItem(nil, 0, 1, false)

	segmentTime := tview.NewFlex().
		AddItem(newText("Segment Time"), 0, 1, false).
		AddItem(t.segmentTimeView, 0, 1, false).
		AddItem(nil, 0, 1, false)

	bestPossibleTime := tview.NewFlex().
		AddItem(newText("Best Possible Time"), 0, 1, false).
		AddItem(t.bestPossibleTimeView, 0, 1, false).
		AddItem(nil, 0, 1, false)

	gold := tview.NewFlex().
		AddItem(newText("Gold"), 0, 1, false).
		AddItem(t.goldView, 0, 1, false).
		AddItem(nil, 0, 1, false)

	possibleTimeSave := tview.NewFlex().
		AddItem(newText("Possible Time Save"), 0, 1, false).
		AddItem(t.possibleTimeSaveView, 0, 1, false).
		AddItem(nil, 0, 1, false)

	sumOfGold := tview.NewFlex().
		AddItem(newText("Sum of Gold"), 0, 1, false).
		AddItem(t.sumOfGoldView, 0, 1, false).
		AddItem(nil, 0, 1, false)

	splitsFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(totalTime, 1, 0, false).
		AddItem(segmentTime, 1, 0, false).
		AddItem(bestPossibleTime, 1, 0, false).
		AddItem(gold, 1, 0, false).
		AddItem(possibleTimeSave, 1, 0, false).
		AddItem(sumOfGold, 1, 0, false)

	return tview.NewFlex().SetDirection(tview.FlexRow).SetFullScreen(true).
		AddItem(t.splitsTable, 0, 10, true).
		AddItem(nil, 0, 1, false).
		AddItem(splitsFlex, 0, 4, false)
}

func (t *timerState) getDrawFunc() func() {
	return func() {

		runDuration := time.Since(t.runStart)
		lastSplit := t.segmentStart.Sub(t.runStart)
		diff := runDuration - t.routeData.GetComparisonSplit(t.splitIndex)

		// Draw the current split row.
		plusMinus, plusMinusColor, showPlusMinus := t.getPlusMinus(runDuration)

		splitName := t.routeData.GetSplitName(t.splitIndex)
		setTableCell(t.splitsTable, t.splitIndex, 0, splitName, tcell.ColorYellow)
		if showPlusMinus {
			setTableCell(t.splitsTable, t.splitIndex, 1, plusMinus, plusMinusColor)
		}

		t.totalTimeView.SetText(elapsedStr(t.runStart))
		t.segmentTimeView.SetText(elapsedStr(t.segmentStart))
		t.goldView.SetText(durationStr(t.routeData.GetGold(t.splitIndex)))
		t.possibleTimeSaveView.SetText(durationStr(t.routeData.GetTimeSave(t.splitIndex)))
		t.bestPossibleTimeView.SetText(durationStr(t.routeData.GetBPT(t.splitIndex, lastSplit, diff)))
		t.sumOfGoldView.SetText(safeDurationStr(t.sumOfGold))
	}
}

func (t *timerState) isDone() bool {
	// The run is finished once the last segment time is filled in.
	return t.segments[len(t.segments)-1] != 0
}

func newTimerState(routeData *route.Data) *timerState {
	now := time.Now()

	t := &timerState{
		routeData:            routeData,
		splitIndex:           0,
		runStart:             now,
		segmentStart:         now,
		segments:             make([]time.Duration, routeData.Length),
		totalTimeView:        newText(elapsedStr(now)),
		segmentTimeView:      newText(elapsedStr(now)),
		goldView:             newText(durationStr(routeData.GetGold(0))),
		possibleTimeSaveView: newText(durationStr(routeData.GetTimeSave(0))),
		bestPossibleTimeView: newText(durationStr(routeData.GetGold(0))),
		sumOfGoldView:        newText(safeDurationStr(routeData.SumOfGold)),
	}
	if routeData.SumOfGold != nil {
		sob := *routeData.SumOfGold
		t.sumOfGold = &sob
	}

	t.setSplitsTable()
	return t
}
