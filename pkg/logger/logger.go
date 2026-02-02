// Package logger предоставляет глобальный логгер на основе zap
// с поддержкой передачи через context.
package logger

import (
	"context"

	"go.uber.org/zap"
)

// Log — глобальный экземпляр логгера.
// По умолчанию инициализирован как no-op до вызова Initialize.
var Log *zap.Logger = zap.NewNop()

// Initialize создаёт и настраивает глобальный логгер.
// level — уровень логирования (debug, info, warn, error).
// fields — дополнительные поля, которые будут добавлены ко всем записям.
func Initialize(level string, fields ...zap.Field) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()

	cfg.Level = lvl
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.MessageKey = "message"

	configuredLogger, err := cfg.Build()
	configuredLogger = configuredLogger.With(fields...)

	if err != nil {
		return err
	}

	Log = configuredLogger
	return nil
}

// loggerKey — ключ для хранения логгера в context.
type loggerKey struct{}

// ContextWithLogger возвращает новый context с привязанным логгером.
func ContextWithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// GetFromContext извлекает логгер из context.
// Если логгер не найден, возвращает глобальный Log.
func GetFromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*zap.Logger); ok && logger != nil {
		return logger
	}
	return Log
}
