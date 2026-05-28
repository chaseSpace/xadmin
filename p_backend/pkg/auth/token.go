package auth

import (
	"context"
	"fmt"
	"monorepo/config"
	"time"

	"monorepo/pkg/xerr"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UID       int32  `json:"uid"`
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}

func IssueTokenSignature(ctx context.Context, uid int32, sessionID string) (string, error) {
	if uid < 1 {
		return "", xerr.NewWithDetail(xerr.CodeUnauthorized, "Invalid user id")
	}
	if sessionID == "" {
		return "", xerr.NewWithDetail(xerr.CodeInternalError, "session id required")
	}

	cfg := config.GetConfig()
	secret := cfg.App.Auth.Secret
	if secret == "" {
		return "", xerr.NewWithDetail(xerr.CodeInternalError, "Auth secret not configured")
	}

	now := time.Now()
	expireTime := now.Add(tokenTTL(cfg.App.Auth))

	claims := Claims{
		UID:       uid,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "xadmin",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", xerr.NewWithError(xerr.CodeInternalError, err, "Failed to sign token")
	}

	return tokenString, nil
}

func VerifyToken(ctx context.Context, tokenString string) (int32, error) {
	claims, err := VerifyTokenClaims(ctx, tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UID, nil
}

func VerifyTokenClaims(ctx context.Context, tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, xerr.NewWithDetail(xerr.CodeUnauthorized, "Empty token")
	}

	cfg := config.GetConfig()
	secret := cfg.App.Auth.Secret
	if secret == "" {
		return nil, xerr.NewWithDetail(xerr.CodeInternalError, "Auth secret not configured")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, xerr.NewWithError(xerr.CodeUnauthorized, err, "invalid token")
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, xerr.NewWithDetail(xerr.CodeUnauthorized, "invalid token claims")
}
