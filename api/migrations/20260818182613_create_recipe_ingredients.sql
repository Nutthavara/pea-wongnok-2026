-- +goose Up
-- +goose StatementBegin
CREATE TABLE
  recipe_ingredients (
    id SERIAL PRIMARY KEY,
    recipe_id INTEGER REFERENCES recipes (id),
    description TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    deleted_at TIMESTAMPTZ
  );

CREATE INDEX idx_recipe_ingredients_recipe_id ON recipe_ingredients (recipe_id);

CREATE INDEX idx_recipe_ingredients_deleted_at ON recipe_ingredients (deleted_at);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS recipe_ingredients;

-- +goose StatementEnd