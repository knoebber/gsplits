-- --- Use for debugging the getRoute function.
-- --- $ sqlite3 ~/.gsplits.db
-- --- sqlite> .read test_route_query.sql
.mode column
.headers on
SELECT sn.id AS split_name_id,
       sn.name AS split_name, 
       MIN(golds.milliseconds) AS gold_split,
       route_best.milliseconds AS route_best_split,
       r.id AS route_id,
       r.name AS route_name,
       COUNT(DISTINCT run.id) AS total_runs,
       c.id AS category_id,
       c.name AS category_name,
       MIN(run.milliseconds) AS pb
FROM route AS r
JOIN split_name AS sn ON sn.route_id = r.id
JOIN split AS golds ON golds.split_name_id = sn.id 
JOIN run ON run.route_id = r.id
JOIN category AS c ON c.id = r.category_id
JOIN ( SELECT r.id AS route_id,
              sn.id AS split_name_id,
              s.milliseconds
       FROM category AS c
       JOIN route AS r ON r.category_id = c.id
       JOIN run ON run.route_id = r.id
       JOIN split_name AS sn ON sn.route_id = r.id
       JOIN split AS s ON s.split_name_id = sn.id AND s.run_id = run.id
       GROUP BY r.id, sn.id
       HAVING MIN(run.milliseconds) = run.milliseconds
       ORDER BY route_id, sn.position
     ) AS route_best ON route_best.split_name_id = sn.id
-- WHERE r.name = "" -- Insert the route name to debug here.
GROUP BY sn.id
ORDER BY r.id, sn.position

-- Sub query with extra columns
-- SELECT r.ID route_id,
--        run.id AS run_id,
--        run.milliseconds AS run_ms,
--        sn.name AS split_name,
--        s.id AS split_id,
--        s.milliseconds AS split_ms
-- FROM category AS c
-- JOIN route AS r ON r.category_id = c.id
-- JOIN run ON run.route_id = r.id
-- JOIN split_name AS sn ON sn.route_id = r.id
-- JOIN split AS s ON s.split_name_id = sn.id AND s.run_id = run.id
-- GROUP BY r.id, sn.id
-- HAVING MIN(run.milliseconds) = run.milliseconds
-- ORDER BY route_id, sn.position;
