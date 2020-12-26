package target

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
)

func testLoggerType(x interface{}) {}

func main() {
	logger := logrus.New()
	err := errors.New("test")

	// Ignore me
	err = fmt.Errorf("fmt wrap: %w", err)

	// Bad (type *logrus.Logger)
	logger.Error(err)
	logger.Error(42, err)
	logger.Error(fmt.Errorf("wrap: %w", err))
	logger.Errorf("logf wrap: %v", err)
	logger.Errorf("logf wrap: %d: %v", 42, err)
	testLoggerType(logger)

	// Bad (type *logrus.Entry)
	loggerEntry := logrus.NewEntry(logger)
	loggerEntry.Error(err)
	loggerEntry.Error(42, err)
	loggerEntry.Error(fmt.Errorf("wrap: %w", err))
	loggerEntry.Errorf("logf wrap: %v", err)
	loggerEntry.Errorf("logf wrap: %d: %v", 42, err)
	testLoggerType(loggerEntry)

	// Bad (interface logrus.FieldLogger)
	var loggerIface logrus.FieldLogger = logger
	loggerIface.Error(err)
	loggerIface.Error(42, err)
	loggerIface.Error(fmt.Errorf("wrap: %w", err))
	loggerIface.Errorf("logf wrap: %v", err)
	loggerIface.Errorf("logf wrap: %d: %v", 42, err)
	testLoggerType(loggerIface)

	// Good
	logger.WithError(err).Error()
	logger.WithError(err).Error(42)
	logger.WithError(err).Error("log")
	logger.WithError(err).Errorf("logf")
	logger.WithError(err).Errorf("logf: %d", 42)

	loggerEntry.WithError(err).Error()
	loggerEntry.WithError(err).Error(42)
	loggerEntry.WithError(err).Error("log")
	loggerEntry.WithError(err).Errorf("logf")
	loggerEntry.WithError(err).Errorf("logf: %d", 42)

	loggerIface.WithError(err).Error()
	loggerIface.WithError(err).Error(42)
	loggerIface.WithError(err).Error("log")
	loggerIface.WithError(err).Errorf("logf")
	loggerIface.WithError(err).Errorf("logf: %d", 42)
}
