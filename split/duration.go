package split

import (
	"database/sql"
	"time"

	"github.com/knoebber/gsplits/db"
)

// Duration is the amount of time that a split took.
type Duration struct {
	ID       int64
	RunID    int64         `validate:"required"`
	NameID   int64         `validate:"required"`
	Duration time.Duration `validate:"required"`
}

func (Duration) String() string {
	return "split duration"
}

// Save inserts the duration into the splits table.
func (d *Duration) Save(tx *sql.Tx) (sql.Result, error) {
	if err := db.Validate(d); err != nil {
		return nil, err
	}
	ms := d.Duration.Nanoseconds() / 1e6
	return tx.Exec(
		"INSERT INTO split(run_id, split_name_id, milliseconds) VALUES (?, ?, ?)",
		d.RunID,
		d.NameID,
		ms,
	)
}
