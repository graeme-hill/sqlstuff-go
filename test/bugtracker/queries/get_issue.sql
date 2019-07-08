SELECT
  id, "name", fields, created, modified, p."name" AS project_name
FROM issues i
  JOIN projects p ON p.tid = i.tid AND p."key" = i.project_key
WHERE i.tid = $tid AND i.id = $id
LIMIT 1;

SELECT
  tag_key, created
FROM issue_tags
WHERE tid = $tid AND issue_id = $id;
