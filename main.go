package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/knoebber/gsplits/db"
	"github.com/knoebber/gsplits/route"
	"github.com/knoebber/gsplits/split"
)

// Start showing time save within this many nano seconds
const plusMinusThreshold = (10 * 1e9) * -1

// The amount of padding characters to put around times.
const timePadding = 13

// formatter for adding a timestamp with a color and a padding.
const (
	aheadColor  = "\033[1;32m%-*s\033[0m"
	behindColor = "\033[1;31m%-*s\033[0m"
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

func printSplits(routeName string, splitNames []split.Name) {
	fmt.Printf("\n%s %s %s\n", divider, routeName, divider)
	for i, ns := range splitNames {
		fmt.Printf("%d). %s\n", i+1, ns.Name)
	}
}

func printInfo(d *route.Data) {
	printSplits(d.RouteName, d.SplitNames)
	fmt.Print("\n\n")
	if d.Category.Best != nil {
		fmt.Printf("%s Best: %s\n",
			d.Category.Name,
			formatTimeElapsed(*d.Category.Best),
		)
		if d.RouteBestTime != nil && *d.Category.Best < *d.RouteBestTime {
			fmt.Printf("%s Best: %s\n",
				d.RouteName,
				formatTimeElapsed(*d.RouteBestTime),
			)
		}
	}
	if d.SumOfGold != nil {
		fmt.Printf("Sum of gold splits: %s\n", formatTimeElapsed(*d.SumOfGold))
	}
	fmt.Print("\n\n")
}

func formatTimePlusMinus(totalElapsed time.Duration, routeBestTotal time.Duration, current bool) string {
	diff := totalElapsed - routeBestTotal
	if (current && diff < plusMinusThreshold) || routeBestTotal < 1 {
		return strings.Repeat(" ", timePadding)
	}

	if diff == 0 {
		return strconv.Itoa(0)
	} else if diff < 0 {
		return fmt.Sprintf(aheadColor, timePadding, "-"+formatTimeElapsed(time.Duration(diff*-1)))
	} else {
		return fmt.Sprintf(behindColor, timePadding, "+"+formatTimeElapsed(time.Duration(diff)))
	}
}

func printStatusLine(
	maxNameWidth int,
	splitName string,
	routeBestTotal time.Duration,
	totalElapsed time.Duration,
	splitElapsed time.Duration,
	goldSplit string,
	current bool,
) {

	timePlusMinus := formatTimePlusMinus(totalElapsed, routeBestTotal, current)

	var endChar rune
	// While its the current print a carriage return so it stays on the same line
	if current {
		endChar = '\r'
	} else {
		endChar = '\n'
	}

	fmt.Printf("== %-*s == %s %-*s||| Split => %-*s Gold => %-*s%c",
		maxNameWidth,
		splitName,
		timePlusMinus,
		timePadding,
		formatTimeElapsed(totalElapsed),
		timePadding,
		formatTimeElapsed(splitElapsed),
		timePadding,
		goldSplit,
		endChar,
	)
}

func exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func main() {
	var (
		routeID int64
		err     error
		d       *route.Data
	)
	if err = db.Start(); err != nil {
		exit(err)
	}

	defer db.Close()

	routeName := strings.TrimSpace(strings.Join(os.Args[1:], " "))

	// Search for route name in database by the passed in name.
	if routeName != "" {
		routeID, err = findRoute(routeName)
		if err != nil {
			exit(err)
		}
		d, err = route.GetData(routeID)
		if err != nil {
			exit(err)
		}
	} else {
		// Use the category wizard to either create or get an existing category.
		routeID, err = wizard()
		if err != nil {
			exit(err)
		}

		d, err = route.GetData(routeID)
		if err != nil {
			exit(err)
		}
	}

	if err := startSplits(d); err != nil {
		exit(err)
	}
}

func startSplits(d *route.Data) error {
	var (
		totalElapsed   time.Duration
		splitElapsed   time.Duration
		routeBestTotal time.Duration
		goldSplit      string
		i              int
		splitTimes     []time.Duration
	)

	if len(d.SplitNames) == 0 {
		return errors.New("splits is empty")
	}

	enter := make(chan bool)
	go waitForEnter(enter)

	start := time.Now()
	lastSplitEnd := time.Now()

	splitTimes = make([]time.Duration, len(d.SplitNames))

	hasSplits := len(d.RouteBests) > 0

	if hasSplits {
		routeBestTotal = d.RouteBests[0]
	}

	for {
		if hasSplits {
			goldSplit = formatTimeElapsed(d.Golds[i])
		} else {
			goldSplit = "N/A"
		}

		// Proccess milliseconds and enter presses async.
		select {
		case <-time.After(time.Millisecond):
			totalElapsed = time.Since(start)
			splitElapsed = time.Since(lastSplitEnd)
			printStatusLine(
				d.MaxNameWidth,
				d.SplitNames[i].Name,
				routeBestTotal,
				totalElapsed,
				splitElapsed,
				goldSplit,
				true,
			)
		case <-enter:
			splitTimes[i] = splitElapsed
			lastSplitEnd = time.Now()

			printStatusLine(
				d.MaxNameWidth,
				d.SplitNames[i].Name,
				routeBestTotal,
				totalElapsed,
				splitElapsed,
				goldSplit,
				false,
			)

			i++
			if i == len(d.SplitNames) {
				fmt.Printf("\n%s\n", divider)
				fmt.Printf("FINISH! %s\n", formatTimeElapsed(totalElapsed))
				fmt.Printf("%s\n", divider)

				if promptYN("Save?") {
					_, err := saveRun(d.RouteID, splitTimes, totalElapsed)
					return err
				}
				return nil
			}
			if hasSplits {
				routeBestTotal += d.RouteBests[i]
			}
			go waitForEnter(enter)
		}
	}
}
