package models

import (
	"database/sql"
	"errors"
)

// Category models a speed running category.
// Example: Mario 64 16 star
type Category struct {
	ID        int64
	Name      string
	Best      int64
	TotalRuns int64
}

// Creates a new category.
func createCategory(db *sql.DB, c *Category) *Category {
	if c.Name == "" {
		panic(errors.New("failed to save category, name is not set"))
	}
	res, err := db.Exec("INSERT INTO category(name) VALUES(?)", c.Name)
	if err != nil {
		panic(err)
	}

	categoryID, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}
	c.ID = categoryID
	return c
}

// Gets all saved categories.
func getCategories(db *sql.DB) []Category {
	var (
		result []Category
		c      Category
	)

	query := `
        SELECT c.id, 
               c.name, 
               MIN(run.milliseconds) AS pb,
               COUNT(*) AS total_runs
        FROM category AS C
        JOIN route AS r ON r.category_id = c.id
        JOIN run ON run.route_id = r.id
        GROUP BY c.id
        ORDER BY c.id`

	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Best,
			&c.TotalRuns,
		); err != nil {
			panic(err)
		}
		result = append(result, c)
	}

	return result
}
