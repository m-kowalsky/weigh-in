-- name: CreateWeighIn :one
Insert into weigh_ins (created_at, updated_at, weight, weight_unit, log_date, note, cheated, alcohol, weigh_in_diet)
Values ( ?, ?, ?, ?, ?, ?, ?, ?, ?)
Returning *;
