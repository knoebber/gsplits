package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/knoebber/gsplits/category"
	"github.com/knoebber/gsplits/route"
)

const divider = "=========="

func getCategoryName() string {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("New category name: ")
	scanner.Scan()
	return scanner.Text()
}

func setupNewRoute(categoryID int64) (routeID int64, err error) {
	var (
		newRouteName string
	)

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("New route name: ")
	scanner.Scan()
	newRouteName = scanner.Text()

	i := 0

	fmt.Printf("%s\n", divider)
	fmt.Println("Input the names of each split")
	fmt.Println("Push enter without input to finish")
	fmt.Println("")

	splitNames := []string{}
	for {
		fmt.Printf("%d.) ", i+1)
		scanner.Scan()
		name := scanner.Text()
		i++
		if name == "" {
			break
		}
		splitNames = append(splitNames, name)
	}

	exitWhenNo("Save?")
	return saveRoute(categoryID, newRouteName, splitNames)
}

// Walks the user through setting up or getting a route.
func wizard(routeName string) (routeID int64, err error) {
	var (
		categories []category.Name
		routes     []route.Name
		categoryID int64
	)

	categories, err = category.All()
	if err != nil {
		return
	}

	if len(categories) == 0 || !promptYN("Use existing category?") {
		categoryID, err = saveCategory(getCategoryName())
		if err != nil {
			return
		}
	} else {
		fmt.Println("Choose a category")
		for i, category := range categories {
			fmt.Printf("(%d) %s\n", i+1, category.Name)
		}
		categoryID = categories[promptListSelect(len(categories))].ID
	}

	routes, err = route.GetByCategory(categoryID)
	if err != nil {
		return
	}

	if len(routes) == 0 || !promptYN("Use existing route?") {
		routeID, err = setupNewRoute(categoryID)
		if err != nil {
			return
		}
	} else {
		fmt.Println("Choose a route")
		for i, route := range routes {
			fmt.Printf("(%d) %s\n", i+1, route.Name)
		}
		routeID = routes[promptListSelect(len(routes))].ID
	}
	return
}

// Presents a prompt to the user to pick an option number.
// Assumes that the list shown is 1 indexed.
func promptListSelect(max int) int {
	var (
		a   string
		i   int
		err error
	)
	for {
		fmt.Print("Option number: ")
		fmt.Scanln(&a)
		i, err = strconv.Atoi(a)
		if err == nil && i <= max && i > 0 {
			return i - 1
		}
	}
	return 0
}

func promptYN(prompt string) bool {
	var a string
	if prompt != "" {
		fmt.Println(prompt)
	}
	for {
		fmt.Print("(Y\\n) ")
		fmt.Scanln(&a)
		if strings.ToLower(a) == "y" {
			return true
		} else if strings.ToLower(a) == "n" {
			return false
		}
	}
}

func exitWhenNo(prompt string) {
	if !promptYN(prompt) {
		os.Exit(0)
	}
}
