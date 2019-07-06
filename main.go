package main

import (
	"database/sql"
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

func formatTimeElapsed(d time.Duration) string {
	seconds := strconv.Itoa(int(d.Seconds()) % 60)
	if len(seconds) == 1 {
		seconds = "0" + seconds
	}

	return fmt.Sprintf("%s%s%s.%d",
		nToTime(int(d.Hours())%24),
		nToTime(int(d.Minutes())%60),
		seconds,
		int(d.Nanoseconds()/1e7)%100) // Show hundreths of a second.
}

func main() {
	db := initDB()
	defer db.Close()

	route := strings.Join(os.Args[1:], " ")

	var r *Route

	// Try to get the route from the database by the passed in name.
	// TODO get category as well.
	r = getRoute(db, route)
	if r == nil {
		if route != "" {
			fmt.Printf("route '%s' not found\n", route)
		}
		// Use the category wizard to either create or get an existing category.
		r = wizard(db, route)
	}

	printRouteSplits(r)
	exitWhenNo("Start? ")
	fmt.Printf("\n%s %s %s\n\n", divider, r.Name, divider)
	startSplits(r, db)
}

func startSplits(r *Route, db *sql.DB) {
	var (
		totalElapsed time.Duration
		splitElapsed time.Duration
		total        string
		split        string
		i            int
	)

	if len(r.Splits) == 0 {
		panic("splits is empty!")
	}

	enter := make(chan bool)
	go waitForEnter(enter)

	start := time.Now()
	lastSplitEnd := time.Now()

	// The database model for saving the run.
	run := &Run{
		Route:  r,
		Splits: make([]Split, len(r.Splits)),
	}

	for {
		total = formatTimeElapsed(totalElapsed)
		split = formatTimeElapsed(splitElapsed)
		// Proccess milliseconds and enter presses async.
		select {
		case <-time.After(time.Millisecond):
			totalElapsed = time.Since(start)
			splitElapsed = time.Since(lastSplitEnd)
			fmt.Printf("TOTAL: %s\t\tSPLIT: %s\r", total, split)
		case <-enter:
			fmt.Println(r.Splits[i].Name)
			fmt.Printf("TOTAL: %s\t\tSPLIT: %s\n", total, split)
			fmt.Printf("%s\n", divider)
			run.Splits[i] = Split{
				SplitNameID:  r.Splits[i].ID,
				Milliseconds: splitElapsed.Nanoseconds() / 1e6,
			}

			lastSplitEnd = time.Now()

			i++

			if i == len(r.Splits) {
				fmt.Printf("\n%s\n", divider)
				fmt.Printf("FINISH! %s\n", total)
				fmt.Printf("%s\n", divider)

				if promptYN("Save?") {
					run.Milliseconds = totalElapsed.Nanoseconds() / 1e6
					saveRun(db, run)
				}
				os.Exit(0)
			}
			go waitForEnter(enter)
		}
	}
}
