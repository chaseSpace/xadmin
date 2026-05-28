package middleware

import (
	"strings"
	"time"

	"monorepo/internal/support/ipblacklist"
	"monorepo/pkg/logger"
	"monorepo/pkg/xerr"
	"monorepo/pkg/xfiber"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type IPBlacklistMatcher interface {
	Match(ip string, now time.Time) (ipblacklist.Match, bool)
}

func IPBlacklist(matchers ...IPBlacklistMatcher) fiber.Handler {
	var matcher IPBlacklistMatcher
	if len(matchers) > 0 {
		matcher = matchers[0]
	}
	if matcher == nil {
		matcher = ipblacklist.DefaultStore()
	}

	return func(c *fiber.Ctx) error {
		if shouldSkipIPBlacklist(c.Path()) {
			return c.Next()
		}

		match, blocked := matcher.Match(c.IP(), time.Now())
		if blocked {
			logger.Warn("Request blocked by IP blacklist",
				zap.String("ip", match.IP),
				zap.Int64("blacklist_id", match.ID),
				zap.String("method", c.Method()),
				zap.String("path", c.Path()),
			)
			return xfiber.StdResponse(c, nil, xerr.NewBiz(xerr.CodeForbidden, "auth.ip_blocked"))
		}

		return c.Next()
	}
}

func shouldSkipIPBlacklist(path string) bool {
	return false
	normalized := strings.TrimRight(strings.ToLower(strings.TrimSpace(path)), "/")
	return normalized == "/v1/system/ip-blacklist" || strings.HasPrefix(normalized, "/v1/system/ip-blacklist/")
}
