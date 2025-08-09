-- +goose Up
Alter table diets
Add column is_default bool not null default false;

-- +goose Down
Alter table diets
Drop column is_default;
