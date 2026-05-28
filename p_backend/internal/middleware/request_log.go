package middleware

import (
	"monorepo/internal/support/requestmeta"
	"monorepo/pkg/logger"
	"monorepo/pkg/xi18n"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"monorepo/util"
)

// RequestResponseLogger 创建一个记录请求和响应日志的中间件
func RequestResponseLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 检查请求必须含 X-Trace-ID
		if strings.TrimSpace(c.Get("X-Trace-ID")) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    400,
				"message": "missing X-Trace-ID header",
			})
		}

		// 获取请求体
		reqBody := string(c.Request().Body())
		truncatedReqBody := util.TruncateString(reqBody, 300) // 限制请求体长度为100字符

		rid := util.NewRequestId()

		// 请求前打印日志
		logger.ReqLogger.Info("Request incoming",
			zap.String("method", c.Method()),
			zap.String("path", c.OriginalURL()),
			zap.String("ip", c.IP()),
			zap.Int("request_body_len", len(reqBody)),
			zap.String("request_id", rid),
		)

		// 记录开始时间
		start := time.Now()
		c.SetUserContext(requestmeta.WithStartTime(c.UserContext(), start))
		c.SetUserContext(requestmeta.WithRequestID(c.UserContext(), rid))
		c.SetUserContext(xi18n.WithLang(c.UserContext(), c.Get("Accept-Language")))

		// 处理请求
		err := c.Next()

		// 计算处理时间
		duration := time.Since(start)

		// 获取响应状态
		status := c.Response().StatusCode()

		// 获取响应体
		respBody := string(c.Response().Body())

		bodyLen := 300
		if strings.HasSuffix(string(c.Request().URI().Path()), "/GetFile") {
			bodyLen = 10
		}
		truncatedRespBody := util.TruncateString(respBody, bodyLen, "...") // 限制响应体长度为100字符

		// 请求后打印日志
		logger.ReqLogger.Info("Request completed",
			zap.String("method", c.Method()),
			zap.String("path", c.OriginalURL()),
			zap.Int("status", status),
			zap.String("duration", duration.String()),
			zap.Int64("reply_length", int64(len(respBody))),
			zap.String("request_id", rid),
			zap.String("request_body", truncatedReqBody),
			zap.String("response_body", truncatedRespBody),
		)

		return err
	}
}
