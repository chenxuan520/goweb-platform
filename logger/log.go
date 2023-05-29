package logger

import (
	"fmt"
	"github.com/chenxuan520/goweb-platform/utils"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"

	"os"
)

type Log struct {
	Level         string `mapstructure:"level" json:"level" yaml:"level" ini:"level"`                                    // 级别
	Format        string `mapstructure:"format" json:"format" yaml:"format" ini:"level"`                                 // 输出
	Prefix        string `mapstructure:"prefix" json:"prefix" yaml:"prefix" ini:"level"`                                 // 日志前缀
	Director      string `mapstructure:"director" json:"director"  yaml:"director" ini:"level"`                          // 日志文件夹
	ShowLine      bool   `mapstructure:"show-line" json:"showLine" yaml:"showLine" ini:"showLine"`                       // 显示行
	EncodeLevel   string `mapstructure:"encode-level" json:"encodeLevel" yaml:"encode-level" ini:"encode-level"`         // 编码级
	StacktraceKey string `mapstructure:"stacktrace-key" json:"stacktraceKey" yaml:"stacktrace-key" ini:"stacktrace-key"` // 栈名
	LogInConsole  bool   `mapstructure:"log-in-console" json:"logInConsole" yaml:"log-in-console" ini:"log-in-console"`  // 输出控制台
}

//日志封装
var _defaultLogger *zap.Logger

//default zap config
var _defaultConfig *Log

func init() {
	_defaultConfig = &Log{
		Level:         "warn",
		Format:        "console",
		Prefix:        "[Log]",
		Director:      "logs",
		ShowLine:      false,
		EncodeLevel:   "LowercaseLevelEncoder",
		StacktraceKey: "stacktrace",
		LogInConsole:  true,
	}
}

func Init(config *Log) (logger *zap.Logger) {
	if config == nil {
		config = _defaultConfig
	}
	if ok := utils.Exists(fmt.Sprintf("%s", config.Director)); !ok { // 判断是否有Director文件夹
		fmt.Printf("create %v directory\n", config.Director)
		_ = os.Mkdir(fmt.Sprintf("%s", config.Director), os.ModePerm)
	}
	debugPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.DebugLevel
	})
	infoPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.InfoLevel
	})
	warnPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.WarnLevel
	})
	errorPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= zap.ErrorLevel
	})
	cores := make([]zapcore.Core, 0)
	switch config.Level {
	case "info":
		cores = append(cores, getEncoderCore(config.LogInConsole, config.Prefix, config.Format, config.EncodeLevel, config.StacktraceKey, fmt.Sprintf("%s/server_info.log", config.Director), infoPriority))
		cores = append(cores, getEncoderCore(config.LogInConsole, config.Prefix, config.Format, config.EncodeLevel, config.StacktraceKey, fmt.Sprintf("%s/server_warn.log", config.Director), warnPriority))
		cores = append(cores, getEncoderCore(config.LogInConsole, config.Prefix, config.Format, config.EncodeLevel, config.StacktraceKey, fmt.Sprintf("%s/server_error.log", config.Director), errorPriority))
	case "warn":
		cores = append(cores, getEncoderCore(config.LogInConsole, config.Prefix, config.Format, config.EncodeLevel, config.StacktraceKey, fmt.Sprintf("%s/server_warn.log", config.Director), warnPriority))
		cores = append(cores, getEncoderCore(config.LogInConsole, config.Prefix, config.Format, config.EncodeLevel, config.StacktraceKey, fmt.Sprintf("%s/server_error.log", config.Director), errorPriority))
	case "error":
		cores = append(cores, getEncoderCore(config.LogInConsole, config.Prefix, config.Format, config.EncodeLevel, config.StacktraceKey, fmt.Sprintf("%s/server_error.log", config.Director), errorPriority))
	default:
		cores = append(cores, getEncoderCore(config.LogInConsole, config.Prefix, config.Format, config.EncodeLevel, config.StacktraceKey, fmt.Sprintf("%s/server_debug.log", config.Director), debugPriority))
		cores = append(cores, getEncoderCore(config.LogInConsole, config.Prefix, config.Format, config.EncodeLevel, config.StacktraceKey, fmt.Sprintf("%s/server_info.log", config.Director), infoPriority))
		cores = append(cores, getEncoderCore(config.LogInConsole, config.Prefix, config.Format, config.EncodeLevel, config.StacktraceKey, fmt.Sprintf("%s/server_warn.log", config.Director), warnPriority))
		cores = append(cores, getEncoderCore(config.LogInConsole, config.Prefix, config.Format, config.EncodeLevel, config.StacktraceKey, fmt.Sprintf("%s/server_error.log", config.Director), errorPriority))
	}
	logger = zap.New(zapcore.NewTee(cores[:]...), zap.AddCaller())

	if config.ShowLine {
		logger = logger.WithOptions(zap.AddCaller())
	}
	_defaultLogger = logger
	return logger
}

func getEncoderConfig(prefix, encodeLevel, stacktraceKey string) (config zapcore.EncoderConfig) {
	config = zapcore.EncoderConfig{
		MessageKey:    "message",
		LevelKey:      "level",
		TimeKey:       "time",
		NameKey:       "logger",
		CallerKey:     "caller",
		StacktraceKey: stacktraceKey,
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(prefix + utils.TimeFormatDateV4))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}
	switch {
	case encodeLevel == "LowercaseLevelEncoder":
		config.EncodeLevel = zapcore.LowercaseLevelEncoder
	case encodeLevel == "LowercaseColorLevelEncoder":
		config.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	case encodeLevel == "CapitalLevelEncoder":
		config.EncodeLevel = zapcore.CapitalLevelEncoder
	case encodeLevel == "CapitalColorLevelEncoder":
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	default:
		config.EncodeLevel = zapcore.LowercaseLevelEncoder
	}
	return config
}

func getEncoder(prefix, format, encodeLevel, stacktraceKey string) zapcore.Encoder {
	if format == "json" {
		return zapcore.NewJSONEncoder(getEncoderConfig(prefix, encodeLevel, stacktraceKey))
	}
	return zapcore.NewConsoleEncoder(getEncoderConfig(prefix, encodeLevel, stacktraceKey))
}

func getEncoderCore(logInConsole bool, prefix, format, encodeLevel, stacktraceKey string, fileName string, level zapcore.LevelEnabler) (core zapcore.Core) {
	writer := getWriteSyncer(logInConsole, fileName) // 使用file-rotatelogs进行日志分割
	return zapcore.NewCore(getEncoder(prefix, format, encodeLevel, stacktraceKey), writer, level)
}

func getWriteSyncer(logInConsole bool, file string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   file,
		MaxSize:    10,
		MaxBackups: 200,
		MaxAge:     30,
		Compress:   true,
	}

	if logInConsole {
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(lumberJackLogger))
	}
	return zapcore.AddSync(lumberJackLogger)
}

func Sync() error {
	return _defaultLogger.Sync()
}

//Shutdown 全局关闭
func Shutdown() {
	_defaultLogger.Sync()
}
func GetLogger() *zap.Logger {
	return _defaultLogger
}
