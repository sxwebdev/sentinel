package loop

type logger interface {
	Debugf(format string, args ...any)
	Errorf(format string, args ...any)
}

type emptyLogger struct{}

func (l *emptyLogger) Debugf(format string, args ...any) {}
func (l *emptyLogger) Errorf(format string, args ...any) {}

var _ logger = (*emptyLogger)(nil)
