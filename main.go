package main

import (
	"database/sql"
	"fmt"
	"gsplits/model"
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
	hs := strconv.Itoa(int(d.Nanoseconds()/1e7) % 100)
	if len(hs) == 1 {
		hs = "0" + hs
	}

	return fmt.Sprintf("%s%s%s.%d",
		nToTime(int(d.Hours())%24),
		nToTime(int(d.Minutes())%60),
		seconds,
		hs,
	)
}

// routeBestTotal is in milleseconds.
func formatTimePlusMinus(currentRunTotal time.Duration, routeBestTotal int64) string {
	diff := int64(currentRunTotal) - (routeBestTotal * 1e6)
	if diff == 0 {
		return strconv.Itoa(0)
	} else if diff < 0 {
		return "-" + formatTimeElapsed(time.Duration(diff*-1))
	} else {
		return "+" + formatTimeElapsed(time.Duration(diff))
	}
}

func main() {
	db := model.InitDB()
	defer db.Close()

	route := strings.TrimSpace(strings.Join(os.Args[1:], " "))

	var r *model.Route

	// Try to get the route from the database by the passed in name.
	// TODO get category as well.
	r = model.GetRoute(db, 0, route)
	if r == nil {
		if route != "" {
			fmt.Printf("route '%s' not found\n", route)
		}
		// Use the category wizard to either create or get an existing category.
		r = wizard(db, route)
	}

	printRouteInfo(r)
	exitWhenNo("Start? ")
	fmt.Printf("\n%s %s %s\n\n", divider, r.Name, divider)
	startSplits(r, db)
}

func startSplits(r *model.Route, db *sql.DB) {
	var (
		totalElapsed   time.Duration
		splitElapsed   time.Duration
		statusLine     string
		routeBestTotal int64
		i              int
	)

	if len(r.Splits) == 0 {
		panic("splits is empty!")
	}

	enter := make(chan bool)
	go waitForEnter(enter)

	start := time.Now()
	lastSplitEnd := time.Now()

	// The database model for saving the run.
	run := &model.Run{
		Route:  r,
		Splits: make([]*model.Split, len(r.Splits)),
	}

	routeBestTotal = r.Splits[0].RouteBestSplit
	for {

		statusLine = fmt.Sprintf("%s => %s Total: %s\tSplit: %s\tGold: %s",
			r.Splits[i].Name,
			formatTimePlusMinus(totalElapsed, routeBestTotal),
			formatTimeElapsed(totalElapsed),
			formatTimeElapsed(splitElapsed),
			formatTimeElapsed(time.Duration(r.Splits[i].GoldSplit)),
		)
		// Proccess milliseconds and enter presses async.
		select {
		case <-time.After(time.Millisecond):
			totalElapsed = time.Since(start)
			splitElapsed = time.Since(lastSplitEnd)
			fmt.Print(statusLine + "\r")
		case <-enter:
			fmt.Print(statusLine)
			fmt.Printf("%s\n", divider)

			run.Splits[i] = &model.Split{
				SplitNameID:  r.Splits[i].ID,
				Milliseconds: splitElapsed.Nanoseconds() / 1e6,
			}

			lastSplitEnd = time.Now()

			i++

			if i == len(r.Splits) {
				fmt.Printf("\n%s\n", divider)
				fmt.Printf("FINISH! %s\n", formatTimeElapsed(totalElapsed))
				fmt.Printf("%s\n", divider)

				if promptYN("Save?") {
					run.Milliseconds = totalElapsed.Nanoseconds() / 1e6
					run.Save(db)
				}
				os.Exit(0)
			}
			routeBestTotal += r.Splits[i].RouteBestSplit
			go waitForEnter(enter)
		}
	}
}
