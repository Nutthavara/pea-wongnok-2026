-- +goose Up
-- +goose StatementBegin
CREATE TABLE
  recipes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    image_url VARCHAR(255),
    difficulty_id VARCHAR(255) REFERENCES difficulties (id),
    duration_id VARCHAR(255) REFERENCES durations (id),
    average_rating FLOAT NOT NULL DEFAULT 0,
    creator_id UUID REFERENCES users (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    deleted_at TIMESTAMPTZ
  );

-- Postgres ไม่สร้าง index ให้ foreign key อัตโนมัติ เพิ่มเองเพื่อ join/filter เร็วขึ้น
CREATE INDEX idx_recipes_difficulty_id ON recipes (difficulty_id);

CREATE INDEX idx_recipes_duration_id ON recipes (duration_id);

CREATE INDEX idx_recipes_creator_id ON recipes (creator_id);

CREATE INDEX idx_recipes_deleted_at ON recipes (deleted_at);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS recipes;

-- +goose StatementEnd