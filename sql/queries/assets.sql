-- name: CreateAsset :one
INSERT INTO assets (
    scan_folder_id, file_name, file_path, file_type, file_size,
    thumbnail_path, file_hash,
    image_width, image_height, dominant_color, bit_depth, has_alpha_channel,
    last_modified, last_scanned
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: GetAssetById :one
SELECT * FROM assets
WHERE id = ? LIMIT 1;

-- name: GetAssetByPath :one
SELECT * FROM assets
WHERE file_path = ? LIMIT 1;

-- name: GetAssetByHash :one
SELECT * FROM assets
WHERE file_hash = ? AND file_hash IS NOT NULL
LIMIT 1;

-- name: UpdateAssetFromScan :one
UPDATE assets
SET
    -- Identyfikacja i Status (Move / Rename / Restore)
    file_path = COALESCE(sqlc.narg('file_path'), file_path),
    scan_folder_id = COALESCE(sqlc.narg('scan_folder_id'), scan_folder_id),
    is_deleted = COALESCE(sqlc.narg('is_deleted'), is_deleted),

    -- Metadane Techniczne (Refresh Content)
    file_size = COALESCE(sqlc.narg('file_size'), file_size),
    file_hash = COALESCE(sqlc.narg('file_hash'), file_hash),
    last_modified = COALESCE(sqlc.narg('last_modified'), last_modified),
    last_scanned = COALESCE(sqlc.narg('last_scanned'), last_scanned),

    -- Metadane Obrazu (Thumbnail Generator)
    thumbnail_path = COALESCE(sqlc.narg('thumbnail_path'), thumbnail_path),
    image_width = COALESCE(sqlc.narg('image_width'), image_width),
    image_height = COALESCE(sqlc.narg('image_height'), image_height),
    dominant_color = COALESCE(sqlc.narg('dominant_color'), dominant_color),
    bit_depth = COALESCE(sqlc.narg('bit_depth'), bit_depth),
    has_alpha_channel = COALESCE(sqlc.narg('has_alpha_channel'), has_alpha_channel)
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: UpdateAssetMetadata :one
UPDATE assets
SET
    description = COALESCE(sqlc.narg('description'), description),
    rating = COALESCE(sqlc.narg('rating'), rating),
    is_favorite = COALESCE(sqlc.narg('is_favorite'), is_favorite),
    thumbnail_path = COALESCE(sqlc.narg('thumbnail_path'), thumbnail_path)
WHERE id = ?
RETURNING *;

-- name: ListAssets :many
SELECT a.* FROM assets a
JOIN scan_folders f ON a.scan_folder_id = f.id
WHERE a.is_deleted = 0
  AND f.is_deleted = 0
  AND f.is_active = 1
  AND is_hidden = 0
ORDER BY a.date_added DESC
LIMIT ? OFFSET ?;

-- name: SetAssetHidden :exec
UPDATE assets SET is_hidden = ? WHERE id = ?;

-- name: SetAssetsHiddenByFolderId :exec
UPDATE assets
SET is_hidden = ?
WHERE scan_folder_id = ?;
-- name: ListFavoriteAssets :many
SELECT a.* FROM assets a
JOIN scan_folders f ON a.scan_folder_id = f.id
WHERE a.is_favorite = 1
  AND a.is_deleted = 0
  AND f.is_deleted = 0
  AND f.is_active = 1
  AND is_hidden = 0
ORDER BY a.date_added DESC
LIMIT ? OFFSET ?;

-- name: RefreshAssetTechnicalMetadata :exec
UPDATE assets
SET
    file_size = ?,
    last_modified = ?,
    last_scanned = ?,
    thumbnail_path = ?,
    image_width = ?,
    image_height = ?,
    dominant_color = ?,
    bit_depth = ?,
    has_alpha_channel = ?
WHERE id = ?;

-- name: ListDeletedAssets :many
SELECT * FROM assets
WHERE is_deleted = 1 AND is_hidden = 0
ORDER BY deleted_at DESC
LIMIT ? OFFSET ?;

-- name: ListHiddenAssets :many
SELECT * FROM assets
WHERE is_hidden = 1 AND is_deleted = 0
ORDER BY deleted_at DESC
LIMIT ? OFFSET ?;

-- name: ListUntaggedAssets :many
SELECT a.* FROM assets a
LEFT JOIN asset_tags at ON a.id = at.asset_id
JOIN scan_folders f ON a.scan_folder_id = f.id -- DODANO JOIN
WHERE at.tag_id IS NULL
  AND a.is_deleted = 0
  AND f.is_deleted = 0
  AND f.is_active = 1
  AND is_hidden = 0
GROUP BY a.id
LIMIT ? OFFSET ?;

-- name: ListAssetsForCache :many
SELECT id,file_path,last_modified,is_deleted,scan_folder_id FROM assets;

-- name: SetAssetRating :exec
UPDATE assets SET rating = ? WHERE id = ?;

-- name: ToggleAssetFavorite :exec
UPDATE assets SET is_favorite = NOT is_favorite WHERE id = ?;

-- name: UpdateAssetScanStatus :exec
UPDATE assets
SET last_scanned = ?, file_size = ?, last_modified = ?
WHERE id = ?;

-- name: UpdateAssetLocation :exec
UPDATE assets
SET file_path = ?, scan_folder_id = ?, is_deleted = false, last_scanned = ?
WHERE id = ?;

-- name: SoftDeleteAsset :exec
UPDATE assets
SET is_deleted = 1, deleted_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: RestoreAsset :exec
UPDATE assets
SET is_deleted = 0, deleted_at = NULL
WHERE id = ?;

-- name: DeleteAssetPermanent :exec
DELETE FROM assets WHERE id = ?;

-- name: DeleteAssetByFolder :exec
DELETE FROM assets WHERE scan_folder_id = ?;

-- name: GetLibraryStats :one
SELECT
    COUNT(*) as total_count,
    COALESCE(SUM(file_size), 0) as total_size,
    MAX(last_scanned) as last_scan
FROM assets a
JOIN scan_folders f ON a.scan_folder_id = f.id
WHERE a.is_deleted = 0 AND f.is_deleted = 0 AND f.is_active = 1 AND is_hidden = 0;

-- name: GetSidebarStats :one
SELECT
    (SELECT COUNT(*) FROM assets a
     JOIN scan_folders f ON a.scan_folder_id = f.id
     WHERE a.is_deleted = 0 AND f.is_deleted = 0 AND f.is_active = 1) as all_count,

    (SELECT COUNT(*) FROM assets a
     JOIN scan_folders f ON a.scan_folder_id = f.id
     WHERE a.is_favorite = 1 AND a.is_deleted = 0 AND f.is_deleted = 0 AND f.is_active = 1) as favorites_count,

    (SELECT COUNT(*) FROM assets WHERE is_deleted = 1 AND is_hidden = 0) as trash_count,
    (SELECT COUNT(*) FROM assets WHERE is_hidden = 1 AND is_deleted = 0) as hidden_count,

    (SELECT COUNT(DISTINCT a.id)
     FROM assets a
     LEFT JOIN asset_tags at ON a.id = at.asset_id
     JOIN scan_folders f ON a.scan_folder_id = f.id
     WHERE at.tag_id IS NULL AND a.is_deleted = 0 AND f.is_deleted = 0 AND f.is_active = 1) as uncategorized_count;

-- name: GetAllColors :many
SELECT DISTINCT dominant_color
FROM assets a
JOIN scan_folders f ON a.scan_folder_id = f.id
WHERE a.is_deleted = 0
  AND f.is_deleted = 0
  AND f.is_active = 1
  AND is_hidden = 0
  AND dominant_color IS NOT NULL AND dominant_color != '';

-- name: MoveAssetsToFolder :exec
UPDATE assets SET scan_folder_id = ? WHERE scan_folder_id = ?;

-- name: ClaimAssetsForPath :exec
UPDATE assets
SET scan_folder_id = ?
WHERE file_path LIKE ? || '%'
AND scan_folder_id != ?;
