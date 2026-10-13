-- +goose Up
ALTER TABLE tasks ADD COLUMN due_date TIMESTAMPTZ;

-- +goose Down
ALTER TABLE tasks DROP COLUMN due_date;
