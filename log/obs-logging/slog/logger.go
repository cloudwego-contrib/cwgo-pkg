package slog

import (
	"github.com/cloudwego-contrib/obs-opentelemetry/log/logging"
	cwslog "github.com/cloudwego-contrib/obs-opentelemetry/logging/slog"
	"io"
	"log/slog"
)

const (
	LevelTrace  = slog.Level(-8)
	LevelNotice = slog.Level(2)
	LevelFatal  = slog.Level(12)
)

type Logger struct {
	cwslog.Logger
	config *config
}

func NewLogger(opts ...Option) *Logger {
	cfg := defaultConfig()

	for _, opt := range opts {
		opt.apply(cfg)
	}
	logger := &Logger{
		Logger: *cfg.logger,
		config: cfg,
	}
	logger.setTraceLogger()
	return logger
}

func (l *Logger) setTraceLogger() {
	log := slog.New(NewTraceHandler(l.GetOutput(), l.config.logger.GetHandler(), l.config.traceConfig))
	l.Logger.SetLogger(log)
}

func (l *Logger) SetOutput(writer io.Writer) {
	log := slog.New(NewTraceHandler(writer, l.config.logger.GetHandler(), l.config.traceConfig))
	l.config.logger.SetOutput(writer)
	l.Logger.SetLogger(log)
}
func (l *Logger) SetLevel(level logging.Level) {
	l.Logger.SetLevel(level)
}
