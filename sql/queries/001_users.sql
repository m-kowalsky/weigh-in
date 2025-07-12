-- name: CreateUser :one
Insert into users (created_at, updated_at, email, access_token, full_name)
Values ( ?, ?, ?, ?, ?)
Returning *;

-- name: GetUserByEmail :one
Select * From users
Where email = ?;

-- name: CheckIfUserExistsByEmail :one
Select count(*) from users where email = ?;
