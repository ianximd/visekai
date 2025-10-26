package logger

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

// Init initializes the logger
func Init(level string) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	l, err := config.Build()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
		os.Exit(1)
	}

	logger = l.Sugar()
}

// Info logs info level messages
func Info(msg string, keysAndValues ...interface{}) {
	if logger != nil {
		logger.Infow(msg, keysAndValues...)
	}
}

// Debug logs debug level messages
func Debug(msg string, keysAndValues ...interface{}) {
	if logger != nil {
		logger.Debugw(msg, keysAndValues...)
	}
}

// Warn logs warning level messages
func Warn(msg string, keysAndValues ...interface{}) {
	if logger != nil {
		logger.Warnw(msg, keysAndValues...)
	}
}

// Error logs error level messages
func Error(msg string, keysAndValues ...interface{}) {
	if logger != nil {
		logger.Errorw(msg, keysAndValues...)
	}
}

// Fatal logs fatal level messages and exits
func Fatal(msg string, keysAndValues ...interface{}) {
	if logger != nil {
		logger.Fatalw(msg, keysAndValues...)
	}
}
