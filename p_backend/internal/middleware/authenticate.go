package middleware

import (
	"monorepo/config"
	"monorepo/internal/model"
	authsvc "monorepo/internal/service/auth"
	"monorepo/pkg/auth"
	"monorepo/pkg/logger"
	"monorepo/pkg/xerr"
	"monorepo/pkg/xfiber"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// AuthCtx 认证后的上下文，空结构体作为key
type AuthCtx struct {
	User      *model.AdminUser
	SessionID string
}

func GetUserEntity(c *fiber.Ctx) *AuthCtx {
	if info, ok := c.Locals(AuthCtx{}).(*AuthCtx); ok {
		return info
	}
	return nil
}

func GetUID(c *fiber.Ctx) int32 {
	if ctx := GetUserEntity(c); ctx != nil {
		return ctx.User.UID
	}
	return 0
}

func GetSessionID(c *fiber.Ctx) string {
	if ctx := GetUserEntity(c); ctx != nil {
		return ctx.SessionID
	}
	return ""
}

// MustUID 获取当前请求中的用户 UID，如果无法获取则 panic
func MustUID(c *fiber.Ctx) int32 {
	uid := GetUID(c)
	if uid == 0 {
		panic("user not authenticated or UID is invalid")
	}
	return uid
}

func Authorize() fiber.Handler {
	svc := authsvc.NewService()

	return func(c *fiber.Ctx) error {
		// 跳过指定路径
		allowSkip := false
		for _, skipPath := range config.GetConfig().App.Auth.SkipPathSuffix {
			if strings.Contains(c.Path(), skipPath) {
				allowSkip = true
				break
			}
		}

		token := strings.TrimSpace(c.Get("Authorization"))
		if token == "" {
			if !allowSkip {
				return xfiber.StdResponse(c, nil, xerr.NewWithDetail(xerr.CodeUnauthorized, "missing authorization token"))
			}
			return c.Next() // no auth
		}
		token = normalizeBearerToken(token)

		claims, err := auth.VerifyTokenClaims(c.UserContext(), token)
		if err != nil {
			logger.Warn("Token verification failed",
				zap.String("token_hash", auth.HashToken(token)),
				zap.Error(err),
				zap.String("ip", c.IP()),
				zap.String("path", c.Path()),
			)
			return xfiber.StdResponse(c, nil, err)
		}

		if claims.UID <= 0 || strings.TrimSpace(claims.SessionID) == "" {
			return xfiber.StdResponse(c, nil, xerr.NewWithDetail(xerr.CodeUnauthorized, "invalid uid in token"))
		}

		ok, err := svc.IsSessionActive(c.UserContext(), claims.UID, claims.SessionID, auth.HashToken(token))
		if err != nil {
			return xfiber.StdResponse(c, nil, err)
		}
		if !ok {
			return xfiber.StdResponse(c, nil, xerr.NewWithDetail(xerr.CodeUnauthorized, "session revoked or expired"))
		}

		c.Locals(AuthCtx{}, &AuthCtx{
			User:      &model.AdminUser{UID: claims.UID},
			SessionID: claims.SessionID,
		})

		return c.Next()
	}
}

func OptionalAuthorize() fiber.Handler {
	svc := authsvc.NewService()

	return func(c *fiber.Ctx) error {
		token := strings.TrimSpace(c.Get("Authorization"))
		if token == "" {
			return c.Next()
		}
		token = normalizeBearerToken(token)

		claims, err := auth.VerifyTokenClaims(c.UserContext(), token)
		if err != nil {
			return c.Next()
		}
		if claims.UID <= 0 || strings.TrimSpace(claims.SessionID) == "" {
			return c.Next()
		}

		ok, err := svc.IsSessionActive(c.UserContext(), claims.UID, claims.SessionID, auth.HashToken(token))
		if err != nil || !ok {
			return c.Next()
		}

		c.Locals(AuthCtx{}, &AuthCtx{
			User:      &model.AdminUser{UID: claims.UID},
			SessionID: claims.SessionID,
		})
		return c.Next()
	}
}

func normalizeBearerToken(token string) string {
	if len(token) < 7 {
		return token
	}
	if strings.EqualFold(token[:7], "Bearer ") {
		return strings.TrimSpace(token[7:])
	}
	return token
}
