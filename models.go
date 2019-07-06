package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	// Driver for sql
	_ "github.com/mattn/go-sqlite3"
	"time"
)

// The name of the sqlite3 db file.
// Created as a hidden file in the home directory: ~/.gsplits
const dbName = "gsplits"

// Category models a speed running category.
// Example: Mario 64 16 star
type Category struct {
	ID        int64
	Name      string
	PB        int64 // The fastest the category has been completed.
	TotalRuns int64
}

// A route for a category.
type Route struct {
	ID        int64
	Name      string
	TotalRuns int64
	RouteBest int64 // The fastest the route has been completed.
	Splits    []*SplitName
	Category  *Category
}

// SplitName is the name of a split in a route.
type SplitName struct {
	ID          int64
	Position    int64
	Name        string
	BestInRoute int64 // How fast the split was in the best run of the route.
	GoldSplit   int64 // The fastest the split has been completed.
	Route       *Route
}

// Run is a single run of a route.
type Run struct {
	ID           int64
	Milliseconds int64
	CreatedAt    time.Time
	Splits       []Split
	Route        *Route
}

// Split is a single split in a run.
type Split struct {
	ID           int64
	SplitNameID  int64
	Milliseconds int64
	Run          *Run
}

// Opens or creates a sqlite3 database.
func initDB() *sql.DB {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("%s/.%s.db", home, dbName))
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic("db is nil")
	}
	createTables(db)
	return db
}

// Creates the required tables if they doesn't exist
func createTables(db *sql.DB) {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS category(
                        id   INTEGER PRIMARY KEY,
	                name TEXT NOT NULL UNIQUE
                 );`,
		`CREATE TABLE IF NOT EXISTS route(
                        id          INTEGER PRIMARY KEY,
                        name        TEXT NOT NULL UNIQUE,
                        category_id INTEGER
                 );`,
		`CREATE TABLE IF NOT EXISTS run(
                        id           INTEGER PRIMARY KEY,
                        route_id     INTEGER,
                        milliseconds INTEGER,
	                created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
                 );`,
		`CREATE TABLE IF NOT EXISTS split_name(
                        id       INTEGER PRIMARY KEY,
		        route_id INTEGER,
		        position INTEGER,
                        name     TEXT
                 );`,
		`CREATE TABLE IF NOT EXISTS split(
                        id            INTEGER PRIMARY KEY,
                        run_id        INTEGER,
                        split_name_id INTEGER,
                        milliseconds  INTEGER
                 );`,
	}

	for _, table := range tables {
		_, err := db.Exec(table)
		if err != nil {
			panic(err)
		}
	}

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
			&c.PB,
			&c.TotalRuns,
		); err != nil {
			panic(err)
		}
		result = append(result, c)
	}

	return result
}

// Creates a new route.
func createRoute(db *sql.DB, r *Route) *Route {
	if r == nil {
		panic(errors.New("route is nil"))
	}

	if len(r.Splits) == 0 || r.Category.ID == 0 || r.Name == "" {
		panic(fmt.Errorf("route has empty values: \n%+v\n", *r))
	}

	res, err := db.Exec("INSERT INTO route(name, category_id) VALUES(?,?)", r.Name, r.Category.ID)
	if err != nil {
		panic(err)
	}

	routeID, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}
	r.ID = routeID

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	for i, sn := range r.Splits {
		res, err = tx.Exec("INSERT INTO split_name(route_id, position, name) VALUES (?, ?, ?)",
			r.ID, i, sn.Name)
		if err != nil {

			panic(err)
		}

		splitNameID, err := res.LastInsertId()
		if err != nil {
			panic(err)
		}
		sn.ID = splitNameID
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}
	return r
}

// Gets all route names and their ids for a category.
// This does not populate the rest of the route struct.
func getRoutesByCategory(db *sql.DB, categoryID int64) []Route {
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

// Get a route and category by the route name or the id
// Returns nil if the route isn't found.
func getRoute(db *sql.DB, id int64, name string) *Route {
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
			&sn.BestInRoute,
			&r.ID,
			&r.Name,
			&r.TotalRuns,
			&r.Category.ID,
			&r.Category.Name,
			&r.Category.PB,
		); err != nil {
			panic(err)
		}
		r.Splits = append(r.Splits, sn)
	}

	if len(r.Splits) == 0 {
		return nil
	}
	return r
}

// Saves a run and its splits.
func saveRun(db *sql.DB, r *Run) *Run {
	if r.Route.ID == 0 {
		panic(errors.New("failed to save run, route id not set"))
	}
	if r.Milliseconds == 0 {
		panic(errors.New("failed to save run, milliseconds not set"))
	}

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	res, err := tx.Exec("INSERT INTO run(route_id, milliseconds) VALUES(?, ?)", r.Route.ID, r.Milliseconds)
	if err != nil {
		rollback(tx, err)
	}

	runID, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}
	r.ID = runID

	for _, s := range r.Splits {
		if s.SplitNameID == 0 {
			rollback(tx, errors.New("failed to save split, split_name_id not set"))
		}
		if s.Milliseconds == 0 {
			rollback(tx, errors.New("failed to save split, milliseconds not set"))
		}

		if _, err := tx.Exec("INSERT INTO split(run_id, split_name_id, milliseconds) VALUES (?, ?, ?)",
			runID, s.SplitNameID, s.Milliseconds); err != nil {
			rollback(tx, err)
		}
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	return r
}

// Rolls back a database transaction and panics.
func rollback(tx *sql.Tx, err error) {
	fmt.Println("Rolling back transaction")
	if err := tx.Rollback(); err != nil {
		panic(errors.New("failed to rollback"))
	}
	panic(err)
}
