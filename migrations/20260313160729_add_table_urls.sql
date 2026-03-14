-- +goose Up
SELECT 'up SQL query';
create table urls (id serial primary key, short_url text not null, original_url text not null);
-- +goose Down
SELECT 'down SQL query';
drop table if exists urls;