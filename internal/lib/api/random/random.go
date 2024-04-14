package random

import (
	"fmt"
	"github.com/Memozir/BannerService/config"
	"github.com/Memozir/BannerService/internal/http-server/middlewares/auth"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func GetTestTokens(cfg *config.Config) (string, string) {
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
