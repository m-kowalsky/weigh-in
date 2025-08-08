-- +goose Up
Alter table weigh_ins
Add column log_date_display text not null;

-- +goose Down
Alter table weigh_ins
Drop column log_date_display;
