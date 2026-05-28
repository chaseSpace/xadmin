package db

import (
	"context"
	"fmt"
	"log"
	"monorepo/config"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/k0kubun/pp/v3"
)

// 全局 Redis 客户端变量
var (
	redisClient *redis.Client
	redisOnce   sync.Once
)

// MustInitRedis 初始化 Redis 客户端，使用单例模式
func MustInitRedis() {
	redisOnce.Do(func() {
		loadError := initRedis(&redisClient)
		if loadError != nil {
			log.Fatalf("Failed to init Redis: %v", loadError)
		}
		pp.Printf("%s Successfully initialized %s\n", time.Now(), "Redis")
	})
}

// GetRedis 获取 Redis 客户端实例
func GetRedis() *redis.Client {
	if redisClient == nil {
		MustInitRedis()
	}
	return redisClient
}

// initRedis 内部初始化 Redis 客户端
func initRedis(client **redis.Client) (err error) {
	// 从配置中获取 Redis 配置
	cfg := config.GetConfig()

	// 初始化 Redis 客户端连接
	redisInst := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.App.Redis.Host, cfg.App.Redis.Port),
		Password: cfg.App.Redis.Password,
		DB:       cfg.App.Redis.DB,
		PoolSize: cfg.App.Redis.PoolSize,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err = redisInst.Ping(ctx).Result()
	if err == nil {
		*client = redisInst
	}
	return err
}

// CloseRedis 关闭 Redis 连接
func CloseRedis() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}
