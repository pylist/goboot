package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *slog.Logger

type Config struct {
	Level      string // debug, info, warn, error
	Filename   string // 日志文件路径
	MaxSize    int    // 单个日志文件最大大小(MB)
	MaxBackups int    // 保留旧日志文件的最大数量
	MaxAge     int    // 保留旧日志文件的最大天数
	Compress   bool   // 是否压缩旧日志文件
	Console    bool   // 是否同时输出到控制台
}

func InitLogger(cfg *Config) error {
	if cfg == nil {
		cfg = &Config{
			Level:      "info",
			Filename:   "logs/app.log",
			MaxSize:    100,
			MaxBackups: 10,
			MaxAge:     30,
			Compress:   true,
			Console:    true,
		}
	}

	// 确保日志目录存在
	logDir := filepath.Dir(cfg.Filename)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// 解析日志级别
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// 文件写入器
	fileWriter := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}

	// 构建writer
	var writer io.Writer
	if cfg.Console {
		writer = io.MultiWriter(os.Stdout, fileWriter)
	} else {
		writer = fileWriter
	}

	// 创建handler
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}

	handler := slog.NewJSONHandler(writer, opts)
	Log = slog.New(handler)
	slog.SetDefault(Log)

	return nil
}

// log 内部日志方法，skip 用于指定跳过的调用栈层数
func log(ctx context.Context, level slog.Level, skip int, msg string, args ...any) {
	if !Log.Enabled(ctx, level) {
		return
	}
	var pcs [1]uintptr
	// skip: runtime.Callers -> log -> Debug/Info/Warn/Error -> 实际调用者
	runtime.Callers(skip, pcs[:])
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.Add(args...)
	_ = Log.Handler().Handle(ctx, r)
}

// 便捷方法
func Debug(msg string, args ...any) {
	log(context.Background(), slog.LevelDebug, 3, msg, args...)
}

func Info(msg string, args ...any) {
	log(context.Background(), slog.LevelInfo, 3, msg, args...)
}

func Warn(msg string, args ...any) {
	log(context.Background(), slog.LevelWarn, 3, msg, args...)
}

func Error(msg string, args ...any) {
	log(context.Background(), slog.LevelError, 3, msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelDebug, 3, msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelInfo, 3, msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelWarn, 3, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelError, 3, msg, args...)
}

// With 返回带有附加属性的新logger
func With(args ...any) *slog.Logger {
	return Log.With(args...)
}
