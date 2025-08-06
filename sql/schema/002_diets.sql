-- +goose Up
Create table diets (
id integer primary key AUTOINCREMENT,
diet_type text not null,
user_id integer not null,
foreign key (user_id) references users(id) on delete cascade
);

-- +goose Down
Drop table diets;
