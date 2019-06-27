CREATE TABLE users (
  id INT NOT NULL PRIMARY KEY,
  email VARCHAR(200) NOT NULL
);

CREATE TABLE groups (
  id INT NOT NULL PRIMARY KEY,
  "name" VARCHAR(200) NOT NULL
);

CREATE TABLE user_groups (
  user_id INT NOT NULL,
  group_id INT NOT NULL
);