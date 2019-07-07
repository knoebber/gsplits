-- --- Use for debugging the getRoute function.
-- --- $ sqlite3 ~/.gsplits.db
-- --- sqlite> .read test_route_query.sql
.mode column
.headers on
SELECT
  sn.id AS split_name_id,
  sn.name AS split_name,
  MIN(golds.milliseconds) AS gold_split,
  route_best.milliseconds AS route_best_split,
  MIN(run.milliseconds) AS route_best,
  r.id AS route_id,
  r.name AS route_name,
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
GROUP BY
  sn.id
ORDER BY r.id, sn.position;
