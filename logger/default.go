package logger

import (
	"log"
	"os"
)

var (
	defaultLogger = New(os.Getenv("DEBUG"))
)

// Info will log something for the end user using the default logger.
func Info(pkg string, format string, args ...interface{}) {
	defaultLogger.Info(pkg, format, args...)
}

// Debug will log pure debug information using the default logger.
func Debug(pkg string, format string, args ...interface{}) {
	defaultLogger.Debug(pkg, format, args...)
}

// InfoLogger will return a log.Logger from the default logger.
func InfoLogger(pkg string) *log.Logger {
	return defaultLogger.InfoLogger(pkg)
}

// DebugLogger will return a log.Logger from the default logger.
func DebugLogger(pkg string) *log.Logger {
	return defaultLogger.DebugLogger(pkg)
}
