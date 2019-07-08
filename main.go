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

// Start showing time save within this many nano seconds
// const plusMinusThreshold = (5 * 1e9) * -1

// The amount of padding characters to put around times.
const timePadding = 13

// Colors
const (
	aheadColor  = "\033[1;32m%s\033[0m"
	behindColor = "\033[1;31m%s\033[0m"
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
	hundreths := strconv.Itoa(int(d.Nanoseconds()/1e7) % 100)
	if len(hundreths) == 1 {
		hundreths = "0" + hundreths
	}

	return fmt.Sprintf("%s%s%s.%s",
		nToTime(int(d.Hours())%24),
		nToTime(int(d.Minutes())%60),
		seconds,
		hundreths,
	)
}

func printSplits(r *model.Route) {
	fmt.Printf("\n%s %s %s\n", divider, r.Name, divider)
	for i, s := range r.Splits {
		fmt.Printf("%d). %s\n", i+1, s.Name)
	}
}

func printInfo(r *model.Route) {
	printSplits(r)
	fmt.Print("\n\n")
	if r.Category.Best != nil {
		fmt.Printf("%s Best: %s\n",
			r.Category.Name,
			formatTimeElapsed(time.Duration(*r.Category.Best*1e6)),
		)
		if r.Best != nil && *r.Category.Best < *r.Best {
			fmt.Printf("%s Best: %s\n",
				r.Name,
				formatTimeElapsed(time.Duration(*r.Best*1e6)),
			)
		}
	}
	if r.SumOfGold != nil {
		fmt.Printf("Sum of gold splits: %s\n",
			formatTimeElapsed(time.Duration(*r.SumOfGold*1e6)),
		)
	}
	fmt.Print("\n\n")
}

// routeBestTotal is in milleseconds.
func formatTimePlusMinus(currentRunTotal time.Duration, routeBestTotal int64) string {
	diff := int64(currentRunTotal) - (routeBestTotal * 1e6)
	// if diff < plusMinusThreshold {
	//	return ""
	// }

	if diff == 0 {
		return strconv.Itoa(0)
	} else if diff < 0 {
		return fmt.Sprintf(aheadColor, "-"+formatTimeElapsed(time.Duration(diff*-1)))
	} else {
		return fmt.Sprintf(behindColor, "+"+formatTimeElapsed(time.Duration(diff)))
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

	printInfo(r)
	exitWhenNo("Start? ")
	startSplits(r, db)
}

func startSplits(r *model.Route, db *sql.DB) {
	var (
		totalElapsed   time.Duration
		splitElapsed   time.Duration
		statusLine     string
		goldSplit      string
		timePlusMinus  string
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

	if r.Splits[0].RouteBestSplit != nil {
		routeBestTotal = *r.Splits[0].RouteBestSplit
	}

	for {
		if r.Splits[i].GoldSplit != nil {
			goldSplit = formatTimeElapsed(time.Duration(*r.Splits[i].GoldSplit * 1e6))
		} else {
			goldSplit = "N/A"
		}
		if routeBestTotal > 0 {
			timePlusMinus = formatTimePlusMinus(totalElapsed, routeBestTotal)
		} else {
			timePlusMinus = ""
		}
		statusLine = fmt.Sprintf("== %-*s == %-*s\t%-*s\t||| Split => %-*s Gold => %-*s",
			r.MaxNameWidth,
			r.Splits[i].Name,
			timePadding,
			timePlusMinus,
			timePadding,
			formatTimeElapsed(totalElapsed),
			timePadding,
			formatTimeElapsed(splitElapsed),
			timePadding,
			goldSplit,
		)
		// Proccess milliseconds and enter presses async.
		select {
		case <-time.After(time.Millisecond):
			totalElapsed = time.Since(start)
			splitElapsed = time.Since(lastSplitEnd)
			fmt.Print(statusLine + "\r") // Carriage return to stay on same line.
		case <-enter:
			fmt.Print(statusLine + "\n")
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
			if r.Splits[0].RouteBestSplit != nil {
				routeBestTotal += *r.Splits[i].RouteBestSplit
			}
			go waitForEnter(enter)
		}
	}
}
