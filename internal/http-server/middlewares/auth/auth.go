package auth

import (
	"context"
	"errors"
	"github.com/Memozir/BannerService/config"
	respHelper "github.com/Memozir/BannerService/internal/lib/api/response"
	"github.com/golang-jwt/jwt/v5"
	"log/slog"
	"net/http"
)

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type CustomMiddleware func(next http.Handler) http.Handler

func NewJWTAuthenticationMiddleware(
	log *slog.Logger, cfg *config.Config,
) CustomMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			const op = "middlewares.auth.AuthenticationJWT"

			tokenStr := r.Header.Get("token")

			if len(tokenStr) == 0 {
				respHelper.SetUnauthorized(rw)
				log.Info("unauthorised")
				return
			}

			claims := new(Claims)
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
				return []byte(cfg.SecretKey), nil
			})

			if err != nil {
				if errors.Is(err, jwt.ErrSignatureInvalid) {
					log.Info("invalid signature of token", slog.String("token", tokenStr))
					respHelper.SetUnauthorized(rw)
				}
				log.Info("token parse error", slog.String("token", tokenStr))
				respHelper.SetUnauthorized(rw)
				return
			}

			if !token.Valid {
				log.Info("invalid token", slog.String("token", tokenStr))
				respHelper.SetUnauthorized(rw)
				return
			}

			ctxValue := context.WithValue(r.Context(), "role", claims.Role)
			r.Clone(ctxValue)
			next.ServeHTTP(rw, r)
		})
	}
}

func NewJWTAuthenticationAdminMiddleware(
	log *slog.Logger, cfg *config.Config,
) CustomMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			const op = "middlewares.auth.AuthenticationJWT"

			tokenStr := r.Header.Get("token")

			if len(tokenStr) == 0 {
				respHelper.SetUnauthorized(rw)
				log.Info("unauthorised")
				return
			}

			claims := new(Claims)
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
				return []byte(cfg.SecretKey), nil
			})

			if err != nil {
				if errors.Is(err, jwt.ErrSignatureInvalid) {
					log.Info("invalid signature of token", slog.String("token", tokenStr))
					respHelper.SetUnauthorized(rw)
				}
				log.Info("token parse error", slog.String("token", tokenStr))
				respHelper.SetUnauthorized(rw)
				return
			}

			if !token.Valid {
				log.Info("invalid token", slog.String("token", tokenStr))
				respHelper.SetUnauthorized(rw)
				return
			}

			if claims.Role != "admin" {
				log.Info("not enough rights", slog.String("token", tokenStr))
				respHelper.SetForbidden(rw)
				return
			}

			ctxValue := context.WithValue(r.Context(), "role", claims.Role)
			r.Clone(ctxValue)
			next.ServeHTTP(rw, r)
		})
	}
}
