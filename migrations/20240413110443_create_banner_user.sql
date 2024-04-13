-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS banner_user(
    id bigserial,
    user_role serial references user_role(id) on delete set null,
    constraint banner_user_pk primary key(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS banner_user;
-- +goose StatementEnd
