-- +goose Up
Alter table users
Add column username text;

-- +goose Down
Alter table users
Drop column username;
