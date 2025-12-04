// Package logger 提供基于 zap 的日志系统
//
// 本模块封装了 uber/zap 日志库，提供：
// - 结构化日志记录
// - 多级别日志控制（Debug/Info/Warn/Error）
// - 开发模式和生产模式配置
// - 日志输出格式化
// - 全局日志实例管理
//
// 依赖：
// - go.uber.org/zap: 结构化日志库
//
// 使用示例：
//   logger.Init(logger.Config{Level: "info", Development: false})
//   logger.Info("服务启动", zap.String("host", "localhost"))
//   logger.Error("连接失败", zap.Error(err))
//
// 作者: Catsayer
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// globalLogger 是全局日志实例
	globalLogger *zap.Logger
	// globalSugar 是全局 sugar logger 实例（提供更简洁的 API）
	globalSugar *zap.SugaredLogger
)

// Config 日志配置
type Config struct {
	// Level 日志级别: debug, info, warn, error
	Level string
	// Development 是否为开发模式
	Development bool
	// DisableCaller 是否禁用调用者信息
	DisableCaller bool
	// DisableStacktrace 是否禁用堆栈跟踪
	DisableStacktrace bool
	// OutputPaths 日志输出路径
	OutputPaths []string
	// ErrorOutputPaths 错误日志输出路径
	ErrorOutputPaths []string
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		Level:             "info",
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
	}
}

// Init 初始化全局日志系统
func Init(cfg Config) error {
	// 解析日志级别
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	// 创建 encoder 配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 开发模式使用更易读的格式
	var encoder zapcore.Encoder
	if cfg.Development {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		// 生产模式使用 JSON 格式
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 创建 core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	// 创建 logger 选项
	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	}

	if !cfg.DisableStacktrace {
		opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	if cfg.Development {
		opts = append(opts, zap.Development())
	}

	// 创建 logger
	globalLogger = zap.New(core, opts...)
	globalSugar = globalLogger.Sugar()

	return nil
}

// L 返回全局 logger 实例
func L() *zap.Logger {
	if globalLogger == nil {
		// 如果未初始化，使用默认配置
		_ = Init(DefaultConfig())
	}
	return globalLogger
}

// S 返回全局 sugar logger 实例
func S() *zap.SugaredLogger {
	if globalSugar == nil {
		// 如果未初始化，使用默认配置
		_ = Init(DefaultConfig())
	}
	return globalSugar
}

// Debug 记录 debug 级别日志
func Debug(msg string, fields ...zap.Field) {
	L().Debug(msg, fields...)
}

// Info 记录 info 级别日志
func Info(msg string, fields ...zap.Field) {
	L().Info(msg, fields...)
}

// Warn 记录 warn 级别日志
func Warn(msg string, fields ...zap.Field) {
	L().Warn(msg, fields...)
}

// Error 记录 error 级别日志
func Error(msg string, fields ...zap.Field) {
	L().Error(msg, fields...)
}

// Fatal 记录 fatal 级别日志并退出程序
func Fatal(msg string, fields ...zap.Field) {
	L().Fatal(msg, fields...)
}

// Sync 刷新缓冲的日志
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// With 创建带有额外字段的子 logger
func With(fields ...zap.Field) *zap.Logger {
	return L().With(fields...)
}