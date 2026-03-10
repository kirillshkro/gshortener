-- +goose Up
SELECT 'up SQL query';
create database if not exists shortener;
-- +goose Down
SELECT 'down SQL query';

drop database if exists shortener;