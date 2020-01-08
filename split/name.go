package split

import (
	"database/sql"
	"fmt"

	"github.com/knoebber/gsplits/db"
)

// Name is the name of a split in a route.
type Name struct {
	ID       int64
	RouteID  int64  `validate:"required"`
	Position int    `validate:"required"`
	Name     string `validate:"required"`
}

func (n Name) String() string {
	return fmt.Sprintf("split name %#v", n.Name)
}

// Save saves the name into the split_name table.
func (n *Name) Save(tx *sql.Tx) (sql.Result, error) {
	if err := db.Validate(n); err != nil {
		return nil, err
	}
	return tx.Exec(
		"INSERT INTO split_name(route_id, position, name) VALUES (?, ?, ?)",
		n.RouteID,
		n.Position,
		n.Name,
	)
}

// GetByRoute returns a list of all the split names in the route.
// The result is ordered by position.
func GetByRoute(routeID int64) ([]Name, error) {

	rows, err := db.Connection.
		Query("SELECT id, name FROM split_name WHERE route_id = ? ORDER BY position", routeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get split names: %w", err)
	}

	defer rows.Close()

	result := []Name{}
	for rows.Next() {
		curr := Name{}
		if err := rows.Scan(
			&curr.ID,
			&curr.Name,
		); err != nil {
			return nil, err
		}
		result = append(result, curr)
	}
	return result, nil
}
