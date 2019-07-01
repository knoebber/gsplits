package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Walks the user through setting up a new category.
func categoryWizard(category string) Category {
	scanner := bufio.NewScanner(os.Stdin)
	// TODO prompt user for existing categories.
	if category == "" {
		fmt.Println("No category was provided, example:\n$ gsplits mario64 16 star\n")
		fmt.Println("Setup new category?")
		promptYN()
		fmt.Print("New category name: ")
		scanner.Scan()
		category = scanner.Text()
	} else {
		fmt.Printf("category '%s' not found\n\n", category)
	}
	fmt.Printf("Set up new category '%s'?\n", category)
	promptYN()
	var splitNames []string
	var splitName string
	i := 0
	fmt.Print("\n\n====================\n\n")
	fmt.Println("Input the names of each split")
	fmt.Println("Push enter without input to finish")
	for {
		fmt.Printf("%d.) ", i)
		scanner.Scan()
		splitName = scanner.Text()
		i++
		if splitName == "" {
			break
		}
		splitNames = append(splitNames, splitName)
	}
	printCategorySplits(category, splitNames)
	fmt.Println("save?")
	promptYN()
	fmt.Println("start?")
	promptYN()

	return Category{
		Name:       category,
		SplitNames: splitNames,
	}
}

func printCategorySplits(category string, names []string) {
	fmt.Printf("\n==== %s ====\n", category)
	for i, name := range names {
		fmt.Printf("%d.) %s\n", i, name)
	}
}

func promptYN() {
	var a string
	for {
		fmt.Print("(Y\\n): ")
		fmt.Scanln(&a)
		if strings.ToLower(a) == "y" {
			return
		} else if strings.ToLower(a) == "n" {
			os.Exit(0)
		}
	}
}
