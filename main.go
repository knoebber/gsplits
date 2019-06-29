package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

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
	// Wait for the user to push enter.
	enter := make(chan bool)
	go waitForEnter(enter)

	start := time.Now()
	var elapsed time.Duration
	var t string

	splits := []string{
		"1 star battlefield",
		"5 stars whomps",
		"8 stars snow",
		"dark world",
		"10 star fire",
		"11 star sand",
		"15 star underground",
		"16 star sub",
		"fire sea",
		"sky world",
	}
	i := 0

	fmt.Println("")
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
