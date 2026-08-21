-- +goose Up
CREATE TABLE difficulties (
  id VARCHAR(255) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_difficulties_deleted_at ON difficulties (deleted_at);

INSERT INTO difficulties (id, name) VALUES
  ('easy', 'Easy'),
  ('medium', 'Medium'),
  ('hard', 'Hard');

-- +goose Down
DROP TABLE IF EXISTS difficulties;
