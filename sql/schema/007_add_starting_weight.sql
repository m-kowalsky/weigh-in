-- +goose Up
Alter table users
Add column starting_weight integer;

-- +goose Down
Alter table users
Drop column starting_weight;
