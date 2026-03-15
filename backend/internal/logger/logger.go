// Package logger provides a centralized zap-based structured logger for RechargeMax.
// All services and handlers should use this package instead of the stdlib log or fmt package.
//
// Usage:
//
//	logger.Info("recharge initiated", zap.String("msisdn", msisdn), zap.Int64("amount", amount))
//	logger.Error("payment failed", zap.Error(err), zap.String("reference", ref))
//	logger.Warn("retry attempt", zap.Int("attempt", n))
package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	once     sync.Once
	instance *zap.Logger
	sugar    *zap.SugaredLogger
)

// init builds the singleton logger on first use.
func init() {
	once.Do(func() {
		instance = build()
		sugar = instance.Sugar()
	})
}

func build() *zap.Logger {
	env := os.Getenv("APP_ENV")

	if env == "production" || env == "prod" {
		// JSON encoder for structured log ingestion in production
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		l, err := cfg.Build(zap.AddCallerSkip(1))
		if err != nil {
			panic("logger: failed to build production logger: " + err.Error())
		}
		return l
	}

	// Human-readable console encoder for local development
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	l, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic("logger: failed to build development logger: " + err.Error())
	}
	return l
}

// L returns the raw *zap.Logger for callers that need typed fields (zap.String, zap.Int64…).
func L() *zap.Logger { return instance }

// S returns the *zap.SugaredLogger for callers that prefer printf-style formatting.
func S() *zap.SugaredLogger { return sugar }

// Sync flushes buffered log entries.  Call defer logger.Sync() in main().
func Sync() { _ = instance.Sync() }

// ── Convenience wrappers (mirrors the most-used stdlib log calls) ─────────────

func Debug(msg string, fields ...zap.Field)  { instance.Debug(msg, fields...) }
func Info(msg string, fields ...zap.Field)   { instance.Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)   { instance.Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field)  { instance.Error(msg, fields...) }
func Fatal(msg string, fields ...zap.Field)  { instance.Fatal(msg, fields...) }

// Named returns a child logger with the given name (e.g. "spin_service").
func Named(name string) *zap.Logger { return instance.Named(name) }

// With returns a child logger that always includes the given fields.
func With(fields ...zap.Field) *zap.Logger { return instance.With(fields...) }
