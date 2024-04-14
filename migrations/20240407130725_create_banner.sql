-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS banner(
    id bigserial,
    banner_feature bigserial not null references feature(id) on delete cascade,
    content json not null,
    is_active bool default false,
    created_at timestamp default now(),
    updated_at timestamp,
    constraint banner_pk primary key(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS banner;
-- +goose StatementEnd
