package db

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

var _ badger.Logger = (*BadgerSlogLogger)(nil)

func NewBadgerSlogLogger(logger *slog.Logger) BadgerSlogLogger {
	if logger == nil {
		logger = slog.Default()
	}

	return BadgerSlogLogger{Logger: logger}
}

type BadgerSlogLogger struct {
	Logger *slog.Logger
}

// Debugf implements badger.Logger.
func (b BadgerSlogLogger) Debugf(format string, args ...any) {
	b.doLog(b.Logger.Debug, format, args...)
}

// Errorf implements badger.Logger.
func (b BadgerSlogLogger) Errorf(format string, args ...any) {
	b.doLog(b.Logger.Error, format, args...)
}

// Infof implements badger.Logger.
func (b BadgerSlogLogger) Infof(format string, args ...any) {
	b.doLog(b.Logger.Info, format, args...)
}

// Warningf implements badger.Logger.
func (b BadgerSlogLogger) Warningf(format string, args ...any) {
	b.doLog(b.Logger.Warn, format, args...)
}

func (b BadgerSlogLogger) doLog(delegate func(string, ...any), format string, args ...any) {
	formatted := fmt.Sprintf(format, args...)

	for _, line := range strings.Split(formatted, "\n") {
		if line == "" {
			continue
		}
		delegate(line)
	}
}
