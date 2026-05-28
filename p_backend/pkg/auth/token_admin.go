package auth

import (
	"monorepo/config"
	"time"

	"monorepo/pkg/xerr"

	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/lo"
)

type ClaimsAdmin struct {
	UID int64 `json:"uid"`
	jwt.RegisteredClaims
}

func VerifyAdminToken(name, secret, ip string, externalAdminUID int64) (int64, error) {
	cfg := config.GetConfig().App.AuthAdmin
	for _, v := range cfg.AuthorizedIdentity {
		if v.Name == name && v.Secret == secret {
			if len(v.AllowedIPs) == 0 {
				return externalAdminUID, nil
			}
			if lo.Contains(v.AllowedIPs, ip) {
				return externalAdminUID, nil
			}
		}
	}
	return 0, xerr.NewWithDetail(xerr.CodeUnauthorized, "invalid credentials")
}

func tokenTTL(authConfig config.AuthConfig) time.Duration {
	if authConfig.TokenTTLHours <= 0 {
		return 24 * time.Hour
	}
	return time.Duration(authConfig.TokenTTLHours) * time.Hour
}
