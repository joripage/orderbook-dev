package logging

import (
	"context"
	"fmt"
	"runtime"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with context support
type Logger struct {
	logger *zap.Logger
}

// LogLevel defines the logging level
type LogLevel zapcore.Level

const (
	DEBUG LogLevel = LogLevel(zapcore.DebugLevel)
	INFO  LogLevel = LogLevel(zapcore.InfoLevel)
	WARN  LogLevel = LogLevel(zapcore.WarnLevel)
	ERROR LogLevel = LogLevel(zapcore.ErrorLevel)
	FATAL LogLevel = LogLevel(zapcore.FatalLevel)
)

// contextKey defines a type for context keys
type contextKey string

const (
	requestIDKey contextKey = "request_id"
	loggerKey    contextKey = "logger"
)

// NewLogger creates a new Logger instance
func NewLogger(level LogLevel) *Logger {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.Level(level))
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, _ := config.Build()
	return &Logger{logger: logger}
}

// WithRequestID adds request_id to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// getRequestID retrieves request_id from context
func getRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		return reqID
	}
	return "no-request-id"
}

// GetLogger retrieves or creates a logger for the given context
func GetLogger(ctx context.Context) (*Logger, context.Context) {
	// Check if logger exists in context
	if logger, ok := ctx.Value(loggerKey).(*Logger); ok {
		// Verify if request_id exists (basic key)
		if _, ok := ctx.Value(requestIDKey).(string); ok {
			return logger, ctx
		}
	}

	// Create new logger with request_id
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel) // Default level: INFO
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapLogger, _ := config.Build()
	logger := &Logger{
		logger: zapLogger.With(zap.String("request_id", getRequestID(ctx))),
	}

	// Store logger in context
	ctx = context.WithValue(ctx, loggerKey, logger)
	return logger, ctx
}

// logMessage logs a message with the specified level and context
func (l *Logger) logMessage(ctx context.Context, level LogLevel, msg string, fields ...zap.Field) {
	logger := l.logger
	switch level {
	case DEBUG:
		logger.Debug(msg, fields...)
	case INFO:
		logger.Info(msg, fields...)
	case WARN:
		logger.Warn(msg, fields...)
	case ERROR:
		logger.Error(msg, fields...)
	case FATAL:
		logger.Fatal(msg, fields...)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.logMessage(ctx, DEBUG, msg, fields...)
}

// Info logs an info message
func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.logMessage(ctx, INFO, msg, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.logMessage(ctx, WARN, msg, fields...)
}

// Error logs an error message
func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.logMessage(ctx, ERROR, msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	l.logMessage(ctx, FATAL, msg, fields...)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.logger.Sync()
}

func (c *Logger) basicFields() []zapcore.Field {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	// frame of getCurrentFrame function
	frame, _ := frames.Next() // nolint
	// frame of caller of this function
	frame, _ = frames.Next() // nolint
	logLine := fmt.Sprintf("%s:%d", frame.File, frame.Line)

	fields := []zapcore.Field{
		zap.String("request_id", uuid.New().String()),
		zap.String("log_line", logLine),
	}

	return fields
}
