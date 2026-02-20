package utils

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Logger is the global logger instance
	Logger *zap.Logger
	// SugaredLogger is the global sugared logger instance
	SugaredLogger *zap.SugaredLogger
)

// InitLogger initializes the global logger
func InitLogger(environment string) error {
	var config zap.Config
	
	if environment == "production" {
		// Production configuration: JSON format, INFO level
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	} else {
		// Development configuration: Console format, DEBUG level
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	
	// Customize time format
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	
	// Add caller information
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	
	// Build logger
	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return err
	}
	
	Logger = logger
	SugaredLogger = logger.Sugar()
	
	return nil
}

// LogInfo logs an info message
func LogInfo(message string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(message, fields...)
	}
}

// LogError logs an error message
func LogError(message string, err error, fields ...zap.Field) {
	if Logger != nil {
		allFields := append(fields, zap.Error(err))
		Logger.Error(message, allFields...)
	}
}

// LogWarn logs a warning message
func LogWarn(message string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(message, fields...)
	}
}

// LogDebug logs a debug message
func LogDebug(message string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(message, fields...)
	}
}

// LogFatal logs a fatal message and exits
func LogFatal(message string, err error, fields ...zap.Field) {
	if Logger != nil {
		allFields := append(fields, zap.Error(err))
		Logger.Fatal(message, allFields...)
	} else {
		// Fallback if logger not initialized
		panic(message + ": " + err.Error())
	}
}

// LogTransaction logs a transaction event
func LogTransaction(
	transactionID string,
	userID string,
	amount int64,
	status string,
	message string,
) {
	LogInfo(message,
		zap.String("transaction_id", transactionID),
		zap.String("user_id", userID),
		zap.Int64("amount_kobo", amount),
		zap.String("status", status),
		zap.String("event_type", "transaction"),
	)
}

// LogPayment logs a payment event
func LogPayment(
	paymentReference string,
	gateway string,
	amount int64,
	status string,
	message string,
) {
	LogInfo(message,
		zap.String("payment_reference", paymentReference),
		zap.String("gateway", gateway),
		zap.Int64("amount_kobo", amount),
		zap.String("status", status),
		zap.String("event_type", "payment"),
	)
}

// LogSpin logs a wheel spin event
func LogSpin(
	userID string,
	msisdn string,
	prizeID string,
	prizeValue int64,
	message string,
) {
	LogInfo(message,
		zap.String("user_id", userID),
		zap.String("msisdn", msisdn),
		zap.String("prize_id", prizeID),
		zap.Int64("prize_value_kobo", prizeValue),
		zap.String("event_type", "spin"),
	)
}

// LogAuth logs an authentication event
func LogAuth(
	msisdn string,
	action string,
	success bool,
	message string,
) {
	level := zapcore.InfoLevel
	if !success {
		level = zapcore.WarnLevel
	}
	
	if Logger != nil {
		Logger.Log(level, message,
			zap.String("msisdn", msisdn),
			zap.String("action", action),
			zap.Bool("success", success),
			zap.String("event_type", "auth"),
		)
	}
}

// LogAPIRequest logs an API request
func LogAPIRequest(
	method string,
	path string,
	statusCode int,
	duration time.Duration,
	userID string,
) {
	LogInfo("API request",
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration),
		zap.String("user_id", userID),
		zap.String("event_type", "api_request"),
	)
}

// LogWebhook logs a webhook event
func LogWebhook(
	provider string,
	event string,
	reference string,
	success bool,
	message string,
) {
	level := zapcore.InfoLevel
	if !success {
		level = zapcore.ErrorLevel
	}
	
	if Logger != nil {
		Logger.Log(level, message,
			zap.String("provider", provider),
			zap.String("event", event),
			zap.String("reference", reference),
			zap.Bool("success", success),
			zap.String("event_type", "webhook"),
		)
	}
}

// LogSecurity logs a security event
func LogSecurity(
	event string,
	severity string,
	userID string,
	ipAddress string,
	message string,
) {
	level := zapcore.WarnLevel
	if severity == "critical" {
		level = zapcore.ErrorLevel
	}
	
	if Logger != nil {
		Logger.Log(level, message,
			zap.String("event", event),
			zap.String("severity", severity),
			zap.String("user_id", userID),
			zap.String("ip_address", ipAddress),
			zap.String("event_type", "security"),
		)
	}
}

// Sync flushes any buffered log entries
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}
