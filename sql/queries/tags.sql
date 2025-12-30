-- name: ListTags :many
SELECT t.id, t.name, COUNT(at.asset_id) as asset_count
FROM tags t
LEFT JOIN asset_tags at ON t.id = at.tag_id
GROUP BY t.id
ORDER BY asset_count DESC;

-- name: CreateTag :one
INSERT INTO tags (name) VALUES (?)
ON CONFLICT(name) DO UPDATE SET name=name
RETURNING *;

-- name: GetTagByName :one
SELECT * FROM tags WHERE name = ? LIMIT 1;

-- name: GetAllTags :many
SELECT * FROM tags ORDER BY name ASC;

-- name: AddTagToAsset :exec
INSERT INTO asset_tags (asset_id, tag_id)
VALUES (?, ?)
ON CONFLICT DO NOTHING;

-- name: RemoveTagFromAsset :exec
DELETE FROM asset_tags
WHERE asset_id = ? AND tag_id = ?;

-- name: ClearTagsForAsset :exec
DELETE FROM asset_tags
WHERE asset_id = ?;

-- name: GetTagsByAssetID :many
SELECT t.*
FROM tags t
JOIN asset_tags at ON t.id = at.tag_id
WHERE at.asset_id = ?
ORDER BY t.name ASC;

-- name: GetTagsNamesByAssetID :many
SELECT t.name
FROM tags t
JOIN asset_tags at ON t.id = at.tag_id
WHERE at.asset_id = ?
ORDER BY t.name ASC;
