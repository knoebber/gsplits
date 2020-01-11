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

func (t *timerState) createLayout() *tview.Grid {
	grid := tview.NewGrid()
	grid.SetRows(15)
	grid.SetColumns(4)
	row := 0
	tableRowSpan := 10
	grid.AddItem(t.splitsTable, row, 0, tableRowSpan, 4, 0, 0, true)

	row += tableRowSpan + 1 // One extra for some space.
	for name, val := range map[string]tview.Primitive{
		"Total Time":         t.totalTimeView,
		"Segment Time":       t.segmentTimeView,
		"Best Possible Time": t.bestPossibleTimeView,
		"Gold":               t.goldView,
		"Possible Time Save": t.possibleTimeSaveView,
		"Sum Of Gold":        t.sumOfGoldView,
	} {
		grid.AddItem(newText(name), row, 0, 1, 2, 0, 0, false)
		grid.AddItem(val, row, 3, 1, 1, 0, 0, false)
		row++
	}

	return grid
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
