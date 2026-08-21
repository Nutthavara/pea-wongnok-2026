-- +goose Up
CREATE TABLE recipes (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  image_url TEXT,
  difficulty_id VARCHAR(255) REFERENCES difficulties (id),
  duration_id VARCHAR(255) REFERENCES durations (id),
  average_rating DOUBLE PRECISION NOT NULL DEFAULT 0,
  creator_id UUID REFERENCES users (id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_recipes_difficulty_id ON recipes (difficulty_id);
CREATE INDEX idx_recipes_duration_id ON recipes (duration_id);
CREATE INDEX idx_recipes_creator_id ON recipes (creator_id);
CREATE INDEX idx_recipes_deleted_at ON recipes (deleted_at);

-- +goose Down
DROP TABLE IF EXISTS recipes;
