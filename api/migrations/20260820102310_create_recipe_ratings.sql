-- +goose Up
CREATE TABLE recipe_ratings (
  user_id UUID NOT NULL REFERENCES users (id),
  recipe_id INTEGER NOT NULL REFERENCES recipes (id),
  score DOUBLE PRECISION NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ,
  PRIMARY KEY (user_id, recipe_id)
);

CREATE INDEX idx_recipe_ratings_deleted_at ON recipe_ratings (deleted_at);

-- +goose Down
DROP TABLE IF EXISTS recipe_ratings;
