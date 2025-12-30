-- name: ListMaterialSets :many
SELECT
    ms.*,
    a.thumbnail_path as cover_thumbnail_path,
    (SELECT COUNT(*) FROM asset_material_sets ams WHERE ams.material_set_id = ms.id) as total_assets
FROM material_sets ms
LEFT JOIN assets a ON ms.cover_asset_id = a.id
ORDER BY ms.name;

-- name: GetMaterialSetById :one
SELECT 
    ms.*,
    a.thumbnail_path as cover_thumbnail_path,
    (SELECT COUNT(*) FROM asset_material_sets ams WHERE ams.material_set_id = ms.id) as total_assets
FROM material_sets ms 
LEFT JOIN assets a ON ms.cover_asset_id = a.id
WHERE ms.id = ? LIMIT 1;

-- name: CreateMaterialSet :one
INSERT INTO material_sets (
    name, description, cover_asset_id, custom_cover_url, custom_color, last_modified
) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
RETURNING *;

-- name: UpdateMaterialSet :exec
UPDATE material_sets
SET
    name = ?, description = ?, cover_asset_id = ?,
    custom_cover_url = ?, custom_color = ?,
    last_modified = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteMaterialSet :exec
DELETE FROM material_sets WHERE id = ?;

-- name: AddAssetToMaterialSet :exec
INSERT OR IGNORE INTO asset_material_sets (material_set_id, asset_id) VALUES (?, ?);

-- name: RemoveAssetFromMaterialSet :exec
DELETE FROM asset_material_sets WHERE material_set_id = ? AND asset_id = ?;

-- name: ListAssetsInMaterialSet :many
SELECT a.* FROM assets a
JOIN asset_material_sets ams ON a.id = ams.asset_id
WHERE ams.material_set_id = ? AND a.is_deleted = 0
ORDER BY a.date_added DESC
LIMIT ? OFFSET ?;
