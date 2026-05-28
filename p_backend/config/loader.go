package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/k0kubun/pp/v3"
	"github.com/spf13/viper"
)

// 全局配置变量
var (
	globalConfig *Config
	once         sync.Once
)

// MustLoadConfig 加载配置文件
func MustLoadConfig() *Config {
	var loadError error

	once.Do(func() {
		loadError = loadConfigInternal(&globalConfig)
	})

	if loadError != nil {
		log.Fatalf("Failed to load config: %v", loadError)
	}

	return globalConfig
}

// GetConfig 获取已加载的配置实例
func GetConfig() *Config {
	// 如果配置尚未加载，则先加载
	if globalConfig == nil {
		return MustLoadConfig()
	}
	return globalConfig
}

// loadConfigInternal 内部配置加载函数
func loadConfigInternal(target **Config) error {
	// 获取环境变量，确定是开发环境还是生产环境
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev" // 默认为开发环境
	}

	// 设置配置文件路径到 internal/config 目录
	configPath := filepath.Join(".", "config", env)

	// 创建 Viper 实例
	viperInstance := viper.New()

	// 查找并读取目录下的所有 YAML 文件
	files, err := filepath.Glob(filepath.Join(configPath, "*.yaml"))
	if err != nil {
		return fmt.Errorf("error finding config files: %v", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no config files found in %s", configPath)
	}

	// 为每个找到的配置文件设置配置
	for _, file := range files {
		viperInstance.SetConfigFile(file)
		if err = viperInstance.MergeInConfig(); err != nil {
			return fmt.Errorf("error merging config file %s: %v", file, err)
		}
	}

	var config Config
	if err = viperInstance.Unmarshal(&config); err != nil {
		return fmt.Errorf("error unmarshaling config: %v", err)
	}

	if err = config.Check(); err != nil {
		return err
	}

	if !config.App.Server.IsProd() {
		pp.Printf("--------- %s --------- \n%+v", "Load configuration", config)
		fmt.Println() // 空行占位
	} else {
		pp.Printf("This is PROD environment, All configuration is self check passed!!!\n")
	}

	*target = &config
	return nil
}
