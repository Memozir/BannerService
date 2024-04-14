package update

import (
	"context"
	"encoding/json"
	respHelper "github.com/Memozir/BannerService/internal/lib/api/response"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

type BannerUpdater interface {
	UpdateBanner(
		ctx context.Context, bannerId string, tagIds *[]int64, featureId *int64, content string, isActive bool) error
}

type StorageError interface {
	GetDbError() string
}

type request struct {
	TagIds    *[]int64        `json:"tag_ids,omitempty"`
	FeatureId *int64          `json:"feature_id,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
	IsActive  bool            `json:"is_active,omitempty"`
}

func New(storage BannerUpdater, log *slog.Logger) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		const op = "handlers.banner.update.New"

		bannerId := chi.URLParam(r, "id")

		if len(bannerId) == 0 {
			log.Error("no banner id params", slog.String("op", op))
			respHelper.SetBadRequest(rw)
			return
		}

		var bannerUpdate request
		err := json.NewDecoder(r.Body).Decode(&bannerUpdate)

		if err != nil {
			log.Error(
				"cannot decode request data",
				slog.String("err", err.Error()),
				slog.String("op", op))
			respHelper.SetBadRequest(rw)
			return
		}

		err = storage.UpdateBanner(
			r.Context(),
			bannerId,
			bannerUpdate.TagIds,
			bannerUpdate.FeatureId,
			string(bannerUpdate.Content),
			bannerUpdate.IsActive,
		)

		if err != nil {
			if sErr, ok := err.(StorageError); ok {
				log.Error(sErr.GetDbError(), slog.String("op", op))
				respError := respHelper.Error(sErr.GetDbError())
				respHelper.SetInternalServerError(rw)
				render.JSON(rw, r, &respError)
				return
			}
			log.Error(err.Error(), slog.String("op", op))
			respHelper.SetNotFound(rw)
			respError := respHelper.Error(respHelper.IncorrectDataMsg)
			render.JSON(rw, r, &respError)
			return
		}

	}
}
