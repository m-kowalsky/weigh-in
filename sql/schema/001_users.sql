-- +goose Up
Create table users (
id integer primary key AUTOINCREMENT,
created_at timestamp not null,
updated_at timestamp not null,
email text not null unique,
access_token text not null,
full_name text,
provider text not null
);

-- +goose Down
Drop table users;

