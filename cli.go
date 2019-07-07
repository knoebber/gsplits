package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"gsplits/model"
	"os"
	"strconv"
	"strings"
)

const divider = "=========="

func setupNewCategory(db *sql.DB) *model.Category {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("New category name: ")
	scanner.Scan()
	c := &model.Category{Name: scanner.Text()}
	c.Save(db)
	return c
}

func setupNewRoute(db *sql.DB, c *model.Category) *model.Route {
	var (
		splitNames   []*model.SplitName
		splitName    *model.SplitName
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
	for {
		fmt.Printf("%d.) ", i+1)
		scanner.Scan()
		splitName = &model.SplitName{Name: scanner.Text()}
		i++
		if splitName.Name == "" {
			break
		}
		splitNames = append(splitNames, splitName)
	}

	r := &model.Route{
		Name:     newRouteName,
		Splits:   splitNames,
		Category: c,
	}
	exitWhenNo("Save?")
	r.Save(db)
	return r
}

// Walks the user through setting up or getting a route.
func wizard(db *sql.DB, routeName string) *model.Route {
	var (
		c *model.Category
		r *model.Route
	)

	categories := model.GetCategories(db)
	if len(categories) == 0 || !promptYN("Use existing category?") {
		c = setupNewCategory(db)
	} else {
		fmt.Println("Choose a category")
		for i, category := range categories {
			fmt.Printf("(%d) %s\n", i+1, category.Name)
		}
		c = &categories[promptListSelect(len(categories))]
	}

	routes := model.GetRoutesByCategory(db, c.ID)

	if len(routes) == 0 || !promptYN("Use existing route?") {
		r = setupNewRoute(db, c)
	} else {
		fmt.Println("Choose a route")
		for i, route := range routes {
			fmt.Printf("(%d) %s\n", i+1, route.Name)
		}
		r = model.GetRoute(db, routes[promptListSelect(len(routes))].ID, "")
		if r == nil {
			panic("route is nil")
		}
	}
	return r
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
