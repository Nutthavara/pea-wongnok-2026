-- +goose Up
-- +goose StatementBegin
CREATE TABLE
  recipe_ratings (
    user_id UUID NOT NULL REFERENCES users (id),
    recipe_id INTEGER NOT NULL REFERENCES recipes (id),
    score FLOAT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    deleted_at TIMESTAMPTZ,
    PRIMARY KEY (user_id, recipe_id)
  );

-- composite primary key ด้านบนบังคับ unique (user_id, recipe_id) ให้แล้ว
CREATE INDEX idx_recipe_ratings_deleted_at ON recipe_ratings (deleted_at);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS recipe_ratings;

-- +goose StatementEnd