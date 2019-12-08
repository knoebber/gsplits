package route

import (
	"database/sql"
	"time"

	"github.com/knoebber/gsplits/db"
)

// Run is a single run in a route.
type Run struct {
	ID        int64
	RouteID   int64         `validate:"required"`
	Duration  time.Duration `validate:"required"`
	CreatedAt time.Time     `validate:"required"`
}

func (r Run) String() string {
	return "run"
}

// Save inserts the run into the runs table.
func (r *Run) Save(tx *sql.Tx) (sql.Result, error) {
	r.CreatedAt = time.Now()
	if err := db.Validate(r); err != nil {
		return nil, err
	}
	ms := r.Duration.Nanoseconds() / 1e6
	return tx.Exec("INSERT INTO run(route_id, milliseconds) VALUES(?, ?)", r.RouteID, ms)
}
