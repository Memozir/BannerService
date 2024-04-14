-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS feature(
    id bigserial,
    feature_id bigint not null unique,
    constraint feature_pk primary key(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE if exists feature;
-- +goose StatementEnd
