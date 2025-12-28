-- +goose Up
-- Wyłączamy sprawdzanie kluczy obcych na czas operacji, żeby nas nie blokowały
PRAGMA foreign_keys = OFF;

-- 1. Tworzymy nową tabelę assets_new (bez parent_asset_id, z group_id)
--    Musimy powtórzyć całą definicję tabeli, uwzględniając zmiany z poprzednich migracji (np. is_hidden).
CREATE TABLE assets_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scan_folder_id INTEGER,
    group_id TEXT NOT NULL, -- Nowa kolumna (NOT NULL, bo od razu wypełnimy)

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

    -- Kolumna z migracji 20251219 (dodaję ją tu ręcznie, bo tworzymy tabelę od zera)
    is_hidden BOOLEAN NOT NULL DEFAULT 0,

    CONSTRAINT fk_assets_scan_folder
        FOREIGN KEY (scan_folder_id)
        REFERENCES scan_folders(id)
        ON DELETE SET NULL
);

-- 2. Kopiujemy dane ze starej tabeli do nowej, generując UUID dla group_id w locie
INSERT INTO assets_new (
    id, scan_folder_id, group_id,
    file_name, file_path, file_type, file_size, thumbnail_path,
    rating, description, is_favorite,
    image_width, image_height, dominant_color, bit_depth, has_alpha_channel,
    date_added, last_scanned, last_modified,
    file_hash, is_deleted, deleted_at, is_hidden
)
SELECT
    id, scan_folder_id,
    -- Generowanie UUID v4 (Magia SQLite)
    lower(hex(randomblob(4)) || '-' || hex(randomblob(2)) || '-' || '4' || substr(hex(randomblob(2)), 2) || '-' || substr('89ab', abs(random()) % 4 + 1, 1) || substr(hex(randomblob(2)), 2) || '-' || hex(randomblob(6))),
    file_name, file_path, file_type, file_size, thumbnail_path,
    rating, description, is_favorite,
    image_width, image_height, dominant_color, bit_depth, has_alpha_channel,
    date_added, last_scanned, last_modified,
    file_hash, is_deleted, deleted_at, is_hidden
FROM assets;

-- 3. Podmieniamy tabele
DROP TABLE assets;
ALTER TABLE assets_new RENAME TO assets;

-- 4. Odtwarzamy indeksy (bo DROP TABLE je usunął)
CREATE INDEX idx_assets_scan_folder_id ON assets(scan_folder_id);
CREATE INDEX idx_assets_file_hash ON assets(file_hash);
CREATE INDEX idx_assets_hidden ON assets(is_hidden);
-- Nowy indeks dla grup
CREATE INDEX idx_assets_group_id ON assets(group_id);

-- Włączamy z powrotem sprawdzanie kluczy
PRAGMA foreign_keys = ON;

-- +goose Down
-- W dół robimy to samo, tylko w drugą stronę (tracimy dane o grupach, parent_id będzie NULL)
PRAGMA foreign_keys = OFF;

CREATE TABLE assets_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scan_folder_id INTEGER,
    parent_asset_id INTEGER, -- Przywracamy starą kolumnę

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
    is_hidden BOOLEAN NOT NULL DEFAULT 0,

    CONSTRAINT fk_assets_scan_folder FOREIGN KEY (scan_folder_id) REFERENCES scan_folders(id) ON DELETE SET NULL,
    CONSTRAINT fk_assets_parent FOREIGN KEY (parent_asset_id) REFERENCES assets(id) ON DELETE SET NULL
);

INSERT INTO assets_old (
    id, scan_folder_id, parent_asset_id,
    file_name, file_path, file_type, file_size, thumbnail_path,
    rating, description, is_favorite,
    image_width, image_height, dominant_color, bit_depth, has_alpha_channel,
    date_added, last_scanned, last_modified,
    file_hash, is_deleted, deleted_at, is_hidden
)
SELECT
    id, scan_folder_id, NULL, -- Tracimy relację rodzic-dziecko
    file_name, file_path, file_type, file_size, thumbnail_path,
    rating, description, is_favorite,
    image_width, image_height, dominant_color, bit_depth, has_alpha_channel,
    date_added, last_scanned, last_modified,
    file_hash, is_deleted, deleted_at, is_hidden
FROM assets;

DROP TABLE assets;
ALTER TABLE assets_old RENAME TO assets;

CREATE INDEX idx_assets_scan_folder_id ON assets(scan_folder_id);
CREATE INDEX idx_assets_file_hash ON assets(file_hash);
CREATE INDEX idx_assets_hidden ON assets(is_hidden);
-- idx_assets_parent_id nie był jawnie tworzony w init.sql, ale SQLite tworzy go czasem dla FK

PRAGMA foreign_keys = ON;
