SELECT
  u.id, u.email, u.first_name, u.last_name, COUNT(ug.id) AS group_count
FROM
  users u
LEFT JOIN user_groups ug ON u.id = ug.user_id