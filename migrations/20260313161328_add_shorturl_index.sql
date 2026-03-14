-- +goose Up
SELECT 'up SQL query';
alter table urls create unique index short_url_idx on urls (short_url);
-- +goose Down
SELECT 'down SQL query';
drop index if exists short_url_idx;