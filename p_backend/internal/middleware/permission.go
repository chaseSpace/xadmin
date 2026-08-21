package middleware

import (
	authorizationrepo "monorepo/internal/repo/authorization"
	"monorepo/pkg/xerr"
	"monorepo/pkg/xfiber"

	"github.com/gofiber/fiber/v2"
)

// RequirePermission returns a middleware that checks if the user has ALL specified permission keys.
func RequirePermission(keys ...string) fiber.Handler {
	repo := authorizationrepo.NewRepo()
	return func(c *fiber.Ctx) error {
		uid := GetUID(c)
		if uid == 0 {
			return xfiber.StdResponse(c, nil, xerr.NewBiz(xerr.CodeUnauthorized, "auth.not_logged_in"))
		}

		userKeys, err := repo.ListEffectivePermissionKeysByUIDForMiddleware(c.UserContext(), uid)
		if err != nil {
			return xfiber.StdResponse(c, nil, xerr.NewBiz(xerr.CodeInternalError, "auth.permission_load_failed"))
		}

		for _, key := range keys {
			if _, has := userKeys[key]; !has {
				return xfiber.StdResponse(c, nil, xerr.NewBiz(xerr.CodeForbidden, "auth.no_permission"))
			}
		}
		return c.Next()
	}
}
