// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: 002_diets.sql

package database

import (
	"context"
	"database/sql"
)

const createDiet = `-- name: CreateDiet :one
Insert into diets ( diet_type )
Values ( ? )
Returning id, diet_type
`

func (q *Queries) CreateDiet(ctx context.Context, dietType sql.NullString) (Diet, error) {
	row := q.db.QueryRowContext(ctx, createDiet, dietType)
	var i Diet
	err := row.Scan(&i.ID, &i.DietType)
	return i, err
}
