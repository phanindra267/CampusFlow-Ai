package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger is the structured logger handed to middleware.
type Logger struct {
	*logrus.Logger
}

// New builds a logger at the requested level and format.
//
// The logrus standard logger is configured identically. Code that has no
// injected logger available - notably pkg/response, which logs the underlying
// cause of every failed request - reaches for logrus.WithFields directly, and
// that output has to land in the same JSON stream with the same level as
// everything else rather than in logrus's default text format.
func New(level, format string) *Logger {
	l := newLogger(level, format)
	logrus.SetOutput(l.Out)
	logrus.SetLevel(l.Level)
	logrus.SetFormatter(l.Formatter)
	return &Logger{l}
}

func newLogger(level, format string) *logrus.Logger {
	l := logrus.New()

	l.SetOutput(os.Stdout)

	if format == "json" {
		l.SetFormatter(&logrus.JSONFormatter{})
	} else {
		l.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	}

	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	l.SetLevel(lvl)

	return l
}
