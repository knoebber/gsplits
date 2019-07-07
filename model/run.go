package model

import (
	"database/sql"
	"errors"
	"time"
)

// Run is a single run of a route.
type Run struct {
	ID           int64
	Milliseconds int64
	CreatedAt    time.Time
	Splits       []*Split
	Route        *Route
}

// Split is a single split in a run.
type Split struct {
	ID           int64
	SplitNameID  int64
	Milliseconds int64
	Run          *Run
}

// Save inserts the run and its splits into the database.
func (r *Run) Save(db *sql.DB) {
	var (
		err error
		res sql.Result
		tx  *sql.Tx
	)
	if r.Route.ID == 0 {
		panic(errors.New("failed to save run, route id not set"))
	}
	if r.Milliseconds == 0 {
		panic(errors.New("failed to save run, milliseconds not set"))
	}

	tx, err = db.Begin()
	if err != nil {
		panic(err)
	}

	res, err = tx.Exec("INSERT INTO run(route_id, milliseconds) VALUES(?, ?)", r.Route.ID, r.Milliseconds)
	if err != nil {
		rollback(tx, err)
	}

	r.ID = lastID(res, tx)

	for _, s := range r.Splits {
		if s.SplitNameID == 0 {
			rollback(tx, errors.New("failed to save split, split_name_id not set"))
		}
		if s.Milliseconds == 0 {
			rollback(tx, errors.New("failed to save split, milliseconds not set"))
		}

		if res, err = tx.Exec("INSERT INTO split(run_id, split_name_id, milliseconds) VALUES (?, ?, ?)",
			r.ID, s.SplitNameID, s.Milliseconds); err != nil {
			rollback(tx, err)
		}
		s.ID = lastID(res, tx)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}
}
