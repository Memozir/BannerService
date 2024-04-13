package get

import (
	"encoding/json"
	"net/http"

	respHelper "github.com/Memozir/BannerService/internal/lib/api/response"
)

type BannerGetter interface {
}

var response json.RawMessage

func New(storage BannerGetter) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		tagId := r.URL.Query().Get("tag_id")
		featureId := r.URL.Query().Get("feature_id")
		//useLastRevision := r.URL.Query().Get("use_last_revision")

		if len(tagId) == 0 && len(featureId) == 0 {
			respHelper.SetBadRequest(rw)
			return
		}
	}
}
