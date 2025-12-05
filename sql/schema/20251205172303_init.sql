-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE assets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scan_folder_id INTEGER NULL,
    parent_asset_id INTEGER NULL,

    file_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    file_type TEXT NOT NULL DEFAULT '',
    file_size INTEGER NOT NULL DEFAULT 0,
    thumbnail_path TEXT NOT NULL DEFAULT '',

    rating INTEGER NOT NULL DEFAULT 0,
    description TEXT NULL,
    is_favorite BOOLEAN DEFAULT 0,

    image_width INTEGER NULL,
    image_height INTEGER NULL,
    dominant_color TEXT NULL,
    bit_depth INTEGER NULL,
    has_alpha_channel BOOLEAN NULL,

    date_added DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_scanned DATETIME NOT NULL,
    last_modified DATETIME NOT NULL,

    file_hash TEXT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT 0,
    deleted_at DATETIME NULL,

    -- CONSTRAINT fk_assets_scan_folder
    --     FOREIGN KEY (scan_folder_id)
    --     REFERENCES scan_folders(id)
    --     ON DELETE SET NULL,

    -- CONSTRAINT fk_assets_parent
    --     FOREIGN KEY (parent_asset_id)
    --     REFERENCES assets(id)
    --     ON DELETE SET NULL
);

-- Indeksy dla wydajności (to jest kluczowe przy tysiącach plików)
CREATE UNIQUE INDEX idx_assets_file_path ON assets(file_path);
CREATE INDEX idx_assets_scan_folder_id ON assets(scan_folder_id);
CREATE INDEX idx_assets_file_hash ON assets(file_hash);
CREATE INDEX idx_assets_file_type ON assets(file_type);

-- +goose Down
DROP INDEX IF EXISTS idx_assets_file_type;
DROP INDEX IF EXISTS idx_assets_file_hash;
DROP INDEX IF EXISTS idx_assets_scan_folder_id;
DROP INDEX IF EXISTS idx_assets_file_path;
DROP TABLE IF EXISTS assets;
