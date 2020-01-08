package route

import (
	"database/sql"
	"fmt"

	"github.com/knoebber/gsplits/db"
)

// Name is the routes name.
type Name struct {
	ID         int64
	CategoryID int64  `validate:"required"`
	Name       string `validate:"required"`
}

func (r Name) String() string {
	return fmt.Sprintf("route %s", r.Name)
}

// Save inserts the route into the route table.
func (r *Name) Save(tx *sql.Tx) (sql.Result, error) {
	if err := db.Validate(r); err != nil {
		return nil, err
	}

	return tx.Exec("INSERT INTO route(name, category_id) VALUES(?,?)", r.Name, r.CategoryID)
}

// SearchNames finds route names that contain q.
func SearchNames(q string) ([]Name, error) {
	search := "%" + q + "%"
	rows, err := db.Connection.Query(`SELECT id, name FROM route WHERE name like ? ORDER BY id`, search)
	if err != nil {
		return nil, fmt.Errorf("failed to search routes: %w", err)
	}

	return getNames(rows)
}

// GetByCategory returns a list routes names that are in the category.
func GetByCategory(categoryID int64) ([]Name, error) {
	rows, err := db.Connection.Query(`SELECT id, name FROM route WHERE category_id = ? ORDER BY id`, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get routes: %w", err)
	}
	return getNames(rows)
}

func getNames(rows *sql.Rows) ([]Name, error) {
	defer rows.Close()
	var result []Name

	for rows.Next() {
		curr := Name{}
		if err := rows.Scan(&curr.ID, &curr.Name); err != nil {
			return nil, err
		}
		result = append(result, curr)
	}
	return result, nil
}
