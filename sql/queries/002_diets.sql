-- name: CreateDiet :one
Insert into diets ( diet_type, user_id )
Values ( ?, ? )
Returning *;
