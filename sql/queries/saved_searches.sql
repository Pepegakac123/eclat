-- name: ListSavedSearches :many
SELECT * FROM saved_searches ORDER BY name;

-- name: CreateSavedSearch :one
INSERT INTO saved_searches (name, filter_json) VALUES (?, ?) RETURNING *;

-- name: DeleteSavedSearch :exec
DELETE FROM saved_searches WHERE id = ?;
