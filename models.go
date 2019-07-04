package main

import (
	"database/sql"
	"fmt"
	"os"
	// Driver for sql
	_ "github.com/mattn/go-sqlite3"
	"time"
)

const dbName = "gsplits"

// Category models a speed running category.
// Example: Mario 64 16 star
type Category struct {
	ID   int64
	Name string
}

// Route is a collection of split names for a category.
type Route struct {
	ID         int64
	Name       string
	CategoryID int64
	Splits     []SplitName
}

// SplitName is the name of a split in a route.
type SplitName struct {
	ID       int64
	RouteID  int64
	Position int64
	Name     string
}

// Run is a single run of a route.
type Run struct {
	ID           int64
	RouteID      int64
	Milliseconds int64
	CreatedAt    time.Time
	Splits       []Split
}

// Split is a single split in a run.
type Split struct {
	ID           int64
	RunID        int64
	SplitNameID  int64
	Milliseconds int64
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

func createCategory(db *sql.DB, c *Category) *Category {
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
		name   string
		id     int64
	)

	rows, err := db.Query("SELECT id, name FROM category")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&id, &name); err != nil {
			panic(err)
		}
		result = append(result, Category{Name: name, ID: id})
	}

	return result
}

func createRoute(db *sql.DB, r *Route) *Route {
	if len(r.Splits) == 0 {
		panic("cannot create a route without splits")
	}

	res, err := db.Exec("INSERT INTO route(name, category_id) VALUES(?,?)", r.Name, r.CategoryID)
	if err != nil {
		panic(err)
	}

	routeID, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}

	for i, sn := range r.Splits {
		_, err = db.Exec("INSERT INTO split_name(route_id, position, name) VALUES (?, ?, ?)",
			routeID, i, sn.Name)

		if err != nil {
			panic(err)
		}
	}
	r.ID = routeID
	return r
}

// Gets all routes that are in a category.
func getRoutesByCategory(db *sql.DB, categoryID int64) []*Route {
	var (
		result    []*Route
		route     *Route
		routeID   int64
		splitID   int64
		routeName string
		splitName string
		exists    bool
	)

	query := `
        SELECT r.id, 
               r.Name, 
               sn.Name AS split_name, 
               sn.id AS split_id
        FROM route AS r
        JOIN split_name AS sn ON sn.route_id = r.id
        WHERE category_id = ?
        ORDER BY r.id, sn.position`

	routeMap := map[int64]*Route{}

	rows, err := db.Query(query, categoryID)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&routeID, &routeName, &splitName, &splitID); err != nil {
			panic(err)
		}
		route, exists = routeMap[routeID]
		if !exists {
			route = &Route{
				ID:     routeID,
				Name:   routeName,
				Splits: []SplitName{{Name: splitName}},
			}
			result = append(result, route)
			routeMap[routeID] = route
		} else {
			route.Splits = append(route.Splits, SplitName{Name: splitName, ID: splitID})
		}
	}

	return result
}

// Get a route by its name.
// Returns nil if the route isn't found.
func getRoute(db *sql.DB, name string) *Route {
	var (
		splits    []SplitName
		splitName string
		routeName string
		id        int64
		splitID   int64
	)
	query := `
        SELECT r.id, 
               r.name AS route_name,
               sn.name AS split_name, 
               sn.id AS split_id
        FROM split_name AS sn
        JOIN route AS r ON r.id = sn.route_id
        WHERE r.name = ? COLLATE NOCASE
        ORDER BY sn.position`

	rows, err := db.Query(query, name)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&id, &routeName, &splitName, &splitID); err != nil {
			panic(err)
		}
		splits = append(splits, SplitName{Name: splitName, ID: splitID})
	}

	if len(splits) == 0 {
		return nil
	}
	return &Route{
		ID:     id,
		Name:   routeName,
		Splits: splits,
	}
}

func saveRun(db *sql.DB, r *Run) *Run {
	res, err := db.Exec("INSERT INTO run(route_id, milliseconds) VALUES(?, ?)",
		r.RouteID, r.Milliseconds)
	if err != nil {
		panic(err)
	}

	runID, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}
	r.ID = runID

	for _, s := range r.Splits {
		_, err = db.Exec("INSERT INTO split(run_id, split_name_id, milliseconds) VALUES (?, ?, ?)",
			runID, s.SplitNameID, s.Milliseconds)

		if err != nil {
			panic(err)
		}
	}
	return r
}
