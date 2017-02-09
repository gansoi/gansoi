package logger

import (
	"log"
	"os"
)

var (
	debugLogger = New(os.Stdout, Green, os.Getenv("DEBUG"))
	infoLogger  = New(os.Stderr, Yellow, "*")
)

// Info will log something for the end user to stderr.
func Info(pkg string, format string, args ...interface{}) {
	infoLogger.Log(pkg, format, args...)
}

// Debug will log pure debug information to stderr.
func Debug(pkg string, format string, args ...interface{}) {
	debugLogger.Log(pkg, format, args...)
}

// InfoLogger will return a log.Logger.
func InfoLogger(pkg string) *log.Logger {
	return infoLogger.Logger(pkg)
}

// DebugLogger will return a log.Logger.
func DebugLogger(pkg string) *log.Logger {
	return debugLogger.Logger(pkg)
}
