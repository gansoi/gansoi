package logger

import (
	"log"
	"os"
)

var (
	defaultLogger = New(os.Getenv("DEBUG"))
)

// Red will log something severe using the default logger.
func Red(pkg string, format string, args ...interface{}) {
	defaultLogger.Red(pkg, format, args...)
}

// Yellow will log something that could be of concern using the default logger.
func Yellow(pkg string, format string, args ...interface{}) {
	defaultLogger.Yellow(pkg, format, args...)
}

// Green will log pure debug information using the default logger.
func Green(pkg string, format string, args ...interface{}) {
	defaultLogger.Green(pkg, format, args...)
}

// Logger will return a log.Logger from the default logger.
func Logger(pkg string) *log.Logger {
	return defaultLogger.Logger(pkg)
}
