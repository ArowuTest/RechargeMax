package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

// LogLevel represents logging levels
type LogLevel string

const (
	LogLevelDebug   LogLevel = "DEBUG"
	LogLevelInfo    LogLevel = "INFO"
	LogLevelWarning LogLevel = "WARNING"
	LogLevelError   LogLevel = "ERROR"
	LogLevelFatal   LogLevel = "FATAL"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Error     string                 `json:"error,omitempty"`
	ErrorCode ErrorCode              `json:"error_code,omitempty"`
	File      string                 `json:"file,omitempty"`
	Line      int                    `json:"line,omitempty"`
	Function  string                 `json:"function,omitempty"`
}

// Logger is the application logger
type Logger struct {
	level  LogLevel
	output *log.Logger
}

var defaultLogger *Logger

func init() {
	defaultLogger = &Logger{
		level:  LogLevelInfo,
		output: log.New(os.Stdout, "", 0),
	}
}

// SetLevel sets the logging level
func SetLevel(level LogLevel) {
	defaultLogger.level = level
}

// GetLevel returns the current logging level
func GetLevel() LogLevel {
	return defaultLogger.level
}

// log writes a log entry
func (l *Logger) log(level LogLevel, message string, ctx map[string]interface{}, err error) {
	// Check if this level should be logged
	if !l.shouldLog(level) {
		return
	}
	
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Context:   ctx,
	}
	
	// Add error information
	if err != nil {
		entry.Error = err.Error()
		
		// If it's an AppError, add the error code
		if appErr, ok := IsAppError(err); ok {
			entry.ErrorCode = appErr.Code
		}
	}
	
	// Add caller information for errors
	if level == LogLevelError || level == LogLevelFatal {
		if pc, file, line, ok := runtime.Caller(2); ok {
			entry.File = file
			entry.Line = line
			if fn := runtime.FuncForPC(pc); fn != nil {
				entry.Function = fn.Name()
			}
		}
	}
	
	// Marshal to JSON
	jsonBytes, _ := json.Marshal(entry)
	l.output.Println(string(jsonBytes))
}

// shouldLog checks if a message at the given level should be logged
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelDebug:   0,
		LogLevelInfo:    1,
		LogLevelWarning: 2,
		LogLevelError:   3,
		LogLevelFatal:   4,
	}
	
	return levels[level] >= levels[l.level]
}

// Debug logs a debug message
func Debug(message string, ctx ...map[string]interface{}) {
	var context map[string]interface{}
	if len(ctx) > 0 {
		context = ctx[0]
	}
	defaultLogger.log(LogLevelDebug, message, context, nil)
}

// Info logs an info message
func Info(message string, ctx ...map[string]interface{}) {
	var context map[string]interface{}
	if len(ctx) > 0 {
		context = ctx[0]
	}
	defaultLogger.log(LogLevelInfo, message, context, nil)
}

// Warning logs a warning message
func Warning(message string, ctx ...map[string]interface{}) {
	var context map[string]interface{}
	if len(ctx) > 0 {
		context = ctx[0]
	}
	defaultLogger.log(LogLevelWarning, message, context, nil)
}

// Error logs an error message
func Error(message string, err error, ctx ...map[string]interface{}) {
	var context map[string]interface{}
	if len(ctx) > 0 {
		context = ctx[0]
	}
	defaultLogger.log(LogLevelError, message, context, err)
}

// Fatal logs a fatal message and exits
func Fatal(message string, err error, ctx ...map[string]interface{}) {
	var context map[string]interface{}
	if len(ctx) > 0 {
		context = ctx[0]
	}
	defaultLogger.log(LogLevelFatal, message, context, err)
	os.Exit(1)
}

// LogRequest logs an HTTP request
func LogRequest(ctx context.Context, method, path string, statusCode int, duration time.Duration) {
	Info("HTTP request", map[string]interface{}{
		"method":      method,
		"path":        path,
		"status_code": statusCode,
		"duration_ms": duration.Milliseconds(),
	})
}

// LogTransaction logs a transaction
func LogTransaction(txType, txID, msisdn string, amount float64, status string) {
	Info("Transaction", map[string]interface{}{
		"type":   txType,
		"tx_id":  txID,
		"msisdn": msisdn,
		"amount": amount,
		"status": status,
	})
}

// LogPayment logs a payment event
func LogPayment(paymentRef, msisdn string, amount float64, status string) {
	Info("Payment", map[string]interface{}{
		"payment_ref": paymentRef,
		"msisdn":      msisdn,
		"amount":      amount,
		"status":      status,
	})
}

// LogCommission logs a commission event
func LogCommission(affiliateID, msisdn string, amount, commission float64) {
	Info("Commission", map[string]interface{}{
		"affiliate_id": affiliateID,
		"msisdn":       msisdn,
		"amount":       amount,
		"commission":   commission,
	})
}

// LogSpin logs a wheel spin event
func LogSpin(msisdn, prizeType string, prizeValue float64) {
	Info("Wheel spin", map[string]interface{}{
		"msisdn":      msisdn,
		"prize_type":  prizeType,
		"prize_value": prizeValue,
	})
}

// LogDraw logs a draw event
func LogDraw(drawID string, participants int, winners int) {
	Info("Draw executed", map[string]interface{}{
		"draw_id":      drawID,
		"participants": participants,
		"winners":      winners,
	})
}

// LogFraud logs a fraud detection event
func LogFraud(msisdn, reason string, severity string) {
	Warning("Fraud detected", map[string]interface{}{
		"msisdn":   msisdn,
		"reason":   reason,
		"severity": severity,
	})
}

// LogAudit logs an audit event
func LogAudit(userID, action, resource string, details map[string]interface{}) {
	Info("Audit", map[string]interface{}{
		"user_id":  userID,
		"action":   action,
		"resource": resource,
		"details":  details,
	})
}

// LogDatabaseError logs a database error with context
func LogDatabaseError(operation string, err error, ctx ...map[string]interface{}) {
	var context map[string]interface{}
	if len(ctx) > 0 {
		context = ctx[0]
	} else {
		context = make(map[string]interface{})
	}
	context["operation"] = operation
	
	Error(fmt.Sprintf("Database error: %s", operation), err, context)
}

// LogExternalAPIError logs an external API error
func LogExternalAPIError(service, endpoint string, err error, ctx ...map[string]interface{}) {
	var context map[string]interface{}
	if len(ctx) > 0 {
		context = ctx[0]
	} else {
		context = make(map[string]interface{})
	}
	context["service"] = service
	context["endpoint"] = endpoint
	
	Error(fmt.Sprintf("External API error: %s", service), err, context)
}

// LogValidationError logs a validation error
func LogValidationError(resource string, validationErrors interface{}) {
	Warning("Validation error", map[string]interface{}{
		"resource": resource,
		"errors":   validationErrors,
	})
}

// LogBusinessRuleViolation logs a business rule violation
func LogBusinessRuleViolation(rule, msisdn string, details map[string]interface{}) {
	Warning("Business rule violation", map[string]interface{}{
		"rule":    rule,
		"msisdn":  msisdn,
		"details": details,
	})
}
