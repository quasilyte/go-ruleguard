package logging

// Logger implements a dummy logger to test vendored dependencies.
type Logger struct {
}

// Infof logs a message at error level.
func (l *Logger) Infof(fmt string, args ...interface{}) {
}

// Errorf logs a message at error level.
func (l *Logger) Errorf(fmt string, args ...interface{}) {
}

// GetLogger returns a Logger instance.
func GetLogger() *Logger {
	return &Logger{}
}
