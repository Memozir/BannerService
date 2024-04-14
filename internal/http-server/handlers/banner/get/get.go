package get

import (
	"context"
	"encoding/json"
	"github.com/Memozir/BannerService/internal/cache/redis"
	respHelper "github.com/Memozir/BannerService/internal/lib/api/response"
	"github.com/Memozir/BannerService/internal/storage/models"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type BannerGetter interface {
	GetBanner(ctx context.Context, tagId string, featureId string) (string, bool, error)
	GetAllBanners(
		ctx context.Context, featureId string, tagId string, limit string, offset string,
	) ([]models.BannerFull, error)
}

type CacheStorage interface {
	GetBanner(ctx context.Context, featureId string, tagId string) string
	SetBanner(ctx context.Context, featureId string, tagID string, banner string) error
}

type StorageError interface {
	GetDbError() string
}

func New(storage BannerGetter, cache *redis.RedisCache, log *slog.Logger) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		const op = "handlers.banner.get.New"
		tagId := r.URL.Query().Get("tag_id")
		featureId := r.URL.Query().Get("feature_id")
		useLastRevision := r.URL.Query().Get("use_last_revision")
		role := r.Context().Value("role")
		requestIdStr := middleware.GetReqID(r.Context())
		requestIdSplit := strings.Split(requestIdStr, "-")
		requestId, err := strconv.Atoi(requestIdSplit[len(requestIdSplit)-1])

		if err != nil {
			respHelper.SetInternalServerError(rw)
			apiErr := respHelper.Error(err.Error())
			render.JSON(rw, r, apiErr)
			return
		}

		if len(tagId) == 0 && len(featureId) == 0 {
			respHelper.SetBadRequest(rw)
			return
		}

		if requestId%10 == 0 || useLastRevision == "true" {
			banner, err := cache.GetBanner(r.Context(), featureId, tagId)

			if err == nil {
				var response json.RawMessage = []byte(banner)
				err = json.NewEncoder(rw).Encode(&response)
				if err != nil {
					log.Error("encoding response from cache error", slog.String("op", op))
					respHelper.SetInternalServerError(rw)
					return
				}
			} else {
				log.Error(err.Error(), slog.String("op", op))
			}
		}

		banner, isActive, err := storage.GetBanner(r.Context(), tagId, featureId)

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

		if !isActive && role != "admin" {
			respHelper.SetNotFound(rw)
			return
		}

		err = cache.SetBanner(r.Context(), featureId, tagId, banner)

		if err != nil {
			log.Error(err.Error(), slog.String("op", op))
		}

		var response json.RawMessage = []byte(banner)
		err = json.NewEncoder(rw).Encode(&response)
	}
}

func NewAllBanners(storage BannerGetter, log *slog.Logger) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		const op = "handlers.banner.get.NewAllBanners"

		tagId := r.URL.Query().Get("tag_id")
		featureId := r.URL.Query().Get("feature_id")
		limit := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")

		banners, err := storage.GetAllBanners(r.Context(), featureId, tagId, limit, offset)

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

		render.JSON(rw, r, banners)
	}
}
