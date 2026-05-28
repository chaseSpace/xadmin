package config

import "fmt"

var TestUIDs = []int32{
	60927,  // W.
	973324, // zq
	719819, // me
	531552, // rj
}

type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Biz        BizConfig        `mapstructure:"biz"`
	RollingLog RollingLogConfig `mapstructure:"rolling_log"` // yaml anchor definition
	RequestLog LogConfig        `mapstructure:"request_log"`
	AppLog     LogConfig        `mapstructure:"app_log"`
}

type RollingLogConfig struct {
	MaxSize    int `mapstructure:"max_size"`
	MaxBackups int `mapstructure:"max_backups"`
}

func (c Config) IsNonProd() bool {
	return !c.App.Server.IsProd()
}

func (a Config) Check() error {
	if err := a.Biz.Check(); err != nil {
		return err
	}
	if err := a.App.Check(); err != nil {
		return err
	}
	return nil
}

type AppConfig struct {
	Misc      Misc            `mapstructure:"misc"`
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Auth      AuthConfig      `mapstructure:"auth"`
	AuthAdmin AuthAdminConfig `mapstructure:"auth_admin"`
}

func (a *AppConfig) Check() error {
	if err := a.Misc.Check(); err != nil {
		return err
	}
	if err := a.Server.Check(); err != nil {
		return err
	}
	if err := a.Database.Check(); err != nil {
		return err
	}
	return nil
}

type Misc struct {
	PrintRouteOnStart bool   `mapstructure:"print_route_on_start"`
	Ip2regionXdbPath  string `mapstructure:"ip2region_xdb_path"`
}

func (m Misc) Check() error {
	if m.Ip2regionXdbPath == "" {
		return fmt.Errorf("Ip2regionXdbPath cannot be empty")
	}
	return nil
}

type ServerConfig struct {
	Port        int        `mapstructure:"port"`
	Host        string     `mapstructure:"host"`
	Env         string     `mapstructure:"env"`
	Timezone    string     `mapstructure:"timezone"`
	StoragePath string     `mapstructure:"storage_path"`
	DomainCert  DomainCert `mapstructure:"domain_cert"`
}
type DomainCert struct {
	Domain     []string `mapstructure:"domain"`
	AcmeEmail  string   `mapstructure:"acme_email"`
	AcmeToken  string   `mapstructure:"acme_token"`
	AcmeUser   string   `mapstructure:"acme_user"`
	AcmeServer string   `mapstructure:"acme_server"`
	CertDir    string   `mapstructure:"cert_dir"`
}

// 环境枚举常量
const (
	EnvDev  = "dev"
	EnvBeta = "beta"
	EnvProd = "prod"
)

// IsProd 判断是否为生产环境
func (s *ServerConfig) IsProd() bool {
	return s.Env == EnvProd
}

func (s *ServerConfig) IsDev() bool {
	return s.Env == EnvDev
}

func (s *ServerConfig) Check() error {
	switch s.Env {
	case EnvDev, EnvBeta, EnvProd:
	default:
		return fmt.Errorf("invalid env value: %s", s.Env)
	}
	if s.StoragePath == "" {
		return fmt.Errorf("storage_path cannot be empty")
	}
	if s.DomainCert.AcmeToken == "" {
		return fmt.Errorf("DomainCert.AcmeToken cannot be empty")
	}
	return nil
}

type DatabaseConfig struct {
	Type         string           `mapstructure:"type"`
	Host         string           `mapstructure:"host"`
	Port         int              `mapstructure:"port"`
	User         string           `mapstructure:"user"`
	Password     string           `mapstructure:"password"`
	Name         string           `mapstructure:"name"`
	SSLMode      string           `mapstructure:"ssl_mode"`
	LogPath      string           `mapstructure:"log_path"`
	OutputStdout bool             `mapstructure:"output_stdout"`
	RollingLog   RollingLogConfig `mapstructure:"rolling_log"`
}

func (d DatabaseConfig) Check() error {
	switch d.Type {
	case "", "pgsql", "postgres", "postgresql":
	default:
		return fmt.Errorf("invalid database.type value: %s", d.Type)
	}
	if d.Host == "" {
		return fmt.Errorf("database.host cannot be empty")
	}
	if d.Port <= 0 {
		return fmt.Errorf("database.port cannot be empty")
	}
	if d.User == "" {
		return fmt.Errorf("database.user cannot be empty")
	}
	if d.Name == "" {
		return fmt.Errorf("database.name cannot be empty")
	}
	return nil
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type AuthConfig struct {
	Secret          string   `mapstructure:"secret"`
	TokenTTLMinutes int64    `mapstructure:"token_ttl_minutes"`
	TokenTTLHours   int64    `mapstructure:"token_ttl_hours"`
	SkipPathSuffix  []string `mapstructure:"skip_path_suffix"`
}

type AuthAdminConfig struct {
	AuthorizedIdentity []AuthorizedIdentity `mapstructure:"authorized_identity"`
}

type AuthorizedIdentity struct {
	Name       string   `mapstructure:"name"`
	Secret     string   `mapstructure:"secret"`
	AllowedIPs []string `mapstructure:"allowed_ips"`
}

type BizConfig struct {
	AppName string      `mapstructure:"app_name"`
	About   AboutConfig `mapstructure:"about"`
}

func (b *BizConfig) Check() error {
	if b.AppName == "" {
		return fmt.Errorf("app_name cannot be empty")
	}
	return nil
}

type AboutConfig struct {
	Name        string `mapstructure:"name"`
	Logo        string `mapstructure:"logo"`
	Description string `mapstructure:"description"`
	ContactWx   string `mapstructure:"contact_wx"`
}

type LogConfig struct {
	Level          string                 `mapstructure:"level"`
	Encoding       string                 `mapstructure:"encoding"`
	OutputPath     string                 `mapstructure:"output_path"`
	OutputToStdout bool                   `mapstructure:"output_to_stdout"`
	RollingLog     RollingLogConfig       `mapstructure:"rolling_log"`
	InitialFields  map[string]interface{} `mapstructure:"initial_fields"`
	EncoderConfig  EncoderConfig          `mapstructure:"encoder_config"`
}

type EncoderConfig struct {
	MessageKey   string `mapstructure:"message_key"`
	TimeKey      string `mapstructure:"time_key"`
	LevelKey     string `mapstructure:"level_key"`
	LevelEncoder string `mapstructure:"level_encoder"`
}
