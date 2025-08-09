-- name: CreateDiet :one
Insert into diets ( diet_type, user_id, is_default )
Values ( ?, ?, ? )
Returning *;

-- name: GetDietsByUserId :many
Select * from diets
Where user_id = ?;
