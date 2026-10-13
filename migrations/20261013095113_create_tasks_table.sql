-- +goose Up
CREATE TABLE tasks (
    id BIGSERIAL NOT NULL PRIMARY KEY,
    title TEXT NOT NULL,
    done BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_tasks_created_at ON tasks (created_at DESC, id DESC);
-- +goose Down
DROP TABLE tasks;

