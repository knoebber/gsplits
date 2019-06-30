package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// On enter press erase the line above and send true to the channel.
// This lets the main loop know to split.
func waitForEnter(enter chan bool) {
	var s string
	fmt.Scanln(&s)
	upLine := "\x1b[1A"
	eraseLine := "\x1b[2K"
	fmt.Printf("%s%s", upLine, eraseLine)
	enter <- true

}

func nToTime(n int) string {
	if n == 0 {
		return ""
	}
	s := strconv.Itoa(n)
	if len(s) == 1 {
		return "0" + s + ":"
	}
	return s + ":"
}

func main() {
	db := initDB()
	category := os.Args[1]
	s := getSplitNames(db, category)
	if s == nil {
		setupCategory(db, category)
	}

	// TODO remove.
	/*
		splits := []SplitName{
			SplitName{Name: "1 star battlefield"},
			SplitName{Name: "5 stars whomps"},
			SplitName{Name: "8 stars snow"},
			SplitName{Name: "dark world"},
			SplitName{Name: "11 star fire"},
			SplitName{Name: "12 star sand"},
			SplitName{Name: "15 star underground"},
			SplitName{Name: "16 star sub"},
			SplitName{Name: "fire sea"},
			SplitName{Name: "sky world"},
		}
	*/
	startSplits(nil)
}

func startSplits(splits []SplitName) {
	enter := make(chan bool)
	go waitForEnter(enter)

	start := time.Now()

	// The duration of the run.
	var elapsed time.Duration

	// The current time string.
	var t string

	// Tracks the index of the run splits.
	var i int

	// Proccess milliseconds, enter presses, and db calls async.
	for {
		t = fmt.Sprintf("%s%s%s%d",
			nToTime(int(elapsed.Hours())%24),
			nToTime(int(elapsed.Minutes())%60),
			nToTime(int(elapsed.Seconds())%60),
			int(elapsed.Nanoseconds())/1000%100)

		select {
		case <-time.After(time.Millisecond):
			elapsed = time.Since(start)
			fmt.Printf("%s\r", t)
		case <-enter:
			fmt.Printf("\n%s -> %s\n", splits[i].Name, t)
			i++
			if i == len(splits) {
				fmt.Print("\n=================\n")
				fmt.Print(t)
				fmt.Print("\n=================\n")
				os.Exit(0)
			}
			go waitForEnter(enter)
		}
	}
}
