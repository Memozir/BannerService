-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tag(
    id BIGSERIAL,
    tag_id bigint not null unique,
    constraint tag_pk primary key(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE if exists tag;
-- +goose StatementEnd
