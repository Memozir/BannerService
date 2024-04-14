package delete

import (
	"context"
	respHelper "github.com/Memozir/BannerService/internal/lib/api/response"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
)

type StorageError interface {
	GetDbError() string
}

type BannerDeleter interface {
	DeleteBanner(ctx context.Context, bannerId int64) error
}

func New(storage BannerDeleter, log *slog.Logger) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		const op = "handlers.banner.update.New"

		bannerId := chi.URLParam(r, "id")

		if len(bannerId) == 0 {
			respHelper.SetNotFound(rw)
			respError := respHelper.Error(respHelper.IncorrectDataMsg)
			render.JSON(rw, r, &respError)
			return
		}

		bannerIdInt, err := strconv.Atoi(bannerId)

		if err != nil {
			respHelper.SetInternalServerError(rw)
			apiErr := respHelper.Error(err.Error())
			render.JSON(rw, r, apiErr)
			return
		}

		err = storage.DeleteBanner(r.Context(), int64(bannerIdInt))

		if err != nil {
			log.Error(err.Error(), slog.String("op", op))

			if sErr, ok := err.(StorageError); ok {
				respError := respHelper.Error(sErr.GetDbError())
				respHelper.SetInternalServerError(rw)
				render.JSON(rw, r, &respError)
				return
			}
			respHelper.SetNotFound(rw)
			respError := respHelper.Error(respHelper.IncorrectDataMsg)
			render.JSON(rw, r, &respError)
			return
		}
		respHelper.SetDeleted(rw)
	}
}
