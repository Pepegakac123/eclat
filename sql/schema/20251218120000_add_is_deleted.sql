-- +goose Up
ALTER TABLE scan_folders ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE scan_folders DROP COLUMN is_deleted;
