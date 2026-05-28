package logger

import (
	"fmt"
	"io"
	"monorepo/config"
	"monorepo/pkg/logrotate"
	"monorepo/pkg/xerr"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 全局 logger 变量
// 这里将请求日志和应用日志进行分离

var ReqLogger *zap.Logger
var logger *zap.Logger

// Init 初始化全局 logger
func Init(requestLog, appLog *config.LogConfig) error {
	l, err := newLogger(requestLog)
	if err != nil {
		return xerr.NewWithError(xerr.CodeInternalError, err, "requestLog")
	}
	ReqLogger = l

	l, err = newLogger(appLog)
	if err != nil {
		return xerr.NewWithError(xerr.CodeInternalError, err, "appLog")
	}
	logger = l
	return nil
}

func newLogger(cfg *config.LogConfig) (*zap.Logger, error) {
	var logWriter = io.Discard // 允许不写入文件

	if cfg.OutputPath == "" {
		if !cfg.OutputToStdout {
			return nil, fmt.Errorf("output_path cannot be empty when output_to_stdout is false")
		}
	} else {
		rc := cfg.RollingLog
		logWriter = &logrotate.Rotator{
			Filename:   cfg.OutputPath,
			MaxSize:    rc.MaxSize,
			MaxBackups: rc.MaxBackups,
		}
	}

	if cfg.OutputToStdout {
		logWriter = io.MultiWriter(os.Stdout, logWriter)
	}

	// 创建 zap core
	core := zapcore.NewCore(
		getEncoder(cfg), // 使用你现有的编码器配置
		zapcore.AddSync(logWriter),
		zap.NewAtomicLevelAt(getZapLevel(cfg.Level)),
	)

	// 创建 logger
	zapLogger := zap.New(core)

	// 添加初始字段
	zapLogger = zapLogger.With(getInitialFields(cfg.InitialFields)...)

	return zapLogger, nil
}

// 辅助函数：获取编码器
func getEncoder(cfg *config.LogConfig) zapcore.Encoder {
	encoderConfig := getEncoderConfig(cfg)
	if cfg.Encoding == "json" {
		return zapcore.NewJSONEncoder(encoderConfig)
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// Debug 记录 debug 级别日志
func Debug(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Debug(msg, fields...)
	}
}

// Info 记录 info 级别日志
func Info(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Info(msg, fields...)
	}
}

// Warn 记录 warn 级别日志
func Warn(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Warn(msg, fields...)
	}
}

// Error 记录 error 级别日志
func Error(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Error(msg, fields...)
	}
}

// DPanic 记录 dpanic 级别日志
func DPanic(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.DPanic(msg, fields...)
	}
}

// Panic 记录 panic 级别日志
func Panic(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Panic(msg, fields...)
	}
}

// Fatal 记录 fatal 级别日志
func Fatal(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Fatal(msg, fields...)
	}
}

// Sync 同步日志缓冲区
func Sync() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}

// getZapLevel 将字符串转换为 zap 级别
func getZapLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel // 默认级别
	}
}

// getEncoderConfig 根据配置获取编码器配置
func getEncoderConfig(cfg *config.LogConfig) zapcore.EncoderConfig {
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     cfg.EncoderConfig.MessageKey,
		LevelKey:       cfg.EncoderConfig.LevelKey,
		TimeKey:        cfg.EncoderConfig.TimeKey,
		LineEnding:     "\n\n",
		EncodeLevel:    getLevelEncoder(cfg.EncoderConfig.LevelEncoder),
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.DateTime + ".000"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	return encoderConfig
}

// getLevelEncoder 根据配置获取级别编码器
func getLevelEncoder(levelEncoder string) zapcore.LevelEncoder {
	switch levelEncoder {
	case "uppercase":
		return zapcore.CapitalLevelEncoder // 没有 UppercaseLevelEncoder，使用 CapitalLevelEncoder
	case "capital":
		return zapcore.CapitalLevelEncoder
	case "capitalcolor":
		return zapcore.CapitalColorLevelEncoder
	default: // 默认使用 lowercase
		return zapcore.LowercaseLevelEncoder
	}
}

// getInitialFields 将初始字段映射转换为 zap.Field 列表
func getInitialFields(initialFields map[string]interface{}) []zap.Field {
	var fields []zap.Field
	for key, value := range initialFields {
		fields = append(fields, zap.Any(key, value))
	}
	return fields
}
