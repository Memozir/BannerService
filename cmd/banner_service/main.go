package main

import (
	"context"
	"fmt"
	"github.com/Memozir/BannerService/config"
	"github.com/Memozir/BannerService/internal/cache/redis"
	"github.com/Memozir/BannerService/internal/http-server/handlers/banner/create"
	"github.com/Memozir/BannerService/internal/http-server/handlers/banner/delete"
	"github.com/Memozir/BannerService/internal/http-server/handlers/banner/get"
	"github.com/Memozir/BannerService/internal/http-server/handlers/banner/update"
	"github.com/Memozir/BannerService/internal/http-server/middlewares/auth"
	content_type "github.com/Memozir/BannerService/internal/http-server/middlewares/content-type"
	"github.com/Memozir/BannerService/internal/lib/api/random"
	"github.com/Memozir/BannerService/internal/storage/postgres"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	LogLevelDebug = "debug"
	LogLevelDev   = "dev"
)

func main() {
	appConfig := config.New()
	logger := NewLogger(appConfig)
	logger.Info("Debug", slog.String("logLevel", appConfig.LogLevel))

	adminToken, userToken := random.GetTestTokens(appConfig)
	fmt.Printf("Admin token: %s\nUser token: %s\n", adminToken, userToken)

	storage, err := postgres.NewDb(context.Background(), logger, appConfig)

	if err != nil {
		logger.Error("enable to connect to database", slog.String("error", err.Error()))
	} else {
		logger.Info("successfully connected to postgres")
	}

	cache := redis.NewRedis(appConfig)
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(content_type.SetContentTypeApplicationJson)

	router.Route("/banner", func(router chi.Router) {
		router.With(
			auth.NewJWTAuthenticationAdminMiddleware(logger, appConfig),
		).Post("/", create.New(logger, storage))

		router.With(auth.NewJWTAuthenticationAdminMiddleware(
			logger, appConfig),
		).Get("/", get.NewAllBanners(storage, logger))
	})

	router.Route("/banner-user", func(router chi.Router) {
		router.With(
			auth.NewJWTAuthenticationMiddleware(logger, appConfig),
		).Get("/", get.New(storage, cache, logger))
	})

	router.Route("/banner/{id}", func(router chi.Router) {
		router.With(
			auth.NewJWTAuthenticationAdminMiddleware(logger, appConfig),
		).Patch("/", update.New(storage, logger))

		router.With(
			auth.NewJWTAuthenticationAdminMiddleware(logger, appConfig),
		).Delete("/", delete.New(storage, logger))
	})

	serverDone := make(chan os.Signal)
	signal.Notify(serverDone, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	serverAddr := fmt.Sprintf("%s:%s", appConfig.ServerHost, appConfig.ServerPort)
	server := http.Server{
		Addr:         serverAddr,
		Handler:      router,
		WriteTimeout: appConfig.ServerTimeout,
		ReadTimeout:  appConfig.ServerTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			logger.Error("server stopping", slog.String("error", err.Error()))
		}
	}()

	logger.Info("server successfully started")
	<-serverDone

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err = server.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", slog.String("error", err.Error()))
	}

	logger.Info("server stopped")
	storage.Shutdown()
	logger.Info("storage shut downed")
}

func NewLogger(cfg *config.Config) *slog.Logger {
	var log *slog.Logger

	switch cfg.LogLevel {
	case LogLevelDebug:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case LogLevelDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
