-- +goose Up
SELECT 'up SQL query';
create unique index if not exists idx_original_url on urls (original_url);
-- +goose Down
SELECT 'down SQL query';
drop index if exists idx_original_url;
