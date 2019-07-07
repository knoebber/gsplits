package model

import (
	"database/sql"
	"errors"
	"fmt"
)

// A route for a category.
type Route struct {
	ID        int64
	Name      string
	TotalRuns int64
	Best      int64
	SumOfGold int64
	Splits    []*SplitName
	Category  *Category
}

// SplitName is the name of a split in a route.
type SplitName struct {
	ID             int64
	Position       int64
	Name           string
	RouteBestSplit int64 // How fast the split was in the best run of the route.
	GoldSplit      int64 // The fastest the split has been completed.
	Route          *Route
}

// Save inserts the run into the database.
func (r *Route) Save(db *sql.DB) {
	if r == nil {
		panic(errors.New("route is nil"))
	}

	if len(r.Splits) == 0 || r.Category.ID == 0 || r.Name == "" {
		panic(fmt.Errorf("route has empty values: \n%+v\n", *r))
	}

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	res, err := tx.Exec("INSERT INTO route(name, category_id) VALUES(?,?)", r.Name, r.Category.ID)
	if err != nil {
		panic(err)
	}

	r.ID = lastID(res, tx)

	for i, sn := range r.Splits {
		res, err = tx.Exec("INSERT INTO split_name(route_id, position, name) VALUES (?, ?, ?)",
			r.ID, i, sn.Name)
		if err != nil {

			panic(err)
		}

		sn.ID = lastID(res, tx)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}
}

// GetRoutesByCategory returns a list of all route names and their ids for a category.
// This does not populate the rest of the route struct.
func GetRoutesByCategory(db *sql.DB, categoryID int64) []Route {
	var (
		result []Route
		r      Route
	)

	query := `
        SELECT id, name
        FROM route
        WHERE category_id = ?
        ORDER BY id`

	rows, err := db.Query(query, categoryID)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&r.ID, &r.Name); err != nil {
			panic(err)
		}
		result = append(result, r)
	}

	return result
}

// GetRoute gets a route by its primary key or its name.
// Returns nil if the route isn't found.
func GetRoute(db *sql.DB, id int64, name string) *Route {
	var (
		where string
		err   error
		rows  *sql.Rows
		sn    *SplitName
	)

	if name == "" && id == 0 {
		return nil
	}

	if id > 0 {
		where = "r.id = ?"
	} else {
		where = "r.name = ?"
	}

	query := fmt.Sprintf(`
SELECT sn.id AS split_name_id,
       sn.name AS split_name, 
       MIN(golds.milliseconds) AS gold_split,
       route_best.milliseconds AS route_best_split,
       r.id AS route_id,
       r.name AS route_name,
       COUNT(DISTINCT run.id) AS total_runs,
       c.id AS category_id,
       c.name AS category_name,
       MIN(run.milliseconds) AS pb
FROM route AS r
JOIN split_name AS sn ON sn.route_id = r.id
JOIN split AS golds ON golds.split_name_id = sn.id 
JOIN run ON run.route_id = r.id
JOIN category AS c ON c.id = r.category_id
JOIN ( SELECT r.id AS route_id,
              sn.id AS split_name_id,
              s.milliseconds
       FROM category AS c
       JOIN route AS r ON r.category_id = c.id
       JOIN run ON run.route_id = r.id
       JOIN split_name AS sn ON sn.route_id = r.id
       JOIN split AS s ON s.split_name_id = sn.id AND s.run_id = run.id
       GROUP BY r.id, sn.id
       HAVING MIN(run.milliseconds) = run.milliseconds
       ORDER BY route_id, sn.position
     ) AS route_best ON route_best.split_name_id = sn.id
WHERE %s
GROUP BY sn.id
ORDER BY r.id, sn.position
`, where)
	if id > 0 {
		rows, err = db.Query(query, id)
	} else {
		rows, err = db.Query(query, name)
	}
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	r := &Route{}
	r.Category = &Category{}
	r.Splits = []*SplitName{}

	for rows.Next() {
		sn = &SplitName{}
		if err := rows.Scan(
			&sn.ID,
			&sn.Name,
			&sn.GoldSplit,
			&sn.RouteBestSplit,
			&r.ID,
			&r.Name,
			&r.TotalRuns,
			&r.Category.ID,
			&r.Category.Name,
			&r.Category.Best,
		); err != nil {
			panic(err)
		}

		r.SumOfGold += sn.GoldSplit
		r.Best += sn.RouteBestSplit

		r.Splits = append(r.Splits, sn)
	}

	if len(r.Splits) == 0 {
		return nil
	}
	return r
}
