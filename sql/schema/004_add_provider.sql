-- +goose Up
Alter table users
Add column provider text;

-- +goose Down
Alter table users
Drop column provider;
