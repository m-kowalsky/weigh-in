-- +goose Up
Alter table users
Add column weight_unit text not null default "lbs";

-- +goose Down
Alter table users
Drop column weight_unit;
