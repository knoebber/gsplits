package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/knoebber/gsplits/db"
	"github.com/knoebber/gsplits/route"
	"github.com/rivo/tview"
)

var app *tview.Application

func exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func main() {
	var (
		routeID   int64
		err       error
		routeData *route.Data
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
		routeData, err = route.GetData(routeID)
		if err != nil {
			exit(err)
		}
	} else {
		// Use the category wizard to either create or get an existing category.
		routeID, err = wizard()
		if err != nil {
			exit(err)
		}

		routeData, err = route.GetData(routeID)
		if err != nil {
			exit(err)
		}
	}

	app = tview.NewApplication()
	if err := showPreview(routeData); err != nil {
		exit(err)
	}
}
