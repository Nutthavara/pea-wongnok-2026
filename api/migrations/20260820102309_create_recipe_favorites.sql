-- +goose Up
CREATE TABLE user_favorites (
  user_id UUID NOT NULL REFERENCES users (id),
  recipe_id INTEGER NOT NULL REFERENCES recipes (id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ,
  PRIMARY KEY (user_id, recipe_id)
);

CREATE INDEX idx_user_favorites_deleted_at ON user_favorites (deleted_at);

-- +goose Down
DROP TABLE IF EXISTS user_favorites;
