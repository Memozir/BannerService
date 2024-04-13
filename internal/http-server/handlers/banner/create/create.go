package create

import (
	"context"
	"encoding/json"
	responseHelper "github.com/Memozir/BannerService/internal/lib/api/response"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

type BannerCreator interface {
	CreateBanner(ctx context.Context, tagIds []int64, featureId int64, content string, isActive bool) (int64, error)
}

type StorageError interface {
	GetDbError() string
}

type request struct {
	TagIds    []int64         `json:"tag_ids"`
	FeatureId int64           `json:"feature_id"`
	Content   json.RawMessage `json:"content"`
	IsActive  bool            `json:"is_active"`
}

type response struct {
	BannerId int64 `json:"banner_id"`
}

func New(log *slog.Logger, bannerCreator BannerCreator) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		const op = "handlers.banner.create.New"
		var req request
		err := render.Decode(r, &req)

		if err != nil {
			log.Error(err.Error(), slog.String("op", op))
			responseHelper.SetBadRequest(rw)
			return
		}

		bannerId, err := bannerCreator.CreateBanner(
			r.Context(),
			req.TagIds,
			req.FeatureId,
			string(req.Content),
			req.IsActive,
		)

		if err != nil {
			log.Error(err.Error(), slog.String("op", op))

			if sErr, ok := err.(StorageError); ok {
				respError := responseHelper.Error(sErr.GetDbError())
				responseHelper.SetInternalServerError(rw)
				render.JSON(rw, r, &respError)
				return
			}
			responseHelper.SetNotFound(rw)
			respError := responseHelper.Error(responseHelper.IncorrectDataMsg)
			render.JSON(rw, r, &respError)
			return
		}

		resp := response{
			BannerId: bannerId,
		}

		responseHelper.SetCreated(rw)
		render.JSON(rw, r, &resp)
	}
}
