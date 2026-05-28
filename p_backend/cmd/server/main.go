package main

import (
	"context"
	"fmt"
	"log"
	"monorepo/config"
	"monorepo/internal/bootstrap"
	accounthandler "monorepo/internal/handler/account"
	assetshandler "monorepo/internal/handler/assets"
	authhandler "monorepo/internal/handler/auth"
	organizationhandler "monorepo/internal/handler/organization"
	permissionhandler "monorepo/internal/handler/permission"
	resourcehandler "monorepo/internal/handler/resource"
	systemhandler "monorepo/internal/handler/system"
	"monorepo/internal/job"
	mw "monorepo/internal/middleware"
	systemsvc "monorepo/internal/service/system"
	"monorepo/pkg/db"
	"monorepo/pkg/logger"
	"monorepo/pkg/xerr"
	fiber2 "monorepo/pkg/xfiber"
	"time"

	"github.com/caddyserver/certmagic"
	"github.com/chasespace/goutil"
	"github.com/chasespace/goutil/uip"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/google/uuid"
	"github.com/libdns/namedotcom"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg := config.MustLoadConfig()
	bootstrap.InitProgramTimezone(cfg.App.Server.Timezone)

	var err error
	if err = logger.Init(&cfg.RequestLog, &cfg.AppLog); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	uip.MustInitIp2Region(cfg.App.Misc.Ip2regionXdbPath)
	db.MustInitDatabase()
	db.MustInitRedis()
	setupAutoCert(cfg.App.Server.IsProd(), cfg.App.Server.DomainCert)
	if err = systemsvc.LoadIPBlacklistStore(context.Background()); err != nil {
		logger.Fatal("Failed to load IP blacklist store", zap.Error(err))
	}

	fmt.Println() // 空行占位

	scheduler := job.NewScheduler()
	if err = scheduler.Start(); err != nil {
		logger.Fatal("Failed to start scheduler", zap.Error(err))
	}

	// 创建 Fiber App
	app := fiber.New(fiber.Config{
		ServerHeader:      "mono",
		EnablePrintRoutes: false,
	})

	{
		app.Use(mw.Recover())
		app.Use(mw.QPSLimiter())
		app.Use(mw.RequestResponseLogger())
		app.Use(mw.IPBlacklist())
	}

	// 路由注册
	v1 := app.Group("/v1", mw.QPSLimiter(mw.UserLimiterConfig))
	{
		authhandler.RegisterRoutes("/auth", v1, mw.Authorize())
		accounthandler.RegisterRoutes("/account", v1, mw.Authorize())
		organizationhandler.RegisterRoutes("/organization", v1, mw.Authorize())
		permissionhandler.RegisterRoutes("/permission", v1, mw.Authorize())
		resourcehandler.RegisterRoutes("/resource", v1, mw.Authorize())
		systemhandler.RegisterRoutes("/system", v1, mw.Authorize())
		assetshandler.RegisterRoutes("/assets", v1, mw.Authorize(), mw.OptionalAuthorize()) // 文件访问按资源元数据决定是否强制鉴权

	}

	if cfg.App.Misc.PrintRouteOnStart {
		fiber2.PrintUsefulRoutes(app)
	}

	fmt.Println() // 空行占位

	app.Use(func(ctx *fiber.Ctx) error { // 404 Handler , 必须放在所有路由后
		return fiber2.StdResponse(ctx, nil, xerr.NewWithDetail(xerr.CodeNotFound, "request path not found!"))
	})

	var errs = make(chan error)
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.App.Server.Host, cfg.App.Server.Port)
		logger.Info("Starting server", zap.String("address", addr))

		if len(cfg.App.Server.DomainCert.Domain) == 0 {
			errs <- app.Listen(addr)
		} else if cfg.App.Server.IsDev() {
			errs <- app.Listen(addr)
		} else {
			certmagic.HTTPSPort = cfg.App.Server.Port
			errs <- certmagic.HTTPS(cfg.App.Server.DomainCert.Domain, adaptor.FiberApp(app))
		}
	}()

	// 监听退出信号并执行优雅关闭
	goutil.Listening(30*time.Second, errs, onClose(app, scheduler))
}

// 三方库的初始化，不依赖项目包启动顺序
func init() {
	uuid.EnableRandPool()
}

func setupAutoCert(isProd bool, cfg config.DomainCert) {
	// 免去CLI 询问
	certmagic.DefaultACME.Agreed = true

	certmagic.DefaultACME.Email = cfg.AcmeEmail

	// 证书缓存路径
	certmagic.Default.Storage = &certmagic.FileStorage{Path: cfg.CertDir}

	if !isProd { // 避免在开发过程中消耗 Let's Encrypt 的速率限制，但是发放的证书无效
		//certmagic.DefaultACME.CA = certmagic.LetsEncryptStagingCA
	}
	certmagic.DefaultACME.DNS01Solver = &certmagic.DNS01Solver{
		DNSManager: certmagic.DNSManager{
			DNSProvider: &namedotcom.Provider{
				Token:  cfg.AcmeToken,
				User:   cfg.AcmeUser,
				Server: cfg.AcmeServer,
			},
		},
	}
}

func onClose(app *fiber.App, scheduler *job.Scheduler) func(ctx context.Context) {
	return func(ctx context.Context) {
		// 执行 Fiber 应用的优雅关闭
		if err := app.ShutdownWithContext(ctx); err != nil {
			logger.Error("Error during shutdown", zap.Error(err))
		} else {
			logger.Info("Server shutdown completed")
		}

		if scheduler != nil {
			scheduler.Stop(ctx)
		}

		_ = db.CloseDatabase()
		_ = db.CloseRedis()
		uip.CloseIp2Region()
	}
}
