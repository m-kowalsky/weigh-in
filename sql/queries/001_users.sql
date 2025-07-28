-- name: CreateUser :one
Insert into users (created_at, updated_at, email, access_token, full_name, provider)
Values ( ?, ?, ?, ?, ?, ?)
Returning *;

-- name: GetUserByEmail :one
Select * From users
Where email = ?;

-- name: CheckIfUserExistsByEmail :one
Select count(*) from users where email = ?;

-- name: GetUserById :one
Select * from users
Where id = ?;
