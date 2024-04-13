-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
create table if not exists user_role(
    id serial,
    name varchar(20) unique not null,
    role_code int4 unique not null,
    constraint user_role_pk primary key(id)
);
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_role;
-- +goose StatementEnd
