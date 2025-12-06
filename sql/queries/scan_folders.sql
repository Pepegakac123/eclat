-- name: ListScanFolders :many
SELECT * FROM scan_folders
WHERE is_deleted = 0
ORDER BY path ASC;

-- name: GetScanFolderByPath :one
SELECT * FROM scan_folders
WHERE path = ? LIMIT 1;

-- name: CreateScanFolder :one
INSERT INTO scan_folders (path, is_active, last_scanned)
VALUES (?, 1, NULL)
RETURNING *;

-- name: UpdateScanFolderStatus :exec
UPDATE scan_folders
SET is_active = ?
WHERE id = ?;

-- name: UpdateScanFolderLastScanned :exec
UPDATE scan_folders
SET last_scanned = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: SoftDeleteScanFolder :exec
UPDATE scan_folders
SET is_deleted = 1
WHERE id = ?;

-- name: RestoreScanFolder :exec
UPDATE scan_folders
SET is_deleted = 0
WHERE id = ?;
