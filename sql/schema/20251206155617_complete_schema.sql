-- +goose Up
-- Włączamy klucze obce dla pewności
PRAGMA foreign_keys = ON;

-- 1. Tagi
CREATE TABLE tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    date_created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. Kolekcje (Material Sets)
CREATE TABLE material_sets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    cover_asset_id INTEGER,
    custom_cover_url TEXT,
    custom_color TEXT,
    date_added DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_modified DATETIME NOT NULL,
    FOREIGN KEY (cover_asset_id) REFERENCES assets(id) ON DELETE SET NULL
);

-- 3. Zapisane Wyszukiwania (Saved Searches)
CREATE TABLE saved_searches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    filter_json TEXT NOT NULL, -- JSON z filtrami
    date_added DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 4. Ustawienia Systemowe (Klucz-Wartość)
CREATE TABLE system_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- ==========================================
-- TABELE ŁĄCZĄCE (MANY-TO-MANY)
-- ==========================================

-- 5. Asset <-> Tag
CREATE TABLE asset_tags (
    asset_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (asset_id, tag_id),
    FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- 6. Asset <-> Material Set
CREATE TABLE asset_material_sets (
    asset_id INTEGER NOT NULL,
    material_set_id INTEGER NOT NULL,
    PRIMARY KEY (asset_id, material_set_id),
    FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE CASCADE,
    FOREIGN KEY (material_set_id) REFERENCES material_sets(id) ON DELETE CASCADE
);

-- Indeksy dla wydajności (zgodnie z AssetDbContext.cs)
CREATE INDEX idx_tags_name ON tags(name);
CREATE INDEX idx_material_sets_name ON material_sets(name);
CREATE INDEX idx_saved_searches_name ON saved_searches(name);

-- +goose Down
DROP TABLE IF EXISTS asset_material_sets;
DROP TABLE IF EXISTS asset_tags;
DROP TABLE IF EXISTS system_settings;
DROP TABLE IF EXISTS saved_searches;
DROP TABLE IF EXISTS material_sets;
DROP TABLE IF EXISTS tags;
