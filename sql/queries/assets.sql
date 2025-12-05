-- name: CreateAsset :one
INSERT INTO assets (
    scan_folder_id, file_name, file_path, file_type, file_size,
    thumbnail_path, last_modified, last_scanned, file_hash,
    image_width, image_height, dominant_color, bit_depth, has_alpha_channel
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: GetAsset :one
SELECT * FROM assets
WHERE id = ? LIMIT 1;

-- name: ListAssets :many
SELECT * FROM assets
WHERE is_deleted = 0
ORDER BY date_added DESC
LIMIT ? OFFSET ?;

-- name: CountAssets :one
SELECT count(*) FROM assets
WHERE is_deleted = 0;

-- name: UpdateAssetScanStatus :exec
UPDATE assets
SET last_scanned = ?, file_size = ?, last_modified = ?
WHERE id = ?;
