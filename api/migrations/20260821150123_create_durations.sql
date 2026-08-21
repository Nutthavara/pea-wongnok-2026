-- +goose Up
CREATE TABLE durations (
  id VARCHAR(255) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_durations_deleted_at ON durations (deleted_at);

INSERT INTO durations (id, name) VALUES
  ('10m', '5 - 10 mins'),
  ('30m', '10 - 30 mins'),
  ('60m', '~1 Hour'),
  ('long', 'More than 1 hour');

-- +goose Down
DROP TABLE IF EXISTS durations;
