-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE scan_folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    last_scanned DATETIME,
    date_added DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE assets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scan_folder_id INTEGER,
    parent_asset_id INTEGER,

    file_name TEXT NOT NULL,
    file_path TEXT NOT NULL UNIQUE,
    file_type TEXT NOT NULL DEFAULT '',
    file_size INTEGER NOT NULL DEFAULT 0,
    thumbnail_path TEXT NOT NULL DEFAULT '',

    rating INTEGER NOT NULL DEFAULT 0,
    description TEXT,
    is_favorite BOOLEAN DEFAULT 0,

    image_width INTEGER,
    image_height INTEGER,
    dominant_color TEXT,
    bit_depth INTEGER,
    has_alpha_channel BOOLEAN,

    date_added DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_scanned DATETIME NOT NULL,
    last_modified DATETIME NOT NULL,

    file_hash TEXT,
    is_deleted BOOLEAN NOT NULL DEFAULT 0,
    deleted_at DATETIME,

    CONSTRAINT fk_assets_scan_folder
        FOREIGN KEY (scan_folder_id)
        REFERENCES scan_folders(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_assets_parent
        FOREIGN KEY (parent_asset_id)
        REFERENCES assets(id)
        ON DELETE SET NULL
);

-- Indeksy
CREATE INDEX idx_assets_scan_folder_id ON assets(scan_folder_id);
CREATE INDEX idx_assets_file_hash ON assets(file_hash);
CREATE INDEX idx_assets_file_type ON assets(file_type);

-- +goose Down
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS scan_folders;
