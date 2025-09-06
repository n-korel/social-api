CREATE TABLE IF NOT EXISTS roles (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE,
  level int NOT NULL DEFAULT 0,
  description TEXT
);

INSERT INTO
  roles (name, description, level)
VALUES
  (
    'user',
    'Can create posts and comments',
    1
  );

INSERT INTO
  roles (name, description, level)
VALUES
  (
    'moderator',
    'Can update other users posts',
    2
  );

INSERT INTO
  roles (name, description, level)
VALUES
  (
    'admin',
    'Can update and delete other users posts',
    3
  );