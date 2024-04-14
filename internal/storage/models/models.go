package models

import (
	"encoding/json"
)

type BannerFull struct {
	BannerId  int64           `db:"banner_id" json:"banner_id"`
	TagIds    []int64         `json:"tag_ids"`
	FeatureId int64           `db:"feature_id" json:"feature_id"`
	Content   json.RawMessage `db:"content" json:"content"`
	IsActive  bool            `db:"is_active" json:"is_active"`
	CreatedAt string          `db:"created_at" json:"created_at"`
	UpdatedAt string          `db:"updated_at" json:"updated_at"`
}
