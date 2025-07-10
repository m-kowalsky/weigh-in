package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var Db *sql.DB

func openDb() error {

	db, err := sql.Open("sqlite3", "weigh_in.db")
	if err != nil {
		return err
	}
	Db = db
	return nil
}

func closeDb() error {
	return Db.Close()
}

func setupDbSchema() error {

	users_init := "Create table if not exists users ( id integer primary key autoincrement, created_at timestamp not null, updated_at timestamp not null, email text unique);"

	_, err := Db.Exec(users_init)
	if err != nil {
		return err
	}

	weigh_ins_init := `Create table if not exists weigh_ins ( 
	id integer primary key autoincrement,
	created_at timestamp not null,
	updated_at timestamp not null,
	weight integer, weight_unit text,
	log_date timestamp not null,
	note text,
	cheated boolean default false,
	alcohol boolean default false,
	weigh_in_diet text,
	foreign key(weigh_in_diet) references diets(diet_type));`

	_, err = Db.Exec(weigh_ins_init)
	if err != nil {
		return err
	}

	diet_init := "Create table if not exists diets ( id integer primary key autoincrement, diet_type text);"

	_, err = Db.Exec(diet_init)
	if err != nil {
		return err
	}

	return nil
}
