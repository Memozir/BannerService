package main

import (
	"context"
	"fmt"
	"github.com/Memozir/BannerService/config"
	"github.com/Memozir/BannerService/internal/cache/redis"
	"github.com/Memozir/BannerService/internal/http-server/handlers/banner/create"
	"github.com/Memozir/BannerService/internal/http-server/handlers/banner/get"
	"github.com/Memozir/BannerService/internal/http-server/middlewares/auth"
	"github.com/Memozir/BannerService/internal/storage/postgres"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/golang-jwt/jwt/v5"
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

func getTestTokens(cfg *config.Config) (string, string) {
	secretKey := []byte(cfg.SecretKey)
	expTime := time.Now().Add(24 * 30 * time.Hour)

	claimsAdmin := &auth.Claims{
		Role: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
		},
	}

	claimsUser := &auth.Claims{
		Role: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
		},
	}

	tokenAdmin := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsAdmin)
	tokenStringAdmin, err := tokenAdmin.SignedString(secretKey)

	if err != nil {
		fmt.Println(err.Error())
		return "", ""
	}

	tokenUser := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsUser)
	tokenStringUser, err := tokenUser.SignedString(secretKey)

	if err != nil {
		fmt.Println(err.Error())
		return "", ""
	}

	return tokenStringAdmin, tokenStringUser
}

func main() {
	appConfig := config.New()
	logger := NewLogger(appConfig)
	logger.Info("Debug", slog.String("logLevel", appConfig.LogLevel))

	adminToken, userToken := getTestTokens(appConfig)
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
