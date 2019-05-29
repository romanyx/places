package log

import (
	"fmt"
	"io"
	"sync"

	"github.com/sirupsen/logrus"
)

func init() {
	l = logger{
		Logger: logrus.StandardLogger(),
		mu:     &sync.RWMutex{},
	}
	l.SetLevel(logrus.InfoLevel)
}

var l logger

// logger contains a standard logger for all logging.
type logger struct {
	*logrus.Logger
	mu *sync.RWMutex
}

// SetOutput sets logger output.
func SetOutput(out io.Writer) {
	l.mu.Lock()
	{
		l.SetOutput(out)
	}
	l.mu.Unlock()
}

// SetLevel sets log level.
func SetLevel(level string) {
	l.mu.Lock()
	{
		switch level {
		case "debug":
			l.SetLevel(logrus.DebugLevel)
		case "info":
			l.SetLevel(logrus.InfoLevel)
		case "warn":
			l.SetLevel(logrus.WarnLevel)
		case "error":
			l.SetLevel(logrus.ErrorLevel)
		case "fatal":
			l.SetLevel(logrus.FatalLevel)
		case "panic":
			l.SetLevel(logrus.PanicLevel)
		default:
			l.SetLevel(logrus.DebugLevel)
		}
	}
	l.mu.Unlock()
}

// Warn prints at warn level.
func Warn(err error, fields map[string]interface{}) {
	l.mu.RLock()
	{
		l.WithFields(logrus.Fields(fields)).
			Warn(fmt.Sprintf("%v", err))
	}
	l.mu.RUnlock()
}

// Error prints at error level.
func Error(err error, fields map[string]interface{}) {
	l.mu.RLock()
	{
		l.WithFields(logrus.Fields(fields)).
			Error(fmt.Sprintf("%+v", err))
	}
	l.mu.RUnlock()
}

// Debug prints at info level.
func Debug(debug string, fields map[string]interface{}) {
	l.mu.RLock()
	{
		l.WithFields(logrus.Fields(fields)).
			Debug(debug)
	}
	l.mu.RUnlock()
}

// Info prints at info level.
func Info(info string, fields map[string]interface{}) {
	l.mu.RLock()
	{
		l.WithFields(logrus.Fields(fields)).
			Info(info)
	}
	l.mu.RUnlock()
}

// Fatal printf at fatal level.
func Fatal(err error, fields map[string]interface{}) {
	l.mu.RLock()
	{
		l.WithFields(logrus.Fields(fields)).
			Fatal(fmt.Sprintf("%+v", err))
	}
	l.mu.RUnlock()
}
