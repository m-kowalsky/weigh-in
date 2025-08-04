-- +goose Up
Create table weigh_ins (
id integer primary key AUTOINCREMENT,
created_at timestamp not null, 
updated_at timestamp not null, 
weight integer not null, 
weight_unit text not null,
log_date timestamp not null,
note text,
cheated boolean not null default false,
alcohol boolean not null default false,
weigh_in_diet text not null,
user_id integer not null,
foreign key(user_id) references users(id) on delete cascade
);

-- +goose Down
Drop table weigh_ins;
