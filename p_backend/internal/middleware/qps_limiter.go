package middleware

import (
	"context"
	"fmt"
	"monorepo/internal/cache"
	"monorepo/pkg/logger"
	"monorepo/pkg/xerr"
	"monorepo/pkg/xfiber"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type KeyType string

func (k KeyType) newKey(val string) string {
	switch k {
	case KeyTypeIP:
		return "ip:" + val
	case KeyTypePath:
		return "path:" + val
	case KeyTypeUID:
		return "uid:" + val
	}
	return val
}

// 其中IP是必须执行的限流键类型，其他是根据配置执行
const (
	KeyTypeIP   KeyType = "ip"
	KeyTypeUID  KeyType = "uid"
	KeyTypePath KeyType = "path"
)

type PathRule struct {
	PathSuffix []string // 路径后缀匹配
	KeyType    KeyType  // 限流键类型
	Limit      int64    // 时间窗口内允许的最大请求数
	Window     time.Duration
	ErrorMsg   string
}

type QPSLimiterConfig struct {
	DefaultLimit     int64         // 默认限制
	DefaultWindow    time.Duration // 默认时间窗口
	DefaultKeyType   KeyType       // 默认键类型
	PathRules        []PathRule    // 特定路径规则
	SkipPathSuffixes []string      // 跳过限流的路径后缀
	ErrorMsg         string        // 自定义错误消息
}

func (config *QPSLimiterConfig) Setup() {
	if config.DefaultLimit <= 0 {
		config.DefaultLimit = defaultConfig.DefaultLimit
	}
	if config.DefaultWindow <= 0 {
		config.DefaultWindow = defaultConfig.DefaultWindow
	}
	if config.DefaultKeyType == "" {
		config.DefaultKeyType = defaultConfig.DefaultKeyType
	}
	if config.ErrorMsg == "" {
		config.ErrorMsg = defaultConfig.ErrorMsg
	}
}

var defaultConfig = QPSLimiterConfig{
	DefaultLimit:   10,
	DefaultWindow:  time.Second * 1,
	DefaultKeyType: KeyTypeIP,
	ErrorMsg:       "请求较快，请稍后再试",
	PathRules: []PathRule{
		{
			PathSuffix: []string{"/account/UniqueSignIn", "/account/UniqueSignup"},
			KeyType:    KeyTypeIP,
			Limit:      2,
			Window:     time.Second * 5,
		},
		{
			PathSuffix: []string{"/assets/GetFile"},
			KeyType:    KeyTypeUID,
			Limit:      10,
			Window:     time.Second,
		},
	},
}

var UserLimiterConfig = QPSLimiterConfig{
	DefaultLimit:   10,
	DefaultWindow:  time.Second * 5,
	DefaultKeyType: KeyTypeUID,
	PathRules:      []PathRule{},
}

func QPSLimiter(configs ...QPSLimiterConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()

		var config = &defaultConfig
		if len(configs) > 0 {
			config = &configs[0]
			config.Setup()
		}

		// 检查是否需要跳过限流
		for _, skipPath := range config.SkipPathSuffixes {
			if strings.HasSuffix(path, skipPath) {
				return c.Next()
			}
		}

		// 查找匹配的路径规则
		var matchedRule *PathRule
		for _, rule := range config.PathRules {
			for _, pathSuffix := range rule.PathSuffix {
				if strings.HasSuffix(path, pathSuffix) {
					matchedRule = &rule // 匹配成功时赋值
					break
				}
			}
			if matchedRule != nil {
				break // 找到匹配规则后提前退出外层循环
			}
		}

		limit := config.DefaultLimit
		window := config.DefaultWindow
		keyType := config.DefaultKeyType

		if matchedRule != nil {
			if matchedRule.Limit > 0 {
				limit = matchedRule.Limit
			}
			if matchedRule.Window > 0 {
				window = matchedRule.Window
			}
			if matchedRule.KeyType != "" {
				keyType = matchedRule.KeyType
			}
		}

		key := generateKey(c, keyType)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		result, err := cache.Allow(ctx, key, limit, window)
		if err != nil {
			return xfiber.StdResponse(c, nil, xerr.NewWithDetail(xerr.CodeInternalError, "QPS limiter check failed"))
		}

		if !result.Allowed {
			logger.Warn("Request rate limited",
				zap.String("key", key),
				zap.String("key_type", string(keyType)),
				zap.Int64("limit", limit),
				zap.Duration("window", window),
				zap.Int64("remaining", result.Remaining),
				zap.Duration("reset_after", result.ResetAfter),
				zap.String("method", c.Method()),
				zap.String("path", path),
				zap.String("ip", c.IP()),
			)

			return xfiber.StdResponse(c, nil, xerr.NewWithDetail(xerr.CodeTooManyRequests, "%s", config.ErrorMsg))
		}

		return c.Next()
	}
}

func generateKey(c *fiber.Ctx, keyType KeyType) string {
	switch keyType {
	case KeyTypeIP:
	case KeyTypeUID:
		val := c.Locals("uid")
		if val == nil {
			val = c.IP()
		}
		return keyType.newKey(fmt.Sprintf("%v", val))
	case KeyTypePath:
		return keyType.newKey(c.Path())
	}
	return KeyTypeIP.newKey(c.IP())
}
