package util

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
	"sync/atomic"
)

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
func InitLogger(level string) {
	logger, _ := initLogger(newLogLevel(level))
	ReplaceGlobals(logger)
}

func initLogger(level zapcore.Level, opts ...zap.Option) (*zap.Logger, error) {
	stdOut, _, err := zap.Open([]string{"stdout"}...)
	if err != nil {
		return nil, err
	}
	cfg := zap.NewProductionConfig()
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg.EncoderConfig),
		zapcore.AddSync(stdOut),
		zap.NewAtomicLevelAt(level),
	)
	return zap.New(core), nil
}

// Logger returns the global Logger. It's safe for concurrent use.
func Logger() *zap.Logger {
	return globalLogger.Load().(*zap.Logger)
}

// ReplaceGlobals replaces the global Logger. It's safe for concurrent use.
func ReplaceGlobals(logger *zap.Logger) {
	globalLogger.Store(logger)
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
	p.lg.Errorw("prometheus client error", v)
}
