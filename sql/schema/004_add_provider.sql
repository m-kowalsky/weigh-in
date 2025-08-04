-- +goose Up
Alter table users
Add column provider text not null default "google";

-- +goose Down
Alter table users
Drop column provider;
