package logger

// This is a very simple logger inspired by a blog post by Dave Cheney:
// https://dave.cheney.net/2015/11/05/lets-talk-about-logging

import (
	"io/ioutil"
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

// Info will log something that could be of concern of end users.
func (l *Log) Info(pkg string, format string, args ...interface{}) {
	l.Printf(pkg, "\033[33m"+format+"\033[0m", args...)
}

// Debug will log pure debug information for developers.
func (l *Log) Debug(pkg string, format string, args ...interface{}) {
	_, print := l.positiveList[pkg]
	if print || l.printAll {
		l.Printf(pkg, "\033[32m"+format+"\033[0m", args...)
	}
}

// InfoLogger will return a log.Logger for info messages.
func (l *Log) InfoLogger(pkg string) *log.Logger {
	return log.New(l, "\033[35m"+pkg+"\033[0m: \033[33m", 0)
}

// DebugLogger will return a log.Logger for debug messages.
func (l *Log) DebugLogger(pkg string) *log.Logger {
	_, print := l.positiveList[pkg]
	if print || l.printAll {
		return log.New(l, "\033[35m"+pkg+"\033[0m: \033[32m", 0)
	}

	return log.New(ioutil.Discard, "", 0)
}

// Printf will print a log entry according to format.
func (l *Log) Printf(pkg string, format string, args ...interface{}) {
	l.logger.Printf("\033[35m"+pkg+"\033[0m: "+format+"\n", args...)
}

// Write implements io.Writer for log.New().
func (l *Log) Write(p []byte) (n int, err error) {
	str := strings.TrimSpace(string(p))

	l.logger.Printf("%s\033[0m\n", str)

	return len(p), nil
}
