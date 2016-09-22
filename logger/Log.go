package logger

import (
	"log"
	"os"
	"strings"
)

type (
	// Log is for printing debug logs.
	Log struct {
		logger       *log.Logger
		positiveList map[string]bool
		printAll     bool
	}
)

// New will return a new logger.
func New(debug string) *Log {
	l := &Log{
		logger:       log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds),
		positiveList: make(map[string]bool),
		printAll:     false,
	}

	positives := strings.Split(debug, ",")
	for _, positive := range positives {
		if positive == "*" {
			l.printAll = true
		} else {
			l.positiveList[positive] = true
		}
	}

	return l
}

// Red will log something severe.
func (l *Log) Red(pkg string, format string, args ...interface{}) {
	l.Printf(pkg, "\033[31m"+format+"\033[0m", args...)
}

// Yellow will log something that could be of concern.
func (l *Log) Yellow(pkg string, format string, args ...interface{}) {
	l.Printf(pkg, "\033[33m"+format+"\033[0m", args...)
}

// Green will log pure debug information.
func (l *Log) Green(pkg string, format string, args ...interface{}) {
	l.Printf(pkg, "\033[32m"+format+"\033[0m", args...)
}

// Logger will return a log.Logger.
func (l *Log) Logger(pkg string) *log.Logger {
	return log.New(l, "\033[35m"+pkg+"\033[0m: \033[36m", 0)
}

// Printf will print a log entry according to format.
func (l *Log) Printf(pkg string, format string, args ...interface{}) {
	_, print := l.positiveList[pkg]
	if print || l.printAll {
		l.logger.Printf("\033[35m"+pkg+"\033[0m: "+format+"\n", args...)
	}
}

// Write implements io.Writer.
func (l *Log) Write(p []byte) (n int, err error) {
	str := strings.TrimSpace(string(p))

	l.logger.Printf("%s\033[0m\n", str)

	return len(p), nil
}
