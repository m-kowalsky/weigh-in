-- +goose Up
Create table users (
id integer primary key AUTOINCREMENT,
created_at timestamp not null,
updated_at timestamp not null,
email text unique
);

-- +goose Down
Drop table users



	-- users_init := "Create table if not exists users ( id integer not null primary key autoincrement, created_at timestamp not null, updated_at timestamp not null, email text unique);"
	--
	-- weigh_ins_init := "Create table if not exists weigh_ins ( id integer not null primary key autoincrement, created_at timestamp not null, updated_at timestamp not null, weight integer, weight_unit text, log_date timestamp not null, note text, cheated boolean defaultfalse, alcohol boolean default false, weigh_in_diet text, foreign key(weigh_in_diet) references diet(diet_type);"
	--
	-- diet_init := "Create table if not exists diet ( id integer not null primary key autoincrement, diet_type text);"
