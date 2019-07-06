SELECT
  u.id, u.email, u.first_name, u.last_name, g.name AS group_name
FROM
  users u
LEFT JOIN user_groups ug ON u.id = ug.user_id
LEFT JOIN groups g ON g.id = ug.group_id;

SELECT name FROM groups;