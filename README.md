# gsplits
CLI Stopwatch designed for speedrun splitting

## Install
First ensure that sqlite3 and go is installed on your system.
Then run `go get github.com/knoebber/gsplits`

This has only been tested on Unix systems - though it should work on all OS'es if you changed the path to the sqlite database.

## Usage
Run `gsplits` from a shell. It will walk you through setting up a category and a route.

After routes are setup, you can go to the route directly by passing a routename to gsplits. It will search for names that match.
In the graphical views, you can scroll tables with the arrow keys or 'j' and 'k'. Use tab to cycle through buttons and enter to select.

On the timer view, press <space> to advance the split. If you advance accidently, use <ctrl-space> to go back one.
Push 'r' to reset the run at anytime.

## Example run output

![example_run](https://github.com/knoebber/gsplits/blob/master/example_run.png)
