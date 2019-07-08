INSERT INTO issues
  (tid, id, "name", project_key, fields, created)
VALUES
  ($tid, $id, $name, $project_key, $fields, CURRENT_TIMESTAMP);