-- +goose Up
Create table diets (
id integer primary key AUTOINCREMENT,
diet_type text not null
);

-- +goose Down
Drop table diets;
