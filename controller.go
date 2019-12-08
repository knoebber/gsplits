package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/knoebber/gsplits/category"
	"github.com/knoebber/gsplits/db"
	"github.com/knoebber/gsplits/route"
	"github.com/knoebber/gsplits/split"
)

type saver interface {
	String() string
	Save(tx *sql.Tx) (sql.Result, error)
}

func save(s saver, tx *sql.Tx) (id int64, err error) {
	var res sql.Result

	res, err = s.Save(tx)
	if err != nil {
		return 0, saveError(s, tx, err)
	}

	id, err = res.LastInsertId()
	if err != nil {
		return 0, idError(s, tx, err)
	}
	return
}

func idError(s saver, tx *sql.Tx, err error) error {
	return fmt.Errorf("failed to get ID from %s: %w", s, db.Rollback(tx, err))
}

func saveError(s saver, tx *sql.Tx, err error) error {
	return fmt.Errorf("failed to save %s: %w", s, db.Rollback(tx, err))
}

func saveCategory(name string) (categoryID int64, err error) {
	var tx *sql.Tx

	tx, err = db.Connection.Begin()
	if err != nil {
		fmt.Errorf("failed to start save category transaction: %w", err)
	}

	categoryName := &category.Name{
		Name: name,
	}
	if categoryID, err = save(categoryName, tx); err != nil {
		return
	}

	err = tx.Commit()
	return
}

func saveRoute(categoryID int64, name string, splitNames []string) (routeID int64, err error) {
	var tx *sql.Tx

	tx, err = db.Connection.Begin()
	if err != nil {
		fmt.Errorf("failed to start save route transaction: %w", err)
	}

	routeName := &route.Name{
		Name:       name,
		CategoryID: categoryID,
	}
	routeID, err = save(routeName, tx)
	if err != nil {
		return
	}

	sn := &split.Name{
		RouteID: routeID,
	}
	for i, splitName := range splitNames {
		sn.Position = i + 1
		sn.Name = splitName

		_, err = save(sn, tx)
		if err != nil {
			return
		}
	}

	err = tx.Commit()
	return
}

func saveRun(routeID int64, durations []time.Duration, totalDuration time.Duration) (runID int64, err error) {
	var (
		tx         *sql.Tx
		splitNames []split.Name
	)

	splitNames, err = split.GetByRoute(routeID)
	if err != nil {
		return
	}

	tx, err = db.Connection.Begin()
	if err != nil {
		fmt.Errorf("failed to start save run transaction: %w", err)
	}

	run := &route.Run{
		Duration: totalDuration,
		RouteID:  routeID,
	}

	runID, err = save(run, tx)
	if err != nil {
		return
	}

	for i, duration := range durations {
		d := &split.Duration{
			RunID:    runID,
			Duration: duration,
			NameID:   splitNames[i].ID,
		}
		if _, err = save(d, tx); err != nil {
			return
		}
	}

	err = tx.Commit()
	return
}
