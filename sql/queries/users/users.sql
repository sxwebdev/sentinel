-- name: GetByEmail :one
SELECT * FROM users WHERE email=? LIMIT 1;

-- name: CheckRootUserExists :one
SELECT EXISTS (SELECT 1 FROM users WHERE role='root' LIMIT 1);
