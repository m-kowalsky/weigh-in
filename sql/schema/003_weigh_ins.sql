-- +goose Up
Create table weigh_ins (
id integer primary key AUTOINCREMENT,
created_at timestamp not null, 
updated_at timestamp not null, 
weight integer, weight_unit text,
log_date timestamp not null,
note text,
cheated boolean not null default false,
alcohol boolean not null default false,
weigh_in_diet text,
foreign key(weigh_in_diet) references diets(diet_type)
);

-- +goose Down
Drop table diets
