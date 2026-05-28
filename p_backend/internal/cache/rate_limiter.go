package cache

import (
	"context"
	"fmt"
	"monorepo/pkg/db"
	"monorepo/pkg/xerr"
	"time"
)

const (
	RateLimitKeyPrefix = "rate_limit:%s"
)

type RateLimitResult struct {
	Allowed    bool
	Remaining  int64
	ResetAfter time.Duration
}

func getRateLimitKey(identifier string) string {
	return fmt.Sprintf(RateLimitKeyPrefix, identifier)
}

// Lua 脚本：原子性地执行 INCR 和 EXPIRE
// 返回值：当前计数器值
const rateLimitLuaScript = `
local current = redis.call("INCR", KEYS[1])
if current == 1 then
	redis.call("EXPIRE", KEYS[1], ARGV[1])
end
return current
`

// Allow 检查并执行限流，使用滑动窗口算法
// identifier: 限流标识（如IP、用户ID等）
// limit: 时间窗口内允许的最大请求数
// window: 时间窗口大小
func Allow(ctx context.Context, identifier string, limit int64, window time.Duration) (*RateLimitResult, error) {
	key := getRateLimitKey(identifier)
	redis := db.GetRedis()

	// 使用 Lua 脚本原子性地执行 INCR 和 EXPIRE
	currentCount, err := redis.Eval(ctx, rateLimitLuaScript, []string{key}, int64(window.Seconds())).Result()
	if err != nil {
		return nil, xerr.WrapRedis(err, "限流检查失败")
	}

	count := currentCount.(int64)
	allowed := count <= limit

	// 计算剩余配额
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	// 获取 key 的 TTL，即距离重置的时间
	ttl, err := redis.TTL(ctx, key).Result()
	if err != nil {
		// 如果获取 TTL 失败，使用 window 作为默认值
		ttl = window
	}

	// 如果 key 不存在或已过期（TTL 为 -2 或 -1），说明是新窗口
	if ttl == -2 { // key 不存在
		ttl = window
	} else if ttl == -1 { // key 存在但没有设置过期时间（理论上不会发生，因为 Lua 脚本已经设置了）
		ttl = window
	}

	resetAfter := ttl
	if resetAfter < 0 {
		resetAfter = 0
	}

	return &RateLimitResult{
		Allowed:    allowed,
		Remaining:  remaining,
		ResetAfter: resetAfter,
	}, nil
}
