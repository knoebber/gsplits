package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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
	category := strings.Join(os.Args[1:], " ")

	var splitNames []string

	// Try to get the split names from the database by the passed in category name.
	splitNames = getSplitNames(db, category)
	if splitNames == nil {
		// Use the category wizard to either create or get an existing category.
		splitNames = useCategory(categoryWizard(category))
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
	startSplits(splitNames)
}

func useCategory(c Category) []string {
	if c.ID == 0 {
		// Category is new and needs to be saved.
	}
	return c.SplitNames
}

func startSplits(splits []string) {
	if len(splits) == 0 {
		panic("splits is empty!")
	}

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
			fmt.Printf("\n%s -> %s\n", splits[i], t)
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
