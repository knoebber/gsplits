package route

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/knoebber/gsplits/category"
	"github.com/knoebber/gsplits/db"
	"github.com/knoebber/gsplits/split"
)

// TODO rename splits => segments
// TODO rename RouteBests ComparisonSegments

// Data contains information about an route.
// Array values should be pulled by their get methods.
// Get methods return a zero value if the index does not exist.
type Data struct {
	RouteName          string          // The name of the route.
	RouteID            int64           // The routes ID.
	Category           *category.Name  // The routes category.
	RouteBestTime      *time.Duration  // The fastest time this route has been completed.
	TotalRuns          int64           // The total amount of runs in this route.
	SumOfGold          *time.Duration  // The sum of the gold splits.
	SplitNames         []split.Name    // The names of the splits in the category.
	ComparisonSplits   []time.Duration // The total time that the comparison run had at each split.
	ComparisonSegments []time.Duration // The segments from the comparison run.
	Golds              []time.Duration // The fastest a split has ever been completed in the route.
	TimeSaves          []time.Duration // The difference of a gold and the route best
	Length             int             // The number of splits in the route.
}

// GetBPT gets the "best possible time".
// It is the current elapsed time added to the rest of the golds left.
func (d *Data) GetBPT(lastTotal time.Duration, splitIndex int) (bpt time.Duration) {
	bpt = lastTotal
	if splitIndex >= len(d.Golds) {
		return
	}

	remaining := d.Golds[splitIndex:]
	for _, gold := range remaining {
		bpt += gold
	}
	return
}

// GetSplitName returns the split name at index.
func (d *Data) GetSplitName(index int) string {
	if index >= len(d.SplitNames) {
		return ""
	}
	return d.SplitNames[index].Name
}

// GetComparisonSplit returns the total time the comparison run had at index.
func (d *Data) GetComparisonSplit(index int) time.Duration {
	if index >= len(d.ComparisonSplits) {
		return 0
	}
	return d.ComparisonSplits[index]
}

// GetComparisonSegment returns the duration of the comparison segment at index.
func (d *Data) GetComparisonSegment(index int) time.Duration {
	if index >= len(d.ComparisonSegments) {
		return 0
	}
	return d.ComparisonSegments[index]
}

// GetGold returns the fastest segment at index for the route.
func (d *Data) GetGold(index int) time.Duration {
	if index >= len(d.Golds) {
		return 0
	}
	return d.Golds[index]
}

// GetTimeSave returns the difference between the comparison segment and the gold segment at index.
func (d *Data) GetTimeSave(index int) time.Duration {
	if index >= len(d.TimeSaves) {
		return 0
	}
	return d.TimeSaves[index]
}

// GetData gets a routes data by its primary key.
// Returns nil if the route isn't found.
func GetData(routeID int64) (*Data, error) {
	var (
		routeBestTime    *int64
		categoryBestTime *int64
		currBest         *int64
		currGold         *int64
		sumOfGold        *int64
		zero             int64
		totalRuns        int64
	)

	if routeID == 0 {
		return nil, errors.New("id is required")
	}

	rows, err := dataQuery(routeID)
	if err != nil {
		return nil, fmt.Errorf("failed route data query: %w", err)
	}
	defer rows.Close()

	d := new(Data)
	d.Category = &category.Name{}
	d.SplitNames = []split.Name{}
	d.ComparisonSplits = []time.Duration{}
	d.ComparisonSegments = []time.Duration{}
	d.Golds = []time.Duration{}
	d.TimeSaves = []time.Duration{}

	for rows.Next() {
		sn := split.Name{}

		if err := rows.Scan(
			&sn.ID,
			&sn.Name,
			&currGold,
			&currBest,
			&d.RouteID,
			&d.RouteName,
			&routeBestTime,
			&totalRuns,
			&d.Category.ID,
			&d.Category.Name,
			&categoryBestTime,
		); err != nil {
			return nil, err
		}

		d.SplitNames = append(d.SplitNames, sn)

		// When the route has a completed run these should be non nil.
		if currBest != nil && currGold != nil {
			if sumOfGold == nil {
				sumOfGold = &zero
			}
			*sumOfGold += *currGold

			curBestDur := time.Duration(*currBest * 1e6)
			if len(d.ComparisonSplits) == 0 {
				d.ComparisonSplits = append(d.ComparisonSplits, curBestDur)
			} else {
				d.ComparisonSplits = append(d.ComparisonSplits, d.ComparisonSplits[len(d.ComparisonSplits)-1]+curBestDur)
			}
			d.ComparisonSegments = append(d.ComparisonSegments, time.Duration(*currBest*1e6))
			d.Golds = append(d.Golds, time.Duration(*currGold*1e6))
			d.TimeSaves = append(d.TimeSaves, time.Duration((*currBest-*currGold)*1e6))
		}

	}

	if sumOfGold != nil {
		dur := time.Duration(*sumOfGold * 1e6)
		d.SumOfGold = &dur
	}
	if routeBestTime != nil {
		dur := time.Duration(*routeBestTime * 1e6)
		d.RouteBestTime = &dur
	}

	if categoryBestTime != nil {
		dur := time.Duration(*categoryBestTime * 1e6)
		d.Category.Best = &dur
	}

	if len(d.SplitNames) == 0 {
		return nil, errors.New("route not found")
	}

	d.Length = len(d.SplitNames)
	return d, nil
}

func dataQuery(routeID int64) (*sql.Rows, error) {
	query := fmt.Sprintf(`
SELECT
  sn.id AS split_name_id,
  sn.name AS split_name,
  MIN(golds.milliseconds) AS gold_split,
  route_best.milliseconds AS route_best_split,
  r.id AS route_id,
  r.name AS route_name,
  MIN(run.milliseconds) AS route_best,
  COUNT(DISTINCT run.id) AS total_runs,
  c.id AS category_id,
  c.name AS category_name,
  category_best.milliseconds AS category_best
FROM
  route AS r
  JOIN category AS c ON c.id = r.category_id
  JOIN split_name AS sn ON sn.route_id = r.id
  LEFT JOIN split AS golds ON golds.split_name_id = sn.id
  LEFT JOIN run ON run.route_id = r.id
  LEFT JOIN (
    SELECT
      r.id AS route_id,
      sn.id AS split_name_id,
      s.milliseconds
    FROM
      category AS c
      JOIN route AS r ON r.category_id = c.id
      JOIN run ON run.route_id = r.id
      JOIN split_name AS sn ON sn.route_id = r.id
      JOIN split AS s ON s.split_name_id = sn.id
      AND s.run_id = run.id
    GROUP BY
      r.id,
      sn.id
    HAVING
      MIN(run.milliseconds) = run.milliseconds
  ) AS route_best ON route_best.split_name_id = sn.id
  LEFT JOIN (
    SELECT
      MIN(run.milliseconds) AS milliseconds,
      c.id
    FROM
      category AS c
      JOIN route AS r ON r.category_id = c.id
      JOIN run ON run.route_id = r.id
    GROUP BY
      c.id
  ) AS category_best ON category_best.id = c.id
  WHERE
    r.id = ?
GROUP BY
  sn.id
ORDER BY r.id, sn.position;
`)
	return db.Connection.Query(query, routeID)
}
