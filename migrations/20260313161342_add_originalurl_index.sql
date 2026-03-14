-- +goose Up
SELECT 'up SQL query';
alter table urls create unique index original_url_idx on urls (original_url);
-- +goose Down
SELECT 'down SQL query';
drop index if exists original_url_idx;