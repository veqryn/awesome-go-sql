-- name: SelectAccountByID :one
SELECT *
FROM accounts
WHERE id = $1;

-- name: SelectAllAccounts :many
SELECT *
FROM accounts
ORDER BY id;

-- name: SelectAllAccountsByFilter :many
SELECT *
FROM accounts
WHERE (CASE WHEN @any_names::bool THEN name = ANY(@names::text[]) ELSE TRUE END)
  AND (CASE WHEN @is_active::bool THEN active = @active ELSE TRUE END)
  AND (CASE WHEN @any_fav_color::bool THEN fav_color = ANY(@fav_colors::COLORS[]) ELSE TRUE END)
;
