-- name: GetSystemSetting :one
SELECT value FROM system_settings WHERE key = ? LIMIT 1;

-- name: SetSystemSetting :exec
INSERT INTO system_settings (key, value)
VALUES (?, ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value;
