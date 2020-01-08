package category

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/knoebber/gsplits/db"
)

// Name models a speed running category.
// Example: Mario 64 16 star
type Name struct {
	ID   int64
	Name string `validate:"required"`
	Best *time.Duration
	//	TotalRuns int64
}

func (c Name) String() string {
	return fmt.Sprintf("category %#v", c.Name)
}

// Save inserts the category into the database.
func (c *Name) Save(tx *sql.Tx) (sql.Result, error) {
	if err := db.Validate(c); err != nil {
		return nil, err
	}
	return db.Connection.Exec("INSERT INTO category(name) VALUES(?)", c.Name)
}

// All returns a slice of all saved category names.
func All() ([]Name, error) {
	var (
		result []Name
		best   *int64
	)

	query := `
        SELECT c.id, 
               c.name, 
               MIN(ifnull(run.milliseconds,0)) AS pb
        FROM category AS C
        JOIN route AS r ON r.category_id = c.id
        LEFT JOIN run ON run.route_id = r.id
        GROUP BY c.id
        ORDER BY c.id`

	rows, err := db.Connection.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		c := Name{}
		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&best,
		); err != nil {
			panic(err)
		}
		if best != nil {
			dur := time.Duration(*best * 1e6)
			c.Best = &dur
		}
		result = append(result, c)
	}

	return result, nil
}
