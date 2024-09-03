package logger

import (
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

type Logger struct {
	*log.Logger
}

var (
	instance *Logger
	once     sync.Once
)

type Option func(*log.Options)

func WithLevel(level log.Level) Option {
	return func(o *log.Options) {
		o.Level = level
	}
}

func WithTimeFormat(format string) Option {
	return func(o *log.Options) {
		o.TimeFormat = format
	}
}

func WithPrefix(prefix string) Option {
	return func(o *log.Options) {
		o.Prefix = prefix
	}
}

func NewLogger(opts ...Option) *Logger {
	once.Do(func() {
		options := log.Options{
			ReportCaller:    true,
			ReportTimestamp: true,
			TimeFormat:      time.RFC3339,
			Prefix:          "yko|",
			Level:           log.InfoLevel,
		}

		for _, opt := range opts {
			opt(&options)
		}

		logger := log.NewWithOptions(os.Stderr, options)
		instance = &Logger{Logger: logger}
	})

	return instance
}

func GetLogger() *Logger {
	if instance == nil {
		return NewLogger()
	}
	return instance
}

func Info(msg string, keyvals ...interface{}) {
	GetLogger().Info(msg, keyvals...)
}

func Error(msg string, keyvals ...interface{}) {
	GetLogger().Error(msg, keyvals...)
}

func Debug(msg string, keyvals ...interface{}) {
	GetLogger().Debug(msg, keyvals...)
}

func Warn(msg string, keyvals ...interface{}) {
	GetLogger().Warn(msg, keyvals...)
}

func Fatal(msg string, keyvals ...interface{}) {
	GetLogger().Fatal(msg, keyvals...)
}
