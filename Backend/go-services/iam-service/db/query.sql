-- name: GetUserForAuth :one
SELECT id, login, password_hash, is_head, system_role, department_id FROM users WHERE login = $1;

-- name: IsLoginExistInDB :one
SELECT id FROM users WHERE login = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT id, login, email, full_name, department_id, is_head, system_role FROM users WHERE id = $1;

-- name: GetUserByLogin :one
SELECT id, login, email, full_name, department_id, is_head, system_role FROM users WHERE login = $1;

-- name: ListUsersByDepartment :many
SELECT id, login, full_name, is_head FROM users WHERE department_id = $1;

-- name: ListUsers :many
SELECT id, full_name, login, is_head, department_id FROM users;

-- name: UpdateDepartmentParent :exec
UPDATE departments SET parent_id = $2 WHERE id = $1;

-- name: ListDepartments :many
SELECT id, name, parent_id FROM departments ORDER BY name ASC;

-- name: CreateDepartment :one
INSERT INTO departments (name, parent_id) VALUES ($1, $2) RETURNING id;

-- name: FireUser :exec
UPDATE users SET department_id = NULL WHERE id = $1;

-- name: MoveUserToDepartment :exec
UPDATE users SET department_id = $2 WHERE id = $1;

-- name: DeleteDepartment :exec
DELETE FROM departments WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (login, password_hash, email, full_name, department_id, is_head, system_role)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;

-- name: PromoteUser :exec
UPDATE users SET is_head = true WHERE id = $1;

-- name: DemoteUser :exec
UPDATE users SET is_head = false WHERE id = $1;

-- name: GetManagersByDepartment :many
SELECT id, login, email, full_name FROM users WHERE is_head = true AND department_id = $1;
