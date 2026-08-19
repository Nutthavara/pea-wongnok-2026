-- +goose Up
-- +goose StatementBegin
CREATE TABLE
  difficulties (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
    deleted_at TIMESTAMPTZ
  );

CREATE INDEX idx_difficulties_deleted_at ON difficulties (deleted_at);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS difficulties;

-- +goose StatementEnd