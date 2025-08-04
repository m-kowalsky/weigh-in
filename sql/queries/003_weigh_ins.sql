-- name: CreateWeighIn :one
Insert into weigh_ins (created_at, updated_at, weight, weight_unit, log_date, note, cheated, alcohol, weigh_in_diet, user_id)
Values ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
Returning *;

-- name: GetWeightChartDataByUser :many
Select log_date, weight from weigh_ins
where user_id = ?
order by log_date;
