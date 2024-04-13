-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS banner_tag(
    id bigserial,
    banner_id bigserial references banner(id) on delete cascade,
    tag_id bigserial references tag(id) on delete cascade
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS banner_tag;
-- +goose StatementEnd
