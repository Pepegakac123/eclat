-- +goose Up
ALTER TABLE assets ADD COLUMN is_hidden BOOLEAN NOT NULL DEFAULT 0;

-- Opcjonalnie: indeks, jeśli będziemy często filtrować
CREATE INDEX idx_assets_hidden ON assets(is_hidden);

-- +goose Down
DROP INDEX idx_assets_hidden;
ALTER TABLE assets DROP COLUMN is_hidden;
