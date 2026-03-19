package badgerdb

import (
	"strings"

	"github.com/tkcrm/mx/logger"
)

type bLogger struct {
	l logger.Logger
}

// Errorf logs an error message with formatting
func (bl *bLogger) Errorf(format string, args ...any) {
	bl.l.Errorf(strings.TrimSpace(format), args...)
}

// Warningf logs a warning message with formatting
func (bl *bLogger) Warningf(format string, args ...any) {
	bl.l.Warnf(strings.TrimSpace(format), args...)
}

// Infof logs an info message with formatting
func (bl *bLogger) Infof(format string, args ...any) {
	bl.l.Infof(strings.TrimSpace(format), args...)
}

// Debugf logs a debug message with formatting
func (bl *bLogger) Debugf(format string, args ...any) {
	// bl.l.Debugf(strings.TrimSpace(format), args...)
}
