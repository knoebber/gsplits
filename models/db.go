package models

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	// Driver for sql
	_ "github.com/mattn/go-sqlite3"
)

// The name of the sqlite3 db file.
// Created as a hidden file in the home directory: ~/.gsplits
const dbName = "gsplits"

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

// Rolls back a database transaction and panics.
func rollback(tx *sql.Tx, err error) {
	fmt.Println("Rolling back transaction")
	if err := tx.Rollback(); err != nil {
		panic(errors.New("failed to rollback"))
	}
	panic(err)
}

func lastId(res sql.Result, tx *sql.Tx) int64 {
	id, err := res.LastInsertId()
	if err != nil {
		rollback(tx, err)
	}
	return id
}
