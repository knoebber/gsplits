package main

import (
	"database/sql"
	// Driver for sql
	_ "github.com/mattn/go-sqlite3"
	"time"
)

// Category models a speed running category.
// Example: Mario 64 16 star
type Category struct {
	ID         int64
	Name       string
	SplitNames []string
}

// Run models a single run of category.
type Run struct {
	ID           int64
	CategoryID   int64
	Milliseconds int64
	CreatedAt    time.Time
}

// SplitName models the names and order of splits in a category.
type SplitName struct {
	ID         int64
	CategoryID int64
	Position   int64
	Name       string
}

// Split models the duration of splits in a run.
type Split struct {
	ID           int64
	RunID        int64
	SplitNameID  int64
	Milliseconds int64
}

// Opens or creates a sqlite3 database.
func initDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./splits.db")
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
                        id INTEGER PRIMARY KEY,
	                name TEXT
                 );`,
		`CREATE TABLE IF NOT EXISTS run(
                        id INTEGER PRIMARY KEY,
                        category_id INTEGER,
                        milliseconds int64,
	                created_at DATETIME DEFAULT CURRENT_TIMESTAMP
                 );`,
		`CREATE TABLE IF NOT EXISTS split_name(
                        id INTEGER PRIMARY KEY,
		        category_id INTEGER,
		        position INTEGER,
                        name string
                 );`,
		`CREATE TABLE IF NOT EXISTS split(
                        run_id INTEGER PRIMARY KEY,
                        split_name_id INTEGER,
                        milliseconds INTEGER
                 );`,
	}

	for _, table := range tables {
		_, err := db.Exec(table)
		if err != nil {
			panic(err)
		}
	}

}

// Returns an order list of split names for the category.
func getSplitNames(db *sql.DB, category string) []string {
	query := `
        SELECT sn.name
        FROM split_name AS sn
        JOIN category AS c ON c.id = sn.category_id
        WHERE c.name = ?
        ORDER BY sn.position`

	rows, err := db.Query(query, category)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []string
	var n string
	for rows.Next() {
		if err := rows.Scan(&n); err != nil {
			panic(err)
		}
		result = append(result, n)
	}
	return result
}

func setupCategory(db *sql.DB, category string) []string {
	return nil
}
