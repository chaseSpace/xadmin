package middleware

import (
	authorizationsvc "monorepo/internal/service/authorization"
	"monorepo/pkg/xfiber"

	"github.com/gofiber/fiber/v2"
)

func RequireSuperAdmin() fiber.Handler {
	service := authorizationsvc.NewService()
	return func(c *fiber.Ctx) error {
		if err := service.EnsureSuperAdmin(c.UserContext(), GetUID(c)); err != nil {
			return xfiber.StdResponse(c, nil, err)
		}
		return c.Next()
	}
}
