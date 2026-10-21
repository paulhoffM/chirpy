-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, hashed_password, email)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;
-- name: GetUser :one
SELECT * FROM users WHERE email =$1;
-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;
-- name: GetChirps :many
SELECT * FROM chirps ORDER BY created_at ASC;
-- name: GetChirp :one
SELECT * FROM chirps WHERE id = $1;

-- name: CreateRefreshToken :one
INSERT INTO refreshToken (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;
-- name: RevokeRefreshToken :exec
Update refreshToken
SET  revoked_at = $2, updated_at = $3
WHERE token = $1;
-- name: GetUserFromRefreshToken :one
SELECT * FROM refreshToken WHERE token = $1;
