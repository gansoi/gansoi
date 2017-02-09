package logger

// This is a very simple logger inspired by a blog post by Dave Cheney:
// https://dave.cheney.net/2015/11/05/lets-talk-about-logging

import (
	"io"
	"io/ioutil"
	"log"
	"strings"
)

type (
	// Log is for printing logs. Well, duh :)
	Log struct {
		logger       *log.Logger
		positiveList map[string]bool
		printAll     bool
		color        string
	}
)

// New will return a new logger.
func New(w io.Writer, color string, packages string) *Log {
	l := &Log{
		color:        color,
		logger:       log.New(w, "", log.Ldate|log.Lmicroseconds),
		positiveList: make(map[string]bool),
		printAll:     false,
	}

	positives := strings.Split(packages, ",")
	for _, positive := range positives {
		if positive == "*" {
			l.printAll = true
		} else {
			l.positiveList[positive] = true
		}
	}

	return l
}

// Log will log a message according to format.
func (l *Log) Log(pkg string, format string, args ...interface{}) {
	_, print := l.positiveList[pkg]
	if print || l.printAll {
		l.Printf(pkg, l.color+format+Reset, args...)
	}
}

// Logger will return a log.Logger.
func (l *Log) Logger(pkg string) *log.Logger {
	_, print := l.positiveList[pkg]
	if print || l.printAll {
		return log.New(l, Purple+pkg+" "+l.color, 0)
	}

	return log.New(ioutil.Discard, "", 0)
}

// Printf will print a log entry according to format.
func (l *Log) Printf(pkg string, format string, args ...interface{}) {
	l.logger.Printf(Purple+pkg+Reset+" "+format+"\n", args...)
}

// Write implements io.Writer for log.New().
func (l *Log) Write(p []byte) (n int, err error) {
	str := strings.TrimSpace(string(p))

	l.logger.Printf("%s"+Reset+"\n", str)

	return len(p), nil
}
