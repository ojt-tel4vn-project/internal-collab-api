package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger

// Config cấu hình cho logger
type Config struct {
	Level      string // debug, info, warn, error
	LogFile    string // Đường dẫn file log
	MaxSize    int    // Kích thước tối đa của file log (MB)
	MaxBackups int    // Số lượng file backup tối đa
	MaxAge     int    // Số ngày giữ log cũ
	Compress   bool   // Nén file log cũ
	Console    bool   // Log ra console
}

// InitLogger khởi tạo logger với cấu hình
func InitLogger(cfg Config) error {
	// Thiết lập level
	level := zapcore.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}

	// Cấu hình encoder (format log)
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var cores []zapcore.Core

	// File writer với log rotation
	if cfg.LogFile != "" {
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.LogFile,
			MaxSize:    cfg.MaxSize, // megabytes
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge, // days
			Compress:   cfg.Compress,
		})

		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			fileWriter,
			level,
		)
		cores = append(cores, fileCore)
	}

	// Console writer
	if cfg.Console {
		consoleEncoderConfig := encoderConfig
		consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Màu sắc cho console

		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(consoleEncoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// Tạo logger
	core := zapcore.NewTee(cores...)
	Log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

// customTimeEncoder format thời gian theo kiểu: 2024-01-30 15:04:05.000
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// InitDefaultLogger khởi tạo logger với cấu hình mặc định
func InitDefaultLogger() error {
	return InitLogger(Config{
		Level:      "info",
		LogFile:    "logs/app.log",
		MaxSize:    100, // 100MB
		MaxBackups: 5,
		MaxAge:     30, // 30 days
		Compress:   true,
		Console:    true,
	})
}

// Các hàm wrapper để sử dụng thuận tiện hơn

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() error {
	return Log.Sync()
}

// WithFields tạo logger mới với các field được thêm sẵn
func WithFields(fields ...zap.Field) *zap.Logger {
	return Log.With(fields...)
}
