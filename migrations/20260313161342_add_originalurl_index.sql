-- +goose Up
SELECT 'up SQL query';
create unique index original_url_idx on urls (original_url);
-- +goose Down
SELECT 'down SQL query';
drop index original_url_idx;