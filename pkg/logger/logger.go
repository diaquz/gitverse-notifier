package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gitverse-notifier/pkg/config"

	"gopkg.in/natefinch/lumberjack.v2"
)

const logTimeFormat = "2006-01-02 15:04:05"

type ctxKey struct{}

var defaultLogger = slog.New(
	slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: slog.LevelInfo,
		},
	),
)

func SetupLogger(conf *config.Config) {
	fileWriter := &lumberjack.Logger{
		Filename:  filepath.Join(conf.LogDirPath, conf.LogFileName),
		MaxSize:   conf.LogMaxSize,
		MaxAge:    conf.LogMaxAge,
		LocalTime: true,
		Compress:  true,
	}

	writer := io.MultiWriter(fileWriter, os.Stdout)
	handler := slog.NewTextHandler(writer, loggerHandlerOptions(conf))
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

func loggerHandlerOptions(conf *config.Config) *slog.HandlerOptions {
	return &slog.HandlerOptions{
		Level: parseLevel(conf.LogLevel),
		AddSource: true,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && a.Value.Kind() == slog.KindTime {
				return slog.String(slog.TimeKey, a.Value.Time().Format(logTimeFormat))
			}

			return a
		},
	}
}

func parseLevel(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func With(ctx context.Context, args ...any) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ctxKey{}, FromContext(ctx).With(args...))
}

// FromContext возвращает логгер для запроса, либо defaultLogger если контекст пустой.
func FromContext(ctx context.Context) *slog.Logger {
	if ctx != nil {
		if log, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok && log != nil {
			return log
		}
	}

	return defaultLogger
}

func Debug(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Debug(msg, args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Info(msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Warn(msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Error(msg, args...)
}

// Fatal логирует ошибку на уровне ERROR и завершает программу
// (у slog нет FATAL уровня).
func Fatal(ctx context.Context, v ...any) {
	FromContext(ctx).Error(fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf логирует ошибку на уровне ERROR и завершает программу
// (у slog нет FATAL уровня).
func Fatalf(ctx context.Context, format string, args ...any) {
	FromContext(ctx).Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}
