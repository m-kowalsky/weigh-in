-- +goose Up
Alter table users
add column provider text not null;

-- +goose Down
Alter table users
Drop column provider;
