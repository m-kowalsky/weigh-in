-- name: CreateDiet :one
Insert into diets ( diet_type )
Values ( ? )
Returning *;
