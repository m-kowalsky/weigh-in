-- name: CreateWeighIn :one
Insert into weigh_ins (created_at, updated_at, weight, weight_unit, log_date, log_date_display, note, cheated, alcohol, weigh_in_diet, user_id)
Values ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
Returning *;

-- name: GetWeightChartDataByUser :many
Select log_date, weight from weigh_ins
where user_id = ? and log_date >= ?
order by log_date;

-- name: GetWeighInsByUser :many
Select * from weigh_ins
where user_id = ?
order by log_date asc;

-- name: GetWeighInById :one
Select * from weigh_ins
where id = ?;

-- name: UpdateWeighIn :exec
Update weigh_ins
Set updated_at = ?, weight = ?, weight_unit = ?, log_date = ?, log_date_display = ?, note = ?, cheated = ?, alcohol = ?,
weigh_in_diet = ?
Where id = ?;
