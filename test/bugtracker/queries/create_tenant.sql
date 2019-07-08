INSERT INTO tenants 
  (id, "key", "name", created)
VALUES
  ($id, $key, $name, CURRENT_TIMESTAMP);