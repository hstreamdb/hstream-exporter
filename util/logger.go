package util

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
	"sync/atomic"
)

type logWrapper struct {
	log   *zap.Logger
	level zap.AtomicLevel
}

var globalLogger atomic.Value

func newLogLevel(s string) zapcore.Level {
	s = strings.ToUpper(s)
	var level zapcore.Level
	switch s {
	case "PANIC":
		level = zapcore.PanicLevel
	case "FATAL":
		level = zapcore.FatalLevel
	case "ERROR":
		level = zapcore.ErrorLevel
	case "WARNING":
		level = zapcore.WarnLevel
	case "INFO":
		level = zapcore.InfoLevel
	case "DEBUG":
		level = zapcore.DebugLevel
	default:
		level = zapcore.InvalidLevel
	}
	return level
}

// InitLogger initializes a zap logger.
func InitLogger(level string) error {
	stdOut, _, err := zap.Open([]string{"stdout"}...)
	if err != nil {
		return err
	}

	logLevel := zap.NewAtomicLevelAt(newLogLevel(level))
	cfg := zap.NewProductionConfig()
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg.EncoderConfig),
		zapcore.AddSync(stdOut),
		logLevel,
	)

	logger := &logWrapper{
		log:   zap.New(core),
		level: logLevel,
	}

	ReplaceGlobals(logger)
	return nil
}

// Logger returns the global Logger. It's safe for concurrent use.
func Logger() *zap.Logger {
	return globalLogger.Load().(*logWrapper).log
}

// ReplaceGlobals replaces the global Logger. It's safe for concurrent use.
func ReplaceGlobals(logger *logWrapper) {
	globalLogger.Store(logger)
}

func UpdateLogLevel(s string) error {
	level := newLogLevel(s)
	if level == zapcore.InvalidLevel {
		return errors.New("unknown log level.")
	}

	logger := globalLogger.Load().(*logWrapper)
	if logger == nil {
		return errors.New("need init logger before update log level.")
	}
	defer logger.log.Sync()
	logger.level.SetLevel(level)
	return nil
}

// Sync flushes any buffered log entries.
func Sync() error {
	return Logger().Sync()
}

// PromErrLogger is a logger used by prometheus http server.
type PromErrLogger struct {
	lg *zap.SugaredLogger
}

func NewPromErrLogger() *PromErrLogger {
	return &PromErrLogger{
		Logger().Sugar(),
	}
}

// Println implement promhttp.Logger interface
func (p *PromErrLogger) Println(v ...interface{}) {
	p.lg.Errorf("prometheus client error: %+v", v)
}
