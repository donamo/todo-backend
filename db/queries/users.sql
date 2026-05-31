-- name: GetUserByID :one
SELECT id, google_subject, email, name, approved, created_at, updated_at
FROM users
WHERE id = $1;

-- name: UpsertUser :one
INSERT INTO users (google_subject, email, name, approved)
VALUES ($1, $2, $3, $4)
ON CONFLICT (google_subject) DO UPDATE
SET email = EXCLUDED.email,
    name = EXCLUDED.name,
    approved = users.approved OR EXCLUDED.approved,
    updated_at = now()
RETURNING id, google_subject, email, name, approved, created_at, updated_at;
