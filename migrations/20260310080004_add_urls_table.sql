-- +goose Up
SELECT 'up SQL query';
create table if not exists urls (id integer generated always as identity primary key, 
	short_url varchar(255) not null, 
	original_url text not null);
-- +goose Down
SELECT 'down SQL query';

drop table if exists urls;
