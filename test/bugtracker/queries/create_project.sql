INSERT INTO projects
  (tid, "key", "name", created)
VALUES
  ($tid, $key, $name, CURRENT_TIMESTAMP);