-- name: CreateUser :one
Insert into users (created_at, updated_at, email)
Values ( ?, ?, ? )
Returning *;


