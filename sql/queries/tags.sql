-- name: ListTags :many
SELECT t.id, t.name, COUNT(at.asset_id) as asset_count
FROM tags t
LEFT JOIN asset_tags at ON t.id = at.tag_id
GROUP BY t.id
ORDER BY asset_count DESC;

-- name: GetTagByName :one
SELECT * FROM tags WHERE name = ? LIMIT 1;

-- name: CreateTag :one
INSERT INTO tags (name) VALUES (?) RETURNING *;

-- name: AddTagToAsset :exec
INSERT OR IGNORE INTO asset_tags (asset_id, tag_id) VALUES (?, ?);

-- name: RemoveTagFromAsset :exec
DELETE FROM asset_tags WHERE asset_id = ? AND tag_id = ?;

-- name: ClearTagsFromAsset :exec
DELETE FROM asset_tags WHERE asset_id = ?;

-- name: GetTagsForAsset :many
SELECT t.* FROM tags t
JOIN asset_tags at ON t.id = at.tag_id
WHERE at.asset_id = ?
ORDER BY t.name;
