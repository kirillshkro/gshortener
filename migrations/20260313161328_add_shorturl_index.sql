-- +goose Up
SELECT 'up SQL query';
create unique index short_url_idx on urls (short_url);
-- +goose Down
SELECT 'down SQL query';
drop index short_url_idx;