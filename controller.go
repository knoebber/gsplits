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

func saveCategory(name string) (categoryID int64, err error) {
	var (
		tx  *sql.Tx
		res sql.Result
	)

	tx, err = db.Connection.Begin()
	if err != nil {
		fmt.Errorf("failed to start save category transaction: %w", err)
	}

	categoryName := &category.Name{
		Name: name,
	}

	res, err = categoryName.Save(tx)
	if err != nil {
		return 0, saveError(categoryName, tx, err)
	}

	categoryID, err = res.LastInsertId()
	if err != nil {
		return 0, idError(categoryName, tx, err)
	}

	err = tx.Commit()
	return
}

func saveRoute(categoryID int64, name string, splitNames []string) (routeID int64, err error) {
	var (
		tx  *sql.Tx
		res sql.Result
	)

	tx, err = db.Connection.Begin()
	if err != nil {
		fmt.Errorf("failed to start save route transaction: %w", err)
	}

	routeName := &route.Name{
		Name:       name,
		CategoryID: categoryID,
	}
	res, err = routeName.Save(tx)
	if err != nil {
		return 0, saveError(routeName, tx, err)
	}

	routeID, err = res.LastInsertId()
	if err != nil {
		return 0, idError(routeName, tx, err)
	}

	sn := &split.Name{
		RouteID: routeID,
	}
	for i, splitName := range splitNames {
		sn.Position = i + 1
		sn.Name = splitName
		_, err := sn.Save(tx)
		if err != nil {
			return 0, saveError(sn, tx, err)
		}
	}

	err = tx.Commit()
	return
}

func saveRun(routeID int64, durations []time.Duration, totalDuration time.Duration) (runID int64, err error) {
	var (
		tx         *sql.Tx
		res        sql.Result
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
	res, err = run.Save(tx)
	if err != nil {
		return 0, saveError(run, tx, err)
	}

	runID, err = res.LastInsertId()
	if err != nil {
		return 0, idError(run, tx, err)
	}

	for i, duration := range durations {
		d := &split.Duration{
			RunID:    runID,
			Duration: duration,
			NameID:   splitNames[i].ID,
		}
		if _, err := d.Save(tx); err != nil {
			return 0, saveError(d, tx, err)
		}
	}

	err = tx.Commit()
	return
}

func idError(s saver, tx *sql.Tx, err error) error {
	return fmt.Errorf("failed to get ID from %s: %w", s, db.Rollback(tx, err))
}

func saveError(s saver, tx *sql.Tx, err error) error {
	return fmt.Errorf("failed to save %s: %w", s, db.Rollback(tx, err))
}
