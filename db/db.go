package db

import (
	"database/sql"
	"fmt"
	"os"
	// Driver for sql
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/go-playground/validator.v9"
)

// Connection is a sql connection.
var Connection *sql.DB

var validate *validator.Validate

// The name of the sqlite3 db file.
// Created as a hidden file in the home directory: ~/.gsplits
const dbName = "gsplits-test"

// Creates the required tables if they doesn't exist
func createTables() error {
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
		_, err := Connection.Exec(table)
		if err != nil {
			return fmt.Errorf("failed to create tables: %w", err)
		}
	}
	return nil
}

// Start opens a connection a sqlite3 database.
// It will create a new database if the db file does not exist.
func Start() error {
	var (
		home string
		err  error
	)

	home, err = os.UserHomeDir()
	if err != nil {
		return err
	}

	Connection, err = sql.Open("sqlite3", fmt.Sprintf("%s/.%s.db", home, dbName))
	if err != nil {
		return fmt.Errorf("failed to open sqlite datebase: %w", err)
	}

	validate = validator.New()
	return createTables()
}

// Close closes the connection.
func Close() {
	Connection.Close()
}

// Validate validates a struct.
// Should be used to check values before insertion.
func Validate(s interface{}) error {
	return validate.Struct(s)
}

// Rollback rolls back a database transaction.
// It always returns an error.
func Rollback(tx *sql.Tx, err error) error {
	if rbError := tx.Rollback(); rbError != nil {
		return fmt.Errorf("failed to rollback database transaction: %w", rbError)
	}
	return fmt.Errorf("rolled back from error: %w", err)
}
