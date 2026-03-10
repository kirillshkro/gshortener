-- +goose Up
SELECT 'up SQL query';
create unique index if not exists idx_short_url on urls (short_url);
-- +goose Down
SELECT 'down SQL query';

drop index if exists idx_short_url;