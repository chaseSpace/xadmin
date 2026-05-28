package db

import (
	"fmt"
	"io"
	"log"
	"monorepo/config"
	"monorepo/pkg/logrotate"
	"monorepo/pkg/xerr"
	"monorepo/proto/xadminpb/commpb"
	"monorepo/util"
	"strings"
	"sync"
	"time"

	"github.com/k0kubun/pp/v3"
	"github.com/samber/lo"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 全局 GORM DB 实例变量
var (
	dbInstance *gorm.DB
	once       sync.Once
)

const (
	databaseTypePostgres = "pgsql"
)

func MustInitDatabase() {
	once.Do(func() {
		loadError := initDatabase(&dbInstance)
		if loadError != nil {
			log.Fatalf("Failed to init database: %s", loadError)
		}
		pp.Printf("%s Successfully initialized %s\n", time.Now(), "Database")
	})
}

func GetDatabase() *gorm.DB {
	if dbInstance == nil {
		MustInitDatabase()
	}
	return dbInstance
}

// initDatabase 获取数据库实例，使用单例模式
func initDatabase(db **gorm.DB) (err error) {
	cfg := config.GetConfig().App.Database
	dbType := normalizeDatabaseType(cfg.Type)
	if dbType == "" {
		dbType = databaseTypePostgres
	}
	dbLogger, err := newGormLogger(cfg)
	if err != nil {
		return err
	}

	var dialector gorm.Dialector
	switch dbType {
	case databaseTypePostgres:
		dialector = postgres.New(postgres.Config{
			DSN:                  postgresDSN(cfg),
			PreferSimpleProtocol: true,
		})
	default:
		return xerr.NewWithDetail(xerr.CodeInternalError, "unsupported database type: %s", cfg.Type)
	}

	dbInst, err := gorm.Open(dialector, &gorm.Config{
		DefaultContextTimeout: time.Second * 10,
		Logger:                dbLogger,
		TranslateError:        true,
	})
	if err != nil {
		return err
	}

	*db = dbInst
	sqlDB, err := dbInst.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return nil
}

func postgresDSN(cfg config.DatabaseConfig) string {
	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "require"
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=6 timezone=UTC",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		sslMode,
	)
}

func normalizeDatabaseType(databaseType string) string {
	switch strings.ToLower(strings.TrimSpace(databaseType)) {
	case "", "pgsql", "postgres", "postgresql":
		return databaseTypePostgres
	default:
		return databaseType
	}
}

func newGormLogger(cfg config.DatabaseConfig) (logger.Interface, error) {
	logWriter := io.Discard

	if cfg.LogPath == "" {
		if !cfg.OutputStdout {
			return nil, xerr.NewWithDetail(xerr.CodeInternalError, "log_path cannot be empty when output_stdout is false")
		}
	} else {
		rc := cfg.RollingLog
		logWriter = &logrotate.Rotator{
			Filename:   cfg.LogPath,
			MaxSize:    rc.MaxSize,
			MaxBackups: rc.MaxBackups,
		}
	}

	if cfg.OutputStdout {
		logWriter = io.MultiWriter(logWriter)
	}

	return logger.New(
		log.New(logWriter, "\r\n", log.LstdFlags), // 使用 lumberjack 作为 io.Writer
		logger.Config{
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      logger.Info, // 日志级别
			Colorful:      false,       // 禁用彩色打印
		},
	), nil
}

func CloseDatabase() error {
	if dbInstance != nil {
		db, err := dbInstance.DB()
		if err != nil {
			pp.Printf("Failed to close database: %v", err)
		}
		return db.Close()
	}
	return nil
}

// --------------------------

type ModelBase struct {
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (mb *ModelBase) BeforeUpdate(tx *gorm.DB) (err error) {
	tx.Statement.SetColumn("updated_at", time.Now())
	return
}

var OrderTypeMap = map[commpb.OrderType]string{
	1: "asc",
	0: "desc",
}

type PaginateArgs struct {
	NoCountQuery        bool
	AppendCreatedAtDesc bool
	SelectOnPageQuery   string
}

func Paginate(query *gorm.DB, page *commpb.PageArgs, sort []*commpb.SortArgs, allowedOrderFields []string, dest interface{}, args ...PaginateArgs) (total int64, err error) {
	if page == nil {
		page = &commpb.PageArgs{}
	}
	arg := PaginateArgs{}
	if len(args) > 0 {
		arg = args[0]
	}

	if !arg.NoCountQuery {
		if err = query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
			return 0, xerr.WrapDBE(err, "paginate count")
		}
		if total == 0 {
			return
		}
	}

	_, ps, offset := util.NormalizePageArgs(page)
	fields := allowedOrderFields
	containsCt := false
	for _, s := range sort {
		if len(fields) > 0 && !lo.Contains(fields, s.OrderField) {
			return 0, xerr.NewWithDetail(xerr.CodeParamError, "invalid sort field: %s", s.OrderField)
		}
		if s.OrderField == "created_at" {
			containsCt = true
		}
		if orderType, ok := OrderTypeMap[s.OrderType]; ok {
			query = query.Order(s.OrderField + " " + orderType)
		} else {
			return 0, xerr.NewWithDetail(xerr.CodeParamError, "invalid order type: %v", s.OrderType)
		}
	}

	if arg.AppendCreatedAtDesc && !containsCt {
		query = query.Order("created_at desc")
	}
	if arg.SelectOnPageQuery != "" {
		query = query.Select(arg.SelectOnPageQuery)
	}

	paged := query.Session(&gorm.Session{}).Offset(offset).Limit(ps)
	if r := paged.Find(dest); r.Error != nil {
		return 0, xerr.WrapDB(r.Error, "paginate find")
	} else if arg.NoCountQuery && r.RowsAffected > 0 {
		total = r.RowsAffected
	}
	return total, nil
}
